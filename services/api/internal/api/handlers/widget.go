package handlers

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/retr0h/freebie/services/api/internal/db"
)

// WidgetResponse is GET /api/v1/users/{userId}/widget: the lock-screen
// widget's lines, worded here so copy changes ship without an app release.
// The widget only lays them out; iOS draws lock-screen widgets in one tint.
type WidgetResponse struct {
	// State is "won" (a live win), "game" (a followed team plays today with a
	// deal on it) or "none".
	State string `json:"state"`
	// Title is the bold first line: "RAMS WON!", "RAMS WON! +2", "LAKERS 7PM".
	Title string `json:"title"`
	// Line1 and Line2 are the rectangular widget's second and third lines.
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
	// Inline is the one line above the clock.
	Inline string `json:"inline"`
	// Short fits the small circle: "RAMS", "7PM".
	Short string `json:"short"`
	// URL opens the app where the widget points: the deal, TODAY or TEAMS.
	URL string `json:"url"`
	// RefreshAt (RFC 3339) is when the lines next change: a code unlocks, a
	// deal ends, a game starts. The widget asks again then, or sooner.
	RefreshAt string `json:"refreshAt"`
}

// widgetWin is the part of a won deal the widget needs.
type widgetWin struct {
	ID          string
	TeamID      string
	OfferName   string
	PartnerName string
	TriggeredAt time.Time
	AvailableAt time.Time // zero when the code works right away
	ExpiresAt   time.Time
}

const (
	widgetMaxWait  = 30 * time.Minute // between refreshes when nothing is due
	widgetGameWait = 15 * time.Minute // during a game, a win can land any minute
)

// GetWidget returns the lock-screen widget's lines for one user (see
// WidgetResponse). Auth is the same as the user's other routes.
func (h *Handler) GetWidget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userId")
	now := time.Now()

	rows, err := h.queries.ListActiveTriggeredEventsForUser(ctx, userID)
	if err != nil {
		h.logger.Error("widget: failed to list won deals", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to load widget")
		return
	}
	var wins []widgetWin
	for _, row := range rows {
		if row.IsDismissed == 1 || !row.ExpiresAt.Valid {
			continue
		}
		win := widgetWin{
			ID: row.ID, TeamID: row.TeamID, OfferName: row.OfferName, PartnerName: row.PartnerName,
			TriggeredAt: row.TriggeredAt, ExpiresAt: row.ExpiresAt.Time,
		}
		if row.AvailableAt.Valid {
			win.AvailableAt = row.AvailableAt.Time
		}
		wins = append(wins, win)
	}

	short := map[string]string{}
	if teams, err := h.queries.ListTeams(ctx); err == nil {
		for _, t := range teams {
			short[t.ID] = t.Name
		}
	}

	var games []TodayGameResponse
	var following bool
	if len(wins) == 0 {
		games, following = h.widgetGames(ctx, userID, now)
	}
	respondJSON(w, http.StatusOK, widgetFor(now, wins, games, following, short))
}

// widgetGames lists today's games, with deals, for the teams the user follows.
func (h *Handler) widgetGames(ctx context.Context, userID string, now time.Time) ([]TodayGameResponse, bool) {
	subs, err := h.queries.ListUserSubscriptions(ctx, userID)
	if err != nil || len(subs) == 0 {
		return nil, false
	}
	followed := map[string]bool{}
	for _, s := range subs {
		followed[s.TeamID] = true
	}
	events, err := h.queries.ListActiveEvents(ctx)
	if err != nil {
		return nil, true
	}
	var mine []db.Event
	for _, e := range events {
		if followed[e.TeamID] {
			mine = append(mine, e)
		}
	}
	day := now.In(laLocation)
	return h.todayGames(ctx, day, day.Format("2006-01-02"), mine), true
}

// widgetFor words the widget. wins are the user's live, unclaimed wins; games
// are today's games for followed teams (only read when there is no win).
func widgetFor(now time.Time, wins []widgetWin, games []TodayGameResponse, following bool, short map[string]string) WidgetResponse {
	refresh := now.Add(widgetMaxWait)
	due := func(t time.Time) {
		if t.After(now) && t.Before(refresh) {
			refresh = t
		}
	}
	name := func(teamID string) string { return cmp.Or(short[teamID], teamID) }

	var live []widgetWin
	for _, win := range wins {
		if win.ExpiresAt.After(now) {
			live = append(live, win)
		}
	}
	if len(live) > 0 {
		// Newest win first; that's the one the lock screen names.
		slices.SortStableFunc(live, func(a, b widgetWin) int { return b.TriggeredAt.Compare(a.TriggeredAt) })
		win := live[0]
		team := name(win.TeamID)
		title := strings.ToUpper(team) + " WON!"
		if n := len(live) - 1; n > 0 {
			title += fmt.Sprintf(" +%d", n)
		}
		for _, l := range live {
			due(l.AvailableAt)
			due(l.ExpiresAt)
		}
		return WidgetResponse{
			State:     "won",
			Title:     title,
			Line1:     win.OfferName,
			Line2:     win.PartnerName + " · " + winStatus(now, win),
			Inline:    team + " won · " + win.OfferName,
			Short:     strings.ToUpper(team),
			URL:       "freebie://deal/" + win.ID,
			RefreshAt: refresh.UTC().Format(time.RFC3339),
		}
	}

	for _, g := range games {
		if g.State == "post" || len(g.Deals) == 0 {
			continue
		}
		start, err := time.Parse(time.RFC3339, g.StartsAt)
		if err != nil {
			continue
		}
		team := cmp.Or(g.TeamName, name(g.TeamID))
		when := strings.ToUpper(clock(start))
		if g.State == "in" {
			when = "LIVE"
			if r := now.Add(widgetGameWait); r.Before(refresh) {
				refresh = r
			}
		}
		due(start)
		vs := "vs "
		if !g.Home {
			vs = "at "
		}
		deal := g.Deals[0].PartnerName
		if n := len(g.Deals) - 1; n > 0 {
			deal += fmt.Sprintf(" +%d", n)
		}
		return WidgetResponse{
			State:     "game",
			Title:     strings.ToUpper(team) + " " + when,
			Line1:     vs + g.Opponent,
			Line2:     "Deal on it: " + deal,
			Inline:    team + " " + strings.ToLower(when) + " · " + deal,
			Short:     when,
			URL:       "freebie://today",
			RefreshAt: refresh.UTC().Format(time.RFC3339),
		}
	}

	if !following {
		return WidgetResponse{
			State: "none", Title: "FREEBIES", Line1: "Follow your LA teams", Line2: "Wins show up here",
			Inline: "Freebies · follow your LA teams", Short: "LA", URL: "freebie://teams",
			RefreshAt: now.Add(2 * widgetMaxWait).UTC().Format(time.RFC3339),
		}
	}
	return WidgetResponse{
		State: "none", Title: "NO WINS YET", Line1: "Your next free food", Line2: "shows up here",
		Inline: "Freebies · no wins yet", Short: "0", URL: "freebie://today",
		RefreshAt: refresh.UTC().Format(time.RFC3339),
	}
}

// winStatus is the deadline part of a won deal's line: "code works Mon",
// "ends 11:59pm", "5 days left".
func winStatus(now time.Time, win widgetWin) string {
	if win.AvailableAt.After(now) {
		if sameDay(now, win.AvailableAt) {
			return "code works " + clock(win.AvailableAt)
		}
		return "code works " + win.AvailableAt.In(laLocation).Format("Mon")
	}
	left := win.ExpiresAt.Sub(now)
	if left < 24*time.Hour {
		return "ends " + clock(win.ExpiresAt)
	}
	days := int(left / (24 * time.Hour))
	if days == 1 {
		return "1 day left"
	}
	return fmt.Sprintf("%d days left", days)
}

// clock is a Pacific wall-clock time: "7pm", "10:30am".
func clock(t time.Time) string {
	t = t.In(laLocation)
	if t.Minute() == 0 {
		return strings.ToLower(t.Format("3pm"))
	}
	return strings.ToLower(t.Format("3:04pm"))
}

func sameDay(a, b time.Time) bool {
	a, b = a.In(laLocation), b.In(laLocation)
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}
