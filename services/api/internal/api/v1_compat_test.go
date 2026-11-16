package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/retr0h/freebie/services/api/internal/config"
	"github.com/retr0h/freebie/services/api/internal/db"
)

// The 1.x app in the App Store calls /api/v1 and can't be updated, so v1 only ever grows. These
// tests pin what that app reads, copied from apps/mobile/src/api/client.ts as it was on main
// before the LA redesign. Adding fields or routes is fine; removing, renaming or retyping one
// below breaks installed apps. If a change needs that, put it under a new version instead.

type kind int

const (
	str kind = iota
	num
	boolean
	object
)

type field struct {
	kind     kind
	optional bool
}

// v1 shapes, one per interface in the 1.x client.
var (
	v1League = map[string]field{
		"id": {kind: str}, "name": {kind: str}, "icon": {kind: str}, "displayOrder": {kind: num},
	}
	v1Event = map[string]field{
		"id": {kind: str}, "offerId": {kind: str}, "teamId": {kind: str}, "teamName": {kind: str},
		"league": {kind: str}, "partnerName": {kind: str}, "offerName": {kind: str},
		"offerDescription": {kind: str}, "triggerCondition": {kind: str}, "isActive": {kind: boolean},
		"teamColor": {kind: str, optional: true}, "icon": {kind: str, optional: true},
		"regionCode": {kind: str, optional: true}, "regionName": {kind: str, optional: true},
		"offerUrl": {kind: str, optional: true}, "affiliateUrl": {kind: str, optional: true},
		"affiliateTagline": {kind: str, optional: true},
	}
	v1User = map[string]field{
		"id": {kind: str}, "deviceId": {kind: str}, "platform": {kind: str},
		"pushToken": {kind: str, optional: true}, "token": {kind: str, optional: true},
	}
	v1Subscription = map[string]field{
		"id": {kind: str}, "userId": {kind: str}, "eventId": {kind: str}, "event": {kind: object},
	}
	v1ActiveDeal = map[string]field{
		"id": {kind: str}, "eventId": {kind: str}, "triggeredAt": {kind: str},
		"expiresAt": {kind: str, optional: true}, "event": {kind: object},
		"isDismissed": {kind: boolean}, "dismissalType": {kind: str, optional: true},
	}
	v1Dismissal = map[string]field{
		"id": {kind: str}, "userId": {kind: str}, "triggeredEventId": {kind: str},
		"type": {kind: str}, "dismissedAt": {kind: str},
	}
	v1Stats = map[string]field{"dealsClaimed": {kind: num}, "subscriptionsCount": {kind: num}}
)

// checkShape reports every field the 1.x app reads that is missing or the wrong JSON type. An
// optional string must be omitted rather than sent empty: the 1.x screens render
// `{value && <Text>…}`, and an empty string there is a bare text node outside <Text>.
func checkShape(t *testing.T, where string, obj map[string]any, shape map[string]field) {
	t.Helper()
	for name, f := range shape {
		v, ok := obj[name]
		if !ok || v == nil {
			if !f.optional {
				t.Errorf("%s: missing required field %q", where, name)
			}
			continue
		}
		var good bool
		switch f.kind {
		case str:
			s, isStr := v.(string)
			good = isStr
			if isStr && f.optional && s == "" {
				t.Errorf("%s: optional field %q is an empty string; omit it instead", where, name)
			}
		case num:
			_, good = v.(float64)
		case boolean:
			_, good = v.(bool)
		case object:
			_, good = v.(map[string]any)
		}
		if !good {
			t.Errorf("%s: field %q has JSON type %T, want %v", where, name, v, f.kind)
		}
	}
}

type v1Client struct {
	t       *testing.T
	handler http.Handler
	token   string
}

func (c *v1Client) do(method, path string, body any) (int, []byte) {
	c.t.Helper()
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			c.t.Fatal(err)
		}
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "/api/v1"+path, r)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

func (c *v1Client) object(method, path string, body any, want int) map[string]any {
	c.t.Helper()
	code, b := c.do(method, path, body)
	if code != want {
		c.t.Fatalf("%s %s: status %d, want %d: %s", method, path, code, want, b)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		c.t.Fatalf("%s %s: %v: %s", method, path, err, b)
	}
	return out
}

func (c *v1Client) list(path string) []map[string]any {
	c.t.Helper()
	code, b := c.do(http.MethodGet, path, nil)
	if code != http.StatusOK {
		c.t.Fatalf("GET %s: status %d: %s", path, code, b)
	}
	var out []map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		c.t.Fatalf("GET %s: %v: %s", path, err, b)
	}
	return out
}

// newV1Client boots the real router on a fully migrated database, so the checks see production
// seed data (the deals, leagues and flags that migrations ship).
func newV1Client(t *testing.T) (*v1Client, *db.Queries) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	srv := NewServer(&config.Config{}, database, logger, nil)
	return &v1Client{t: t, handler: srv.Router()}, db.New(database)
}

// TestV1CompatPublicRoutes covers what the 1.x app loads before it has a user.
func TestV1CompatPublicRoutes(t *testing.T) {
	c, _ := newV1Client(t)

	for i, l := range c.list("/leagues") {
		checkShape(t, fmt.Sprintf("GET /leagues [%d]", i), l, v1League)
	}

	events := c.list("/events")
	if len(events) == 0 {
		t.Fatal("GET /events returned no events")
	}
	for _, e := range events {
		checkShape(t, fmt.Sprintf("GET /events %v", e["id"]), e, v1Event)
		id, _ := e["id"].(string)
		checkShape(t, "GET /events/"+id, c.object(http.MethodGet, "/events/"+id, nil, http.StatusOK), v1Event)
	}

	cfg := c.object(http.MethodGet, "/config", nil, http.StatusOK)
	features, ok := cfg["features"].(map[string]any)
	if !ok {
		t.Fatalf("GET /config: features is %T, want an object", cfg["features"])
	}
	screens, ok := cfg["screens"].(map[string]any)
	if !ok {
		t.Fatalf("GET /config: screens is %T, want an object", cfg["screens"])
	}
	// The 1.x tabs render these screens' blocks; a missing screen is an empty tab.
	for _, name := range []string{"deals", "discover", "profile"} {
		blocks, ok := screens[name].([]any)
		if !ok || len(blocks) == 0 {
			t.Errorf("GET /config: screen %q has no blocks", name)
			continue
		}
		for i, b := range blocks {
			checkShape(t, fmt.Sprintf("GET /config screens.%s[%d]", name, i), b.(map[string]any),
				map[string]field{"type": {kind: str}, "key": {kind: str}, "config": {kind: object}})
		}
	}

	// The 1.x deal list hides a league whose enable_<league> flag is false, so every league with
	// an active deal must not be switched off, or old installs never see that deal.
	for _, e := range events {
		if e["isActive"] != true {
			continue
		}
		flag := "enable_" + strings.ToLower(e["league"].(string))
		if features[flag] == false {
			t.Errorf("active deal %v is in league %v, but %s is false, so the 1.x app hides it",
				e["id"], e["league"], flag)
		}
	}
}

// TestV1CompatUserRoutes walks the 1.x app's signed-in flow: create a user, follow every active
// deal, see it unlock, dismiss it, read stats.
func TestV1CompatUserRoutes(t *testing.T) {
	c, q := newV1Client(t)

	user := c.object(http.MethodPost, "/users", map[string]string{"deviceId": "dev-1x", "platform": "ios"}, http.StatusCreated)
	checkShape(t, "POST /users", user, v1User)
	userID, _ := user["id"].(string)
	c.token, _ = user["token"].(string)
	if c.token == "" {
		t.Fatal("POST /users returned no token")
	}
	checkShape(t, "GET /users/{id}", c.object(http.MethodGet, "/users/"+userID, nil, http.StatusOK), v1User)

	code, b := c.do(http.MethodPut, "/users/"+userID+"/push-token", map[string]string{"pushToken": "ExponentPushToken[x]"})
	if code >= 300 {
		t.Fatalf("PUT push-token: status %d: %s", code, b)
	}

	var active []string
	for _, e := range c.list("/events") {
		if e["isActive"] == true {
			active = append(active, e["id"].(string))
		}
	}
	if len(active) == 0 {
		t.Fatal("no active deals to follow")
	}
	for _, id := range active {
		s := c.object(http.MethodPost, "/users/"+userID+"/subscriptions", map[string]string{"eventId": id}, http.StatusCreated)
		checkShape(t, "POST subscriptions", s, v1Subscription)
	}
	for _, s := range c.list("/users/" + userID + "/subscriptions") {
		checkShape(t, "GET subscriptions", s, v1Subscription)
		checkShape(t, fmt.Sprintf("GET subscriptions %v .event", s["eventId"]), s["event"].(map[string]any), v1Event)
	}

	// Unlock every followed deal the way the worker does.
	for i, id := range active {
		if _, err := q.CreateTriggeredEvent(t.Context(), db.CreateTriggeredEventParams{
			ID: fmt.Sprintf("te-%d", i), EventID: id,
			ExpiresAt: sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
		}); err != nil {
			t.Fatalf("trigger %s: %v", id, err)
		}
	}
	deals := c.list("/users/" + userID + "/active-deals")
	if len(deals) != len(active) {
		t.Fatalf("GET active-deals: %d deals, want %d", len(deals), len(active))
	}
	for _, d := range deals {
		checkShape(t, fmt.Sprintf("GET active-deals %v", d["id"]), d, v1ActiveDeal)
		checkShape(t, fmt.Sprintf("GET active-deals %v .event", d["id"]), d["event"].(map[string]any), v1Event)
	}

	te, _ := deals[0]["id"].(string)
	dis := c.object(http.MethodPost, "/users/"+userID+"/dismissals",
		map[string]string{"triggeredEventId": te, "type": "got_it"}, http.StatusCreated)
	checkShape(t, "POST dismissals", dis, v1Dismissal)

	checkShape(t, "GET stats", c.object(http.MethodGet, "/users/"+userID+"/stats", nil, http.StatusOK), v1Stats)

	if code, b := c.do(http.MethodDelete, "/users/"+userID+"/dismissals/"+te, nil); code >= 300 {
		t.Fatalf("DELETE dismissal: status %d: %s", code, b)
	}
	if code, b := c.do(http.MethodDelete, "/users/"+userID+"/subscriptions/"+active[0], nil); code >= 300 {
		t.Fatalf("DELETE subscription: status %d: %s", code, b)
	}
}
