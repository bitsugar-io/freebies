package sources

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// GameStats holds all relevant stats from a game
// Metrics is a flexible map so each league can provide different stats
type GameStats struct {
	GameID   string
	GameDate time.Time
	TeamID   string
	Opponent string
	HomeGame bool
	Won      bool
	// Postseason is true for a playoff game, when the feed says so (ESPN
	// season type 3, NHL gameType 3). Unknown counts as regular season.
	Postseason bool
	Metrics    map[string]int // "strikeouts" -> 8, "runs" -> 5, "points" -> 112
}

// Source fetches game data for a specific league
type Source interface {
	// League returns the league identifier (e.g., "mlb", "nba")
	League() string

	// Metrics lists every metric key the source can put in GameStats.Metrics.
	// A deal whose trigger_rule metric is not listed can never fire; the
	// checker and `deals audit` use this to say so instead of failing quietly.
	Metrics() []string

	// GetYesterdaysGame fetches stats from yesterday's game for the given team
	// Returns nil, nil if no game was played yesterday
	GetYesterdaysGame(ctx context.Context, teamID string) (*GameStats, error)

	// GetGameByDate fetches stats for a specific date
	// Returns nil, nil if no game was played on that date
	GetGameByDate(ctx context.Context, teamID string, date time.Time) (*GameStats, error)
}

// MultiGameSource is implemented by sources that can return every finished
// game a team played on one date: an MLB doubleheader, or NCAA football and
// basketball on the same day. The checker evaluates a deal against each of
// them; GetGameByDate stays the first game for other callers.
type MultiGameSource interface {
	// GetGamesByDate returns the team's finished, counting games on date
	// (only its Y/M/D are used), earliest first; empty when there are none.
	GetGamesByDate(ctx context.Context, teamID string, date time.Time) ([]*GameStats, error)
}

// SeasonAware is implemented by sources whose feed says whether a game is a
// playoff game (GameStats.Postseason). A rule with season "regular" needs one;
// `deals audit` fails an active deal that has it on any other source.
type SeasonAware interface {
	KnowsPostseason() bool
}

// TeamLister is implemented by sources that know their team codes, so `deals
// audit` can catch an event whose team_id the source would reject at run time.
type TeamLister interface {
	// Teams returns the Freebies team codes (events.team_id) the source maps.
	Teams() []string
}

// ScheduledGame is one game on a team's schedule, in any state.
type ScheduledGame struct {
	// TeamID is the Freebies team code (events.team_id).
	TeamID   string
	Opponent string
	// StartsAt is the scheduled start.
	StartsAt time.Time
	// Home is true only for a real home game (not a neutral site).
	Home bool
	// Sport names the feed the game came from ("football", "basketball")
	// when a league has more than one (NCAA); empty otherwise. A deal whose
	// metric names a sport ("football_home_points") rides only on that
	// sport's games.
	Sport string
	// State is "pre" (not started), "in" (live) or "post" (over).
	State string
}

// Game states for ScheduledGame.State.
const (
	StatePre  = "pre"
	StateIn   = "in"
	StatePost = "post"
)

// Scheduler is implemented by sources that can list a day's games before they
// finish, for the app's "Today" view.
type Scheduler interface {
	// GamesOn returns the games of the source's mapped teams that start on
	// the given America/Los_Angeles date (only its Y/M/D are used), in any
	// state. Postponed and canceled games are left out. A game between two
	// mapped teams appears once per team.
	GamesOn(ctx context.Context, date time.Time) ([]ScheduledGame, error)
}

// Registry holds all registered sources
type Registry struct {
	mu      sync.RWMutex
	sources map[string]Source
}

// Global registry
var registry = &Registry{
	sources: make(map[string]Source),
}

// Register adds a source to the registry
func Register(s Source) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.sources[strings.ToLower(s.League())] = s
}

// Get retrieves a source by league
func Get(league string) (Source, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	s, ok := registry.sources[strings.ToLower(strings.TrimSpace(league))]
	if !ok {
		return nil, fmt.Errorf("no source registered for league: %s", league)
	}
	return s, nil
}

// List returns all registered league names
func List() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	leagues := make([]string, 0, len(registry.sources))
	for league := range registry.sources {
		leagues = append(leagues, league)
	}
	return leagues
}
