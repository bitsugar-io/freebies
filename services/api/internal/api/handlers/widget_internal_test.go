package handlers

import (
	"testing"
	"time"
)

func pt(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.ParseInLocation("2006-01-02 15:04", s, laLocation)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

var shortNames = map[string]string{"LAR": "Rams", "LAL": "Lakers", "ANA": "Ducks"}

func TestWidgetNewestWinAndCount(t *testing.T) {
	now := pt(t, "2026-11-17 09:00")
	wins := []widgetWin{
		{ID: "te-ducks", TeamID: "ANA", OfferName: "$2 McGriddles", PartnerName: "McDonald's",
			TriggeredAt: pt(t, "2026-11-15 22:00"), ExpiresAt: pt(t, "2026-11-17 23:59")},
		{ID: "te-rams", TeamID: "LAR", OfferName: "Free Famous Star", PartnerName: "Carl's Jr.",
			TriggeredAt: pt(t, "2026-11-16 20:00"), AvailableAt: pt(t, "2026-11-18 00:00"), ExpiresAt: pt(t, "2026-11-22 23:59")},
		{ID: "te-old", TeamID: "LAL", OfferName: "Tacos", PartnerName: "JIB",
			TriggeredAt: pt(t, "2026-11-16 23:00"), ExpiresAt: pt(t, "2026-11-17 08:00")}, // expired
	}
	got := widgetFor(now, wins, nil, true, shortNames)
	if got.State != "won" || got.Title != "RAMS WON! +1" || got.Line1 != "Free Famous Star" ||
		got.Line2 != "Carl's Jr. · code works Wed" || got.Inline != "Rams won · Free Famous Star" ||
		got.Short != "RAMS" || got.URL != "freebie://deal/te-rams" {
		t.Errorf("got %+v", got)
	}
	// The next change is 9:30 (the 30-minute cap), before midnight's unlock.
	if got.RefreshAt != now.Add(widgetMaxWait).UTC().Format(time.RFC3339) {
		t.Errorf("refreshAt = %s", got.RefreshAt)
	}
}

func TestWinStatus(t *testing.T) {
	now := pt(t, "2026-11-17 09:00")
	for _, c := range []struct {
		win  widgetWin
		want string
	}{
		{widgetWin{AvailableAt: pt(t, "2026-11-17 10:30"), ExpiresAt: pt(t, "2026-11-17 23:59")}, "code works 10:30am"},
		{widgetWin{ExpiresAt: pt(t, "2026-11-17 23:59")}, "ends 11:59pm"},
		{widgetWin{ExpiresAt: pt(t, "2026-11-18 12:00")}, "1 day left"},
		{widgetWin{ExpiresAt: pt(t, "2026-11-22 23:59")}, "5 days left"},
	} {
		if got := winStatus(now, c.win); got != c.want {
			t.Errorf("winStatus(%+v) = %q, want %q", c.win, got, c.want)
		}
	}
}

func TestWidgetGameDay(t *testing.T) {
	now := pt(t, "2026-11-17 12:00")
	start := pt(t, "2026-11-17 19:00")
	games := []TodayGameResponse{
		{TeamID: "LAL", TeamName: "Lakers", Opponent: "Hornets", Home: true, State: "pre",
			StartsAt: start.UTC().Format(time.RFC3339),
			Deals:    []EventResponse{{PartnerName: "Jack in the Box"}, {PartnerName: "Del Taco"}}},
	}
	got := widgetFor(now, nil, games, true, shortNames)
	if got.State != "game" || got.Title != "LAKERS 7PM" || got.Line1 != "vs Hornets" ||
		got.Line2 != "Deal on it: Jack in the Box +1" || got.Short != "7PM" || got.URL != "freebie://today" {
		t.Errorf("got %+v", got)
	}

	games[0].State = "in"
	if got := widgetFor(now, nil, games, true, shortNames); got.Title != "LAKERS LIVE" ||
		got.RefreshAt != now.Add(widgetGameWait).UTC().Format(time.RFC3339) {
		t.Errorf("live game: %+v", got)
	}
}

func TestWidgetEmpty(t *testing.T) {
	now := pt(t, "2026-11-17 12:00")
	if got := widgetFor(now, nil, nil, false, shortNames); got.URL != "freebie://teams" || got.Title != "FREEBIES" {
		t.Errorf("no follows: %+v", got)
	}
	finished := []TodayGameResponse{{TeamID: "LAL", State: "post", Deals: []EventResponse{{PartnerName: "JIB"}}}}
	if got := widgetFor(now, nil, finished, true, shortNames); got.Title != "NO WINS YET" || got.URL != "freebie://today" {
		t.Errorf("following, nothing today: %+v", got)
	}
}
