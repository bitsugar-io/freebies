package worker

import (
	"strings"
	"testing"
	"time"
)

func TestUnlockedMessage(t *testing.T) {
	now := time.Date(2026, 11, 15, 22, 0, 0, 0, pacific)
	tests := []struct {
		name      string
		in        dealCopy
		wantTitle string
		wantBody  string
	}{
		{
			"deadline tomorrow morning",
			dealCopy{Partner: "McDonald's", Offer: "$2 McGriddles", Condition: "Ducks win at home",
				ExpiresAt: time.Date(2026, 11, 16, 10, 30, 0, 0, pacific), HasExpiry: true},
			"Ducks win at home! $2 McGriddles", "Claim at McDonald's by tomorrow 10:30am.",
		},
		{
			"deadline same day on the hour",
			dealCopy{Partner: "Ono", Offer: "Plate", Condition: "LAFC score",
				ExpiresAt: time.Date(2026, 11, 15, 23, 0, 0, 0, pacific), HasExpiry: true},
			"LAFC score! Plate", "Claim at Ono by 11pm.",
		},
		{
			"later deadline is dated with 11:59pm",
			dealCopy{Partner: "P", Offer: "O", Condition: "C",
				ExpiresAt: time.Date(2026, 11, 20, 23, 59, 0, 0, pacific), HasExpiry: true},
			"C! O", "Claim at P by Fri Nov 20 11:59pm.",
		},
		{
			"delayed code names the unlock day and a dated deadline",
			dealCopy{Partner: "Carl's Jr.", Offer: "Free Famous Star", Condition: "Rams pick one off",
				AvailableAt: time.Date(2026, 11, 17, 0, 0, 0, 0, pacific),
				ExpiresAt:   time.Date(2026, 11, 23, 23, 59, 59, 0, pacific), HasExpiry: true},
			"Rams pick one off! Free Famous Star", "Code unlocks Tue. Claim at Carl's Jr. by Mon Nov 23 11:59pm.",
		},
		{
			"a legacy midnight deadline reads as the end of the day before",
			dealCopy{Partner: "P", Offer: "O", Condition: "C",
				ExpiresAt: time.Date(2026, 11, 24, 0, 0, 0, 0, pacific), HasExpiry: true},
			"C! O", "Claim at P by Mon Nov 23 11:59pm.",
		},
		{
			"code unlocking tomorrow",
			dealCopy{Partner: "P", Offer: "O", Condition: "C",
				AvailableAt: time.Date(2026, 11, 16, 0, 0, 0, 0, pacific)},
			"C! O", "Code unlocks tomorrow. Claim at P.",
		},
		{
			"no condition or expiry",
			dealCopy{Partner: "P", Offer: "O"},
			"O", "Claim at P.",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			title, body := unlockedMessage(tc.in, now)
			if title != tc.wantTitle || body != tc.wantBody {
				t.Fatalf("got %q / %q, want %q / %q", title, body, tc.wantTitle, tc.wantBody)
			}
		})
	}
}

func TestReminderMessage(t *testing.T) {
	now := time.Date(2026, 11, 15, 20, 30, 0, 0, pacific)
	tests := []struct {
		name      string
		exp       time.Time
		wantTitle string
	}{
		{"hours", now.Add(3*time.Hour + 10*time.Minute), "Your $2 McGriddles expires in 3h"},
		{"minutes", now.Add(45 * time.Minute), "Your $2 McGriddles expires in 45m"},
		{"days", now.Add(72 * time.Hour), "Your $2 McGriddles expires in 3d"},
		{"expired clamps to a minute", now.Add(-time.Minute), "Your $2 McGriddles expires in 1m"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			title, _ := reminderMessage(dealCopy{Partner: "McDonald's", Offer: "$2 McGriddles",
				Condition: "Ducks win at home", ExpiresAt: tc.exp, HasExpiry: true}, now)
			if title != tc.wantTitle {
				t.Fatalf("got %q want %q", title, tc.wantTitle)
			}
		})
	}
}

// Item 6: a code that unlocks later the same Pacific day gives the clock time,
// not the day name; a later day keeps the day name.
func TestUnlockWordingSameDay(t *testing.T) {
	mon7am := time.Date(2026, 11, 16, 7, 0, 0, 0, pacific)
	c := dealCopy{Partner: "P", Offer: "O", Condition: "C",
		AvailableAt: time.Date(2026, 11, 16, 10, 0, 0, 0, pacific),
		ExpiresAt:   time.Date(2026, 11, 16, 23, 59, 59, 0, pacific), HasExpiry: true}
	if _, body := unlockedMessage(c, mon7am); body != "Code starts at 10am. Claim at P by 11:59pm." {
		t.Errorf("same-day unlock push body = %q", body)
	}
	if title, body := reminderMessage(c, mon7am); !strings.Contains(title+" "+body, "starts at 10am") ||
		strings.Contains(title+" "+body, "Mon") {
		t.Errorf("same-day reminder = %q / %q, want \"starts at 10am\" and no day name", title, body)
	}
	sat := time.Date(2026, 11, 14, 7, 0, 0, 0, pacific)
	if _, body := unlockedMessage(c, sat); !strings.HasPrefix(body, "Code unlocks Mon.") {
		t.Errorf("later-day unlock push body = %q, want the day name", body)
	}
	sun := time.Date(2026, 11, 15, 7, 0, 0, 0, pacific)
	if _, body := unlockedMessage(c, sun); !strings.HasPrefix(body, "Code unlocks tomorrow.") {
		t.Errorf("next-day unlock push body = %q, want tomorrow", body)
	}
}

func TestDealDeepLink(t *testing.T) {
	if got := dealDeepLink("triggered-x-1"); got != "freebie://deal/triggered-x-1" {
		t.Fatalf("got %q", got)
	}
}
