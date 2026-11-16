package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/retr0h/freebie/services/api/internal/sprites"
)

// laLocation is the zone "today" is evaluated in for season checks.
var laLocation = func() *time.Location {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.FixedZone("PST", -8*60*60)
	}
	return loc
}()

// TeamResponse is one entry of GET /api/v1/teams.
type TeamResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	FullName    string  `json:"fullName"`
	League      string  `json:"league"`
	LeagueName  string  `json:"leagueName"`
	Sport       string  `json:"sport"`
	InSeason    bool    `json:"inSeason"`
	ReturnLabel *string `json:"returnLabel"`
	ActiveDeals int     `json:"activeDeals"`
	// BadgeSpriteID and Badge are the team's sport badge (sprites table).
	BadgeSpriteID *string         `json:"badgeSpriteId"`
	Badge         *sprites.Sprite `json:"badge"`
	// Colors is omitted when the team has no colors set.
	Colors *TeamColors `json:"colors,omitempty"`
}

// TeamColors are the two-tone "#RRGGBB" tile colors.
type TeamColors struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

// InSeason reports whether the MM-DD of day falls within [start, end]
// (inclusive). Ranges where start > end wrap the new year.
func InSeason(day time.Time, start, end string) bool {
	md := day.Format("01-02")
	if start <= end {
		return md >= start && md <= end
	}
	return md >= start || md <= end
}

var monthAbbrev = []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

// ReturnLabel is the label shown for an off-season team: the override if set,
// otherwise the 3-letter uppercase month of seasonStart. It returns nil when
// the team is in season.
func ReturnLabel(inSeason bool, seasonStart string, override *string) *string {
	if inSeason {
		return nil
	}
	if override != nil && *override != "" {
		return override
	}
	var m int
	if _, err := fmt.Sscanf(seasonStart, "%d-", &m); err != nil || m < 1 || m > 12 {
		return nil
	}
	l := monthAbbrev[m-1]
	return &l
}

// parseAt resolves the ?at=YYYY-MM-DD value, defaulting to now in Los Angeles.
func parseAt(at string, now time.Time) (time.Time, error) {
	if strings.TrimSpace(at) == "" {
		return now.In(laLocation), nil
	}
	return time.ParseInLocation("2006-01-02", at, laLocation)
}

// ListTeams returns all LA teams, in-season first, then by display order.
func (h *Handler) ListTeams(w http.ResponseWriter, r *http.Request) {
	day, err := parseAt(r.URL.Query().Get("at"), time.Now())
	if err != nil {
		respondError(w, http.StatusBadRequest, "at must be YYYY-MM-DD")
		return
	}

	teams, err := h.queries.ListTeams(r.Context())
	if err != nil {
		h.logger.Error("failed to list teams", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list teams")
		return
	}

	badges := h.loadSprites(r.Context())
	response := make([]TeamResponse, len(teams))
	for i, t := range teams {
		in := InSeason(day, t.SeasonStart, t.SeasonEnd)
		var override *string
		if t.ReturnLabel.Valid {
			override = &t.ReturnLabel.String
		}
		response[i] = TeamResponse{
			ID: t.ID, Name: t.Name, FullName: t.FullName, League: t.League, Sport: t.Sport,
			InSeason:    in,
			ReturnLabel: ReturnLabel(in, t.SeasonStart, override),
			ActiveDeals: int(t.ActiveDeals),
			LeagueName:  t.LeagueName,
		}
		if t.PrimaryColor.Valid && t.SecondaryColor.Valid {
			response[i].Colors = &TeamColors{Primary: t.PrimaryColor.String, Secondary: t.SecondaryColor.String}
		}
		if t.BadgeSpriteID.Valid {
			id := t.BadgeSpriteID.String
			response[i].BadgeSpriteID = &id
			if b, ok := badges[id]; ok {
				response[i].Badge = &b
			}
		}
	}
	// ListTeams is already ordered by display_order; stable sort keeps it within each group.
	sort.SliceStable(response, func(a, b int) bool {
		return response[a].InSeason && !response[b].InSeason
	})
	respondJSON(w, http.StatusOK, response)
}
