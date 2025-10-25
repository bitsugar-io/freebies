package sources

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GameStats holds all relevant stats from a game
// Metrics is a flexible map so each league can provide different stats
type GameStats struct {
	GameID    string
	GameDate  time.Time
	TeamID    string
	Opponent  string
	HomeGame  bool
	Won       bool
	Metrics   map[string]int // "strikeouts" -> 8, "runs" -> 5, "points" -> 112
}

// Source fetches game data for a specific league
type Source interface {
	// League returns the league identifier (e.g., "mlb", "nba")
	League() string

	// GetYesterdaysGame fetches stats from yesterday's game for the given team
	// Returns nil, nil if no game was played yesterday
	GetYesterdaysGame(ctx context.Context, teamID string) (*GameStats, error)

	// GetGameByDate fetches stats for a specific date
	// Returns nil, nil if no game was played on that date
	GetGameByDate(ctx context.Context, teamID string, date time.Time) (*GameStats, error)
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
	registry.sources[s.League()] = s
}

// Get retrieves a source by league
func Get(league string) (Source, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	s, ok := registry.sources[league]
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
