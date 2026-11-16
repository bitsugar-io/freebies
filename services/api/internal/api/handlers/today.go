package handlers

import (
	"context"
	"net/http"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/ttlcache"
)

const (
	// scheduleCacheTTL bounds how stale a league's "today" games can be; game
	// state (pre/in/post) is the part that moves.
	scheduleCacheTTL = 5 * time.Minute
	// scheduleCacheMax holds a few days of every league.
	scheduleCacheMax = 64
	// scheduleTimeout caps one league's schedule fetch (sources retry for up
	// to their own 30s client timeout, too long to hold the Today view).
	scheduleTimeout = 10 * time.Second
)

// TodayResponse is GET /api/v1/today: what anyone in LA can claim now and the
// day's games that have a deal on the line.
type TodayResponse struct {
	// Date is today in America/Los_Angeles (YYYY-MM-DD).
	Date string `json:"date"`
	// Live is every triggered deal still claimable, all teams, expiresAt ascending.
	Live []LiveDealResponse `json:"live"`
	// Games are today's games for teams with an active deal, startsAt ascending.
	Games []TodayGameResponse `json:"games"`
}

// LiveDealResponse is ActiveDealResponse without the per-user fields.
type LiveDealResponse struct {
	ID          string        `json:"id"`
	EventID     string        `json:"eventId"`
	TriggeredAt string        `json:"triggeredAt"`
	ExpiresAt   *string       `json:"expiresAt,omitempty"`
	AvailableAt *string       `json:"availableAt,omitempty"`
	Event       EventResponse `json:"event"`
}

// TodayGameResponse is one of today's games for a team with active deals.
type TodayGameResponse struct {
	TeamID   string `json:"teamId"`
	TeamName string `json:"teamName"`
	League   string `json:"league"`
	Opponent string `json:"opponent"`
	// StartsAt is RFC 3339 in UTC.
	StartsAt string `json:"startsAt"`
	Home     bool   `json:"home"`
	// State is "pre", "in" or "post".
	State string `json:"state"`
	// Deals are the team's active events.
	Deals []EventResponse `json:"deals"`
}

// SchedulerLookup returns the Scheduler for a league, or false when the
// league has none.
type SchedulerLookup func(league string) (sources.Scheduler, bool)

// registeredScheduler looks a league up in the global source registry.
func registeredScheduler(league string) (sources.Scheduler, bool) {
	src, err := sources.Get(league)
	if err != nil {
		return nil, false
	}
	s, ok := src.(sources.Scheduler)
	return s, ok
}

// todayState is a Handler's schedule cache and lookup. Like spriteCaches it
// lives beside Handler so this file stays self-contained.
type todayState struct {
	mu     sync.Mutex
	lookup SchedulerLookup
	cache  *ttlcache.Cache[[]sources.ScheduledGame]
}

var todayStates sync.Map // *Handler -> *todayState

func (h *Handler) today() *todayState {
	if s, ok := todayStates.Load(h); ok {
		return s.(*todayState)
	}
	s, _ := todayStates.LoadOrStore(h, &todayState{
		lookup: registeredScheduler,
		cache:  ttlcache.New[[]sources.ScheduledGame](scheduleCacheTTL, scheduleCacheMax),
	})
	return s.(*todayState)
}

// SetSchedulerLookup replaces how GetToday finds a league's Scheduler (the
// global source registry by default). Tests use it to stay isolated.
func (h *Handler) SetSchedulerLookup(fn SchedulerLookup) {
	s := h.today()
	s.mu.Lock()
	s.lookup = fn
	s.mu.Unlock()
}

func (s *todayState) scheduler(league string) (sources.Scheduler, bool) {
	s.mu.Lock()
	fn := s.lookup
	s.mu.Unlock()
	return fn(league)
}

// GetToday returns today's live deals and the day's games for teams with an
// active deal. ?at=YYYY-MM-DD picks the day (default today in Los Angeles).
// A league whose schedule fails is logged and left out.
func (h *Handler) GetToday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	day, err := parseAt(r.URL.Query().Get("at"), time.Now())
	if err != nil {
		respondError(w, http.StatusBadRequest, "at must be YYYY-MM-DD")
		return
	}
	date := day.Format("2006-01-02")

	live, err := h.liveDeals(ctx)
	if err != nil {
		h.logger.Error("failed to list live deals", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list live deals")
		return
	}

	events, err := h.queries.ListActiveEvents(ctx)
	if err != nil {
		h.logger.Error("failed to list events", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list events")
		return
	}

	games := h.todayGames(ctx, day, date, events)

	ptrs := make([]*EventResponse, 0, len(live))
	for i := range live {
		ptrs = append(ptrs, &live[i].Event)
	}
	for i := range games {
		for j := range games[i].Deals {
			ptrs = append(ptrs, &games[i].Deals[j])
		}
	}
	h.attachSprites(ctx, ptrs...)

	respondJSON(w, http.StatusOK, TodayResponse{Date: date, Live: live, Games: games})
}

// liveDeals lists every unexpired triggered deal, expiresAt ascending.
func (h *Handler) liveDeals(ctx context.Context) ([]LiveDealResponse, error) {
	rows, err := h.queries.ListActiveTriggeredEvents(ctx)
	if err != nil {
		return nil, err
	}
	const layout = "2006-01-02T15:04:05Z07:00"
	out := make([]LiveDealResponse, len(rows))
	for i, row := range rows {
		var expiresAt *string
		if row.ExpiresAt.Valid {
			t := row.ExpiresAt.Time.Format(layout)
			expiresAt = &t
		}
		out[i] = LiveDealResponse{
			ID:          row.ID,
			EventID:     row.EventID,
			TriggeredAt: row.TriggeredAt.Format(layout),
			ExpiresAt:   expiresAt,
			AvailableAt: nullTimePtr(row.AvailableAt),
			Event: eventToResponse(db.Event{
				ID:               row.ID_2,
				OfferID:          row.OfferID,
				TeamID:           row.TeamID,
				TeamName:         row.TeamName,
				League:           row.League,
				TeamColor:        row.TeamColor,
				Icon:             row.Icon,
				PartnerName:      row.PartnerName,
				OfferName:        row.OfferName,
				OfferDescription: row.OfferDescription,
				TriggerCondition: row.TriggerCondition,
				TriggerRule:      row.TriggerRule,
				RegionCode:       row.RegionCode,
				OfferUrl:         row.OfferUrl,
				AffiliateUrl:     row.AffiliateUrl,
				AffiliateTagline: row.AffiliateTagline,
				IsActive:         row.IsActive,
				CreatedAt:        row.CreatedAt,
				UpdatedAt:        row.UpdatedAt,
				SpriteID:         row.SpriteID,
				InVenue:          row.InVenue,
				Venue:            row.Venue,
			}),
		}
	}
	return out, nil
}

// teamKey identifies a team within its league (team codes repeat across
// leagues, e.g. NHL and MLB "BOS").
type teamKey struct{ league, team string }

// todayGames fetches each active league's schedule concurrently and keeps the
// games of teams with an active event, startsAt ascending.
func (h *Handler) todayGames(ctx context.Context, day time.Time, date string, events []db.Event) []TodayGameResponse {
	deals := map[teamKey][]db.Event{}
	var leagues []string
	for _, e := range events {
		k := teamKey{strings.ToLower(e.League), e.TeamID}
		if !slices.Contains(leagues, k.league) {
			leagues = append(leagues, k.league)
		}
		deals[k] = append(deals[k], e)
	}

	state := h.today()
	results := make([][]sources.ScheduledGame, len(leagues))
	var wg sync.WaitGroup
	for i, league := range leagues {
		sched, ok := state.scheduler(league)
		if !ok {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			games, err := state.cache.Get(ctx, league+"|"+date, func(ctx context.Context) ([]sources.ScheduledGame, error) {
				// Detach from the request so one client hanging up does not
				// fail the shared fetch for everyone waiting on it.
				ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), scheduleTimeout)
				defer cancel()
				return sched.GamesOn(ctx, day)
			})
			if err != nil {
				h.logger.Warn("skipping league schedule", "league", league, "date", date, "error", err)
				return
			}
			results[i] = games
		}()
	}
	wg.Wait()

	out := []TodayGameResponse{}
	for i, league := range leagues {
		for _, g := range results[i] {
			evs := deals[teamKey{league, g.TeamID}]
			if len(evs) == 0 {
				continue
			}
			resp := TodayGameResponse{
				TeamID:   g.TeamID,
				TeamName: evs[0].TeamName,
				League:   evs[0].League,
				Opponent: g.Opponent,
				StartsAt: g.StartsAt.UTC().Format(time.RFC3339),
				Home:     g.Home,
				State:    g.State,
				Deals:    make([]EventResponse, len(evs)),
			}
			for j, e := range evs {
				resp.Deals[j] = eventToResponse(e)
			}
			out = append(out, resp)
		}
	}
	sort.SliceStable(out, func(a, b int) bool {
		if out[a].StartsAt != out[b].StartsAt {
			return out[a].StartsAt < out[b].StartsAt
		}
		return out[a].TeamID < out[b].TeamID
	})
	return out
}
