package espn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/retr0h/freebie/services/api/internal/sources"
)

// Game is a finished game seen from one team's side.
type Game struct {
	Feed  Feed
	Event *Event
	Comp  *Competition
	Us    *Competitor
	Them  *Competitor
	// Home is true only for a real home game: the team is listed as home and
	// the game is not at a neutral site (ESPN still labels one side "home"
	// at neutral sites, which would make home_* metrics lie).
	Home       bool
	OurScore   int
	TheirScore int
}

// Won reports whether the team scored more than its opponent.
func (g *Game) Won() bool { return g.OurScore > g.TheirScore }

// FeedConfig pairs a scoreboard with the function that turns a finished game into
// metrics. A league has one feed, except NCAA (football and basketball).
type FeedConfig struct {
	Feed
	// Build returns the metrics for a finished game. It may fetch more
	// (summary) through c.
	Build func(ctx context.Context, c *Client, g *Game) (map[string]int, error)
}

// Config describes one league's source.
type Config struct {
	League string
	// Teams maps Freebies team codes (events.team_id) to ESPN team ids. The
	// mapping is explicit so a team that is not listed fails loudly instead
	// of silently never matching.
	Teams map[string]string
	// Metrics lists every metric the source can emit; tests assert it
	// matches what Build produces.
	Metrics []string
	Feeds   []FeedConfig
}

var _ sources.MultiGameSource = (*Source)(nil)

// Source implements sources.Source for one league on top of a Client.
type Source struct {
	cfg    Config
	client *Client
}

// NewSource builds a source over client.
func NewSource(cfg Config, client *Client) *Source {
	return &Source{cfg: cfg, client: client}
}

// Client exposes the underlying client (tests tune its transport).
func (s *Source) Client() *Client { return s.client }

// KnowsPostseason: ESPN events carry season.type and slug (see seasonPhase).
func (s *Source) KnowsPostseason() bool { return true }

// League returns the league id.
func (s *Source) League() string { return s.cfg.League }

// Metrics returns the metrics this source can emit.
func (s *Source) Metrics() []string { return slices.Clone(s.cfg.Metrics) }

// Teams returns the team codes the source maps, sorted.
func (s *Source) Teams() []string {
	return slices.Sorted(maps.Keys(s.cfg.Teams))
}

// GetYesterdaysGame fetches yesterday's game (America/Los_Angeles, since
// deals are LA-anchored).
func (s *Source) GetYesterdaysGame(ctx context.Context, teamID string) (*sources.GameStats, error) {
	return s.GetGameByDate(ctx, teamID, time.Now().In(s.client.loc).AddDate(0, 0, -1))
}

// GetGameByDate returns the team's first finished game that started on the
// given Pacific date, or (nil, nil) when it has none (not scheduled, not
// final, postponed, canceled, preseason or All-Star). GetGamesByDate returns
// every game of the day.
func (s *Source) GetGameByDate(
	ctx context.Context,
	teamID string,
	date time.Time,
) (*sources.GameStats, error) {
	games, err := s.GetGamesByDate(ctx, teamID, date)
	var partial *sources.PartialError
	if err != nil && (!errors.As(err, &partial) || len(games) == 0) {
		return nil, err
	}
	if len(games) == 0 {
		return nil, nil
	}
	return games[0], nil
}

// GetGamesByDate returns every finished game the team started on the given
// Pacific date, earliest first, from every feed (NCAA reads football and
// basketball, and both can have a game the same day). A feed that fails does
// not hide another feed's game: the games that were read come back with a
// *sources.PartialError naming the failed feeds, so a deal that rides on the
// failed sport sees an error rather than "no game". When every feed failed
// the error is returned alone.
func (s *Source) GetGamesByDate(
	ctx context.Context,
	teamID string,
	date time.Time,
) ([]*sources.GameStats, error) {
	espnID, ok := s.cfg.Teams[teamID]
	if !ok {
		return nil, fmt.Errorf("unknown team ID: %s", teamID)
	}

	type found struct {
		start time.Time
		stats *sources.GameStats
	}
	var all []found
	var errs []error
	for _, fc := range s.cfg.Feeds {
		var feed []found
		err := func() error {
			events, err := s.client.Games(ctx, fc.Feed, date)
			if err != nil {
				return err
			}
			for i := range events {
				st, err := s.gameStats(ctx, fc, &events[i], espnID, teamID, date)
				if err != nil {
					return err
				}
				if st != nil {
					feed = append(feed, found{events[i].Date.Time, st})
				}
			}
			return nil
		}()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		all = append(all, feed...)
	}
	if len(errs) == len(s.cfg.Feeds) {
		return nil, errors.Join(errs...)
	}
	for _, err := range errs {
		slog.Warn("ESPN feed failed; using the other feeds' games",
			"league", s.cfg.League, "team", teamID, "date", date.Format("2006-01-02"), "error", err)
	}

	sort.SliceStable(all, func(a, b int) bool { return all[a].start.Before(all[b].start) })
	out := make([]*sources.GameStats, len(all))
	for i, f := range all {
		out[i] = f.stats
	}
	if len(errs) > 0 {
		return out, &sources.PartialError{Err: errors.Join(errs...)}
	}
	return out, nil
}

// gameStats builds the stats for e from espnID's side, or returns nil when the
// team did not play in it or it does not count (not final, preseason).
func (s *Source) gameStats(
	ctx context.Context,
	fc FeedConfig,
	e *Event,
	espnID, teamID string,
	date time.Time,
) (*sources.GameStats, error) {
	counts, post := seasonPhase(e)
	if !counts {
		return nil, nil
	}
	g := finalGame(e, espnID)
	if g == nil {
		return nil, nil
	}
	g.Feed = fc.Feed
	if g.OurScore < 0 {
		return nil, fmt.Errorf("ESPN game %s has an unreadable score", g.Event.ID)
	}
	metrics, err := fc.Build(ctx, s.client, g)
	if err != nil {
		return nil, fmt.Errorf("ESPN game %s: %w", g.Event.ID, err)
	}
	id := "espn-" + s.cfg.League + "-" + g.Event.ID
	if fc.Name != "" && len(s.cfg.Feeds) > 1 {
		id = "espn-" + s.cfg.League + "-" + fc.Name + "-" + g.Event.ID
	}
	return &sources.GameStats{
		GameID:     id,
		GameDate:   date,
		TeamID:     teamID,
		Opponent:   g.Them.Team.DisplayName,
		HomeGame:   g.Home,
		Won:        g.Won(),
		Postseason: post,
		Metrics:    metrics,
	}, nil
}

// seasonPhase reports whether e counts for deals at all and, if so, whether it
// is a postseason game. ESPN's season.type, as seen in the feeds:
//
//	1  preseason                  does not count
//	2  regular season             counts
//	3  postseason                 counts, postseason
//	4  off-season                 does not count
//	5  NBA play-in                counts, postseason (it is not regular season)
//	0  missing                    counts, regular season (trimmed fixtures)
//
// Soccer feeds use league-specific ids instead (MLS 13846 "regular-season",
// 13836 "western-conference-playoffs---round-one", NWSL 14023
// "playoffs---quarterfinals"), so any other type is read from season.slug: a
// slug naming playoffs, the post-season or a play-in is postseason, one naming
// the preseason does not count, anything else is regular season.
//
// An All-Star game ("ALLSTAR", filed as regular season) or an exhibition
// ("EXH") competition never counts, whatever its season type.
func seasonPhase(e *Event) (counts, postseason bool) {
	if len(e.Competitions) > 0 {
		switch e.Competitions[0].Type.Abbreviation {
		case "ALLSTAR", "EXH":
			return false, false
		}
	}
	switch e.Season.Type {
	case 0, 2:
		return true, false
	case 1, 4:
		return false, false
	case 3, 5:
		return true, true
	}
	slug := strings.ToLower(e.Season.Slug)
	switch {
	case strings.Contains(slug, "preseason") || strings.Contains(slug, "pre-season"):
		return false, false
	case strings.Contains(slug, "playoff") || strings.Contains(slug, "post-season") ||
		strings.Contains(slug, "postseason") || strings.Contains(slug, "play-in"):
		return true, true
	}
	return true, false
}

// finalGame returns the event from teamID's side if the team played in it and
// it is over and played; nil otherwise. A postponed, canceled, suspended or
// live game is never completed, so it is skipped.
func finalGame(e *Event, teamID string) *Game {
	if len(e.Competitions) == 0 {
		return nil
	}
	comp := &e.Competitions[0]
	var us, them *Competitor
	for i := range comp.Competitors {
		c := &comp.Competitors[i]
		if c.Team.ID == teamID {
			us = c
		} else {
			them = c
		}
	}
	if us == nil || them == nil || len(comp.Competitors) != 2 {
		return nil
	}
	if !comp.Status.Type.Completed || comp.Status.Type.State != "post" {
		return nil
	}
	g := &Game{
		Event: e, Comp: comp, Us: us, Them: them,
		Home: us.HomeAway == "home" && !comp.NeutralSite,
	}
	var err1, err2 error
	g.OurScore, err1 = strconv.Atoi(us.Score)
	g.TheirScore, err2 = strconv.Atoi(them.Score)
	if err1 != nil || err2 != nil {
		g.OurScore = -1
	}
	return g
}

// winMetrics adds the metrics every sport shares.
func (g *Game) winMetrics(m map[string]int) {
	m["win"] = boolToInt(g.Won())
	m["home_win"] = boolToInt(g.Won() && g.Home)
}

// PointsMetrics builds the basketball and football metrics:
// win, home_win, points, home_points.
func PointsMetrics(_ context.Context, _ *Client, g *Game) (map[string]int, error) {
	m := map[string]int{
		"points":      g.OurScore,
		"home_points": conditionalInt(g.Home, g.OurScore),
	}
	g.winMetrics(m)
	return m, nil
}

func conditionalInt(cond bool, v int) int {
	if cond {
		return v
	}
	return 0
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
