package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestDatabaseURLRedaction(t *testing.T) {
	cases := []struct {
		name string
		in   DatabaseURL
		want string
	}{
		{
			name: "turso url with authToken is redacted",
			in:   "libsql://freebie-retr0h.aws-us-east-1.turso.io?authToken=eyJabc.def.ghi",
			want: "libsql://freebie-retr0h.aws-us-east-1.turso.io?authToken=REDACTED",
		},
		{
			name: "sqlite file path passes through",
			in:   "freebie.db",
			want: "freebie.db",
		},
		{
			name: "url without authToken passes through",
			in:   "libsql://example.com",
			want: "libsql://example.com",
		},
		{
			name: "empty string passes through",
			in:   "",
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.in.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
			if got := fmt.Sprintf("%v", tc.in); got != tc.want {
				t.Errorf("Sprintf(%%v) = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDatabaseURLSlogRedaction(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	secret := DatabaseURL("libsql://host?authToken=SUPER_SECRET")
	logger.Info("test", "db", secret)

	out := buf.String()
	if strings.Contains(out, "SUPER_SECRET") {
		t.Errorf("slog leaked secret: %s", out)
	}
	if !strings.Contains(out, "REDACTED") {
		t.Errorf("slog output missing REDACTED marker: %s", out)
	}
}

func TestDatabaseURLJSONRedaction(t *testing.T) {
	secret := DatabaseURL("libsql://host?authToken=SUPER_SECRET")
	b, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(b), "SUPER_SECRET") {
		t.Errorf("json leaked secret: %s", b)
	}
}

func TestDatabaseURLCastReturnsRaw(t *testing.T) {
	raw := "libsql://host?authToken=NEEDED_FOR_OPEN"
	secret := DatabaseURL(raw)
	if got := string(secret); got != raw {
		t.Errorf("explicit string() cast lost data: got %q want %q", got, raw)
	}
}
