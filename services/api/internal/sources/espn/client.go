// Package espn is the shared client for ESPN's public site API
// (https://site.api.espn.com/apis/site/v2/sports), which backs the NBA, WNBA,
// NFL, MLS, NWSL and NCAA sources.
//
// IMPORTANT: this API is unofficial and undocumented. It needs no key and has
// been stable for years, but ESPN does not publish terms for it and its
// general terms restrict commercial use of ESPN data, and the shape may change
// without notice. Budget a licensed feed (Sportradar, Stats Perform, the
// leagues' own) before monetizing the app. See
// .claude/skills/ops/references/data-sources.md for why each league uses it.
//
// The per-league packages (internal/sources/nba, wnba, nfl, mls, nwsl, ncaa)
// are thin: a team-id map, a metric builder and a registration.
package espn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/retr0h/freebie/services/api/internal/httputil"
	"github.com/retr0h/freebie/services/api/internal/ttlcache"
)

const (
	// DefaultBaseURL is ESPN's public site API root.
	DefaultBaseURL = "https://site.api.espn.com/apis/site/v2/sports"
	userAgent      = "freebies-game-checker/1.0"
	cacheTTL       = 10 * time.Minute
	cacheMax       = 256 // bodies kept; each is a scoreboard or game summary
	maxBody        = 16 << 20
)

// Client fetches and caches ESPN scoreboards and game summaries.
type Client struct {
	baseURL string
	http    *http.Client
	loc     *time.Location

	// cache holds response bodies by URL for cacheTTL, at most cacheMax, and
	// shares one in-flight fetch per URL; other URLs fetch concurrently.
	cache *ttlcache.Cache[[]byte]
}

// NewClient returns a client for the public ESPN API. Requests retry with
// backoff on network errors, 429 and 5xx, since trigger checks run once a day
// and a single failed call loses that day's deals.
func NewClient() *Client {
	return NewClientWith(DefaultBaseURL, &http.Client{
		Timeout:   30 * time.Second,
		Transport: &httputil.RetryTransport{Options: &httputil.RetryOptions{}},
	})
}

// NewClientWith builds a client against baseURL using hc. Tests point it at an
// httptest server.
func NewClientWith(baseURL string, hc *http.Client) *Client {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		loc = time.FixedZone("PT", -8*3600)
	}
	return &Client{baseURL: baseURL, http: hc, loc: loc, cache: ttlcache.New[[]byte](cacheTTL, cacheMax)}
}

// SetTransport replaces the HTTP transport; tests use it for a fast backoff.
func (c *Client) SetTransport(rt http.RoundTripper) { c.http.Transport = rt }

// Feed describes one ESPN scoreboard to read for a league.
type Feed struct {
	// Name labels the sport ("football", "basketball"); it prefixes game ids
	// when a league has several feeds.
	Name string
	// Path is "{sport}/{league}", e.g. "basketball/nba".
	Path string
	// Query adds scoreboard parameters, e.g. groups and limit for college.
	Query url.Values
}

// Games returns the events that started on the given Pacific calendar day
// (only the Y/M/D of date are used). ESPN files late games under its own
// (Eastern-ish) day, so the next day is fetched too and the result filtered
// by Pacific start time; a 9pm Pacific kickoff is then never missed.
func (c *Client) Games(ctx context.Context, f Feed, date time.Time) ([]Event, error) {
	day := time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, time.UTC)
	want := day.Format("2006-01-02")

	var out []Event
	seen := map[string]bool{}
	for _, d := range []time.Time{day, day.AddDate(0, 0, 1)} {
		evs, err := c.scoreboard(ctx, f, d.Format("20060102"))
		if err != nil {
			return nil, err
		}
		for _, e := range evs {
			if seen[e.ID] || e.Date.IsZero() || e.Date.In(c.loc).Format("2006-01-02") != want {
				continue
			}
			seen[e.ID] = true
			out = append(out, e)
		}
	}
	return out, nil
}

func (c *Client) scoreboard(ctx context.Context, f Feed, day string) ([]Event, error) {
	q := url.Values{}
	for k, v := range f.Query {
		q[k] = v
	}
	q.Set("dates", day)
	body, err := c.get(ctx, f.Path+"/scoreboard", q)
	if err != nil {
		return nil, fmt.Errorf("fetching ESPN %s scoreboard for %s: %w", f.Path, day, err)
	}
	var sr scoreboardResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("decoding ESPN %s scoreboard for %s: %w", f.Path, day, err)
	}
	return sr.Events, nil
}

// TeamStat reads one box-score team statistic (e.g. "interceptions") for
// teamID from the game summary, as an integer.
func (c *Client) TeamStat(ctx context.Context, path, eventID, teamID, name string) (int, error) {
	sum, err := c.summary(ctx, path, eventID)
	if err != nil {
		return 0, err
	}
	for _, t := range sum.Boxscore.Teams {
		if t.Team.ID != teamID {
			continue
		}
		for _, st := range t.Statistics {
			if st.Name == name {
				v, err := strconv.Atoi(st.DisplayValue)
				if err != nil {
					return 0, fmt.Errorf("ESPN summary %s: stat %s=%q is not a number",
						eventID, name, st.DisplayValue)
				}
				return v, nil
			}
		}
	}
	return 0, fmt.Errorf("ESPN summary %s: no %s stat for team %s", eventID, name, teamID)
}

// summary returns the game summary for event id.
func (c *Client) summary(ctx context.Context, path, id string) (*summaryResponse, error) {
	body, err := c.get(ctx, path+"/summary", url.Values{"event": {id}})
	if err != nil {
		return nil, fmt.Errorf("fetching ESPN %s summary %s: %w", path, id, err)
	}
	var sr summaryResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("decoding ESPN %s summary %s: %w", path, id, err)
	}
	return &sr, nil
}

// get fetches path?query, caching the body so one check run over many events
// does not refetch the same page. Concurrent callers of one URL share a single
// request (and its retries); different URLs do not wait on each other.
func (c *Client) get(ctx context.Context, path string, q url.Values) ([]byte, error) {
	u := c.baseURL + "/" + path + "?" + q.Encode()
	return c.cache.Get(ctx, u, func(ctx context.Context) ([]byte, error) {
		return c.fetch(ctx, u)
	})
}

func (c *Client) fetch(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}
	return body, nil
}
