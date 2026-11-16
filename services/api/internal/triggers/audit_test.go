package triggers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/sources"
	"github.com/retr0h/freebie/services/api/internal/triggers"

	// Every real source, as the server registers them.
	_ "github.com/retr0h/freebie/services/api/internal/sources/mlb"
	_ "github.com/retr0h/freebie/services/api/internal/sources/mls"
	_ "github.com/retr0h/freebie/services/api/internal/sources/nba"
	_ "github.com/retr0h/freebie/services/api/internal/sources/ncaa"
	_ "github.com/retr0h/freebie/services/api/internal/sources/nfl"
	_ "github.com/retr0h/freebie/services/api/internal/sources/nhl"
	_ "github.com/retr0h/freebie/services/api/internal/sources/nwsl"
	_ "github.com/retr0h/freebie/services/api/internal/sources/wnba"
)

// TestSeededActiveDealsCanFire runs every migration and fails if any active
// seeded deal has no source, a metric its source never emits, a team its
// source cannot map, or a rule that never evaluates. It guards future deal
// migrations: a deal that can never fire must ship inactive.
func TestSeededActiveDealsCanFire(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	events, err := db.New(database).ListEvents(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	rows := triggers.Audit(events)
	active := 0
	for _, r := range rows {
		if r.Active {
			active++
		}
	}
	if active == 0 {
		t.Fatal("no active events seeded; the audit is not testing anything")
	}
	for _, r := range triggers.Failures(rows) {
		t.Errorf("active deal %s can never fire: %s: %s", r.EventID, r.Status, r.Detail)
	}
}

type auditSource struct{ metrics, teams []string }

func (auditSource) League() string      { return "auditleague" }
func (a auditSource) Metrics() []string { return a.metrics }
func (a auditSource) Teams() []string   { return a.teams }
func (auditSource) GetYesterdaysGame(context.Context, string) (*sources.GameStats, error) {
	return nil, nil
}
func (auditSource) GetGameByDate(context.Context, string, time.Time) (*sources.GameStats, error) {
	return nil, nil
}

func TestAuditStatuses(t *testing.T) {
	sources.Register(auditSource{metrics: []string{"win"}, teams: []string{"AAA"}})
	rule := func(metric, op string) sql.NullString {
		return sql.NullString{Valid: true, String: `{"metric":"` + metric + `","operator":"` + op + `","value":1}`}
	}
	ev := func(id, league, team string, active int64, r sql.NullString) db.Event {
		return db.Event{ID: id, League: league, TeamID: team, IsActive: active, TriggerRule: r}
	}
	events := []db.Event{
		ev("ok", "auditleague", "AAA", 1, rule("win", "==")),
		ev("nosrc", "nope", "AAA", 1, rule("win", "==")),
		ev("metric", "auditleague", "AAA", 1, rule("goals", ">=")),
		ev("team", "auditleague", "ZZZ", 1, rule("win", "==")),
		ev("op", "auditleague", "AAA", 1, rule("win", "=>")),
		ev("norule", "auditleague", "AAA", 1, sql.NullString{}),
		ev("off", "nope", "AAA", 0, rule("win", "==")),
		ev("offok", "auditleague", "AAA", 0, rule("win", "==")),
	}
	want := map[string]triggers.AuditStatus{
		"ok": triggers.AuditOK, "nosrc": triggers.AuditNoSource,
		"metric": triggers.AuditUnknownMetric, "team": triggers.AuditUnknownTeam,
		"op": triggers.AuditBadRule, "norule": triggers.AuditBadRule,
		"off": triggers.AuditInactive, "offok": triggers.AuditInactive,
	}
	rows := triggers.Audit(events)
	for _, r := range rows {
		if r.Status != want[r.EventID] {
			t.Errorf("%s: status %s, want %s", r.EventID, r.Status, want[r.EventID])
		}
	}
	if got := len(triggers.Failures(rows)); got != 5 {
		t.Errorf("failures = %d, want 5 (inactive events never fail)", got)
	}
	for _, r := range rows {
		if r.EventID == "off" && r.Detail == "" {
			t.Error("an inactive event should say what would be wrong with it")
		}
	}
}

// A rule's own fields are checked too: a misspelled window, a malformed clock
// time, an unknown season, or season "regular" on a source that can't tell
// playoff games apart (auditSource here) all fail; an empty window is fine.
func TestAuditRuleFields(t *testing.T) {
	sources.Register(auditSource{metrics: []string{"win"}, teams: []string{"AAA"}})
	ev := func(id, ruleJSON string) db.Event {
		return db.Event{ID: id, League: "auditleague", TeamID: "AAA", IsActive: 1,
			TriggerRule: sql.NullString{Valid: true, String: ruleJSON}}
	}
	cases := map[string]struct {
		rule string
		want triggers.AuditStatus
	}{
		"empty window":   {`{"metric":"win","operator":"==","value":1}`, triggers.AuditOK},
		"through":        {`{"metric":"win","operator":"==","value":1,"redemption_window":"through_next_business_day"}`, triggers.AuditOK},
		"start time":     {`{"metric":"win","operator":"==","value":1,"available_from_local":"10:00"}`, triggers.AuditOK},
		"typo window":    {`{"metric":"win","operator":"==","value":1,"redemption_window":"sameday"}`, triggers.AuditBadRule},
		"bad time":       {`{"metric":"win","operator":"==","value":1,"expires_at_local":"10am"}`, triggers.AuditBadRule},
		"bad season":     {`{"metric":"win","operator":"==","value":1,"season":"playoffs"}`, triggers.AuditBadRule},
		"season unknown": {`{"metric":"win","operator":"==","value":1,"season":"regular"}`, triggers.AuditBadRule},
	}
	for name, tc := range cases {
		rows := triggers.Audit([]db.Event{ev(name, tc.rule)})
		if rows[0].Status != tc.want {
			t.Errorf("%s: status %s (%s), want %s", name, rows[0].Status, rows[0].Detail, tc.want)
		}
	}
}
