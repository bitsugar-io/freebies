package mlb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// The schedule is read straight from the Stats API rather than through
// mlb-sdk, whose Game type drops gameType, which is the only way to tell
// spring training and exhibitions from real games.

// scheduleResponse is the subset of GET /api/v1/schedule we read.
type scheduleResponse struct {
	// Dates is nil when the key is missing, which is not a schedule (an empty
	// day still has "dates": []).
	Dates *[]struct {
		Games []scheduleGame `json:"games"`
	} `json:"dates"`
}

// scheduleGame is one game on the schedule.
type scheduleGame struct {
	GamePk int `json:"gamePk"`
	// GameType: R regular season; F, D, L, W (and C, P) postseason; S spring
	// training, E exhibition, A All-Star, I intrasquad (see counts).
	GameType string    `json:"gameType"`
	GameDate time.Time `json:"gameDate"`
	Status   struct {
		// AbstractGameState is Preview, Live or Final. MLB marks a
		// postponed game Final, so DetailedState decides (skippedStatus).
		AbstractGameState string `json:"abstractGameState"`
		DetailedState     string `json:"detailedState"`
	} `json:"status"`
	Teams struct {
		Home scheduleSide `json:"home"`
		Away scheduleSide `json:"away"`
	} `json:"teams"`
}

type scheduleSide struct {
	Team struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	Score int `json:"score"`
}

// counts reports whether a game counts for deals and Today: regular season and
// postseason games do; spring training (S), exhibitions (E), the All-Star game
// (A) and intrasquad games (I) do not. A missing gameType counts.
func (g *scheduleGame) counts() bool {
	switch g.GameType {
	case "", "R", "F", "D", "L", "W", "C", "P":
		return true
	}
	return false
}

// postseason reports a playoff game: wild card (F), division series (D),
// league championship (L), World Series (W), and C/P. A missing gameType is
// treated as regular season.
func (g *scheduleGame) postseason() bool {
	switch g.GameType {
	case "F", "D", "L", "W", "C", "P":
		return true
	}
	return false
}

// schedule returns the MLB (sportId 1) games on date, for one team when
// teamID is not 0. Requests go through s.http, which retries.
func (s *Source) schedule(ctx context.Context, teamID int, date time.Time) ([]scheduleGame, error) {
	q := url.Values{"sportId": {"1"}, "date": {date.Format("2006-01-02")}}
	if teamID != 0 {
		q.Set("teamId", strconv.Itoa(teamID))
	}
	u := strings.TrimSuffix(s.baseURL, "/") + "/api/v1/schedule?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	var sr scheduleResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("decoding: %w", err)
	}
	if sr.Dates == nil {
		return nil, fmt.Errorf("response has no dates list: %.200s", body)
	}
	var out []scheduleGame
	for _, d := range *sr.Dates {
		out = append(out, d.Games...)
	}
	return out, nil
}
