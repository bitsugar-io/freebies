package sources

import (
	"context"
	"testing"
	"time"
)

type fakeSource struct{ league string }

func (f fakeSource) League() string    { return f.league }
func (f fakeSource) Metrics() []string { return nil }
func (f fakeSource) GetYesterdaysGame(context.Context, string) (*GameStats, error) {
	return nil, nil
}
func (f fakeSource) GetGameByDate(context.Context, string, time.Time) (*GameStats, error) {
	return nil, nil
}

func TestGetIsCaseInsensitive(t *testing.T) {
	Register(fakeSource{league: "Fake"})
	for _, name := range []string{"fake", "FAKE", "Fake", " fake "} {
		if _, err := Get(name); err != nil {
			t.Errorf("Get(%q): %v", name, err)
		}
	}
	if _, err := Get("nope"); err == nil {
		t.Error("expected error for unregistered league")
	}
}
