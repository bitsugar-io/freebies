package espn_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/retr0h/freebie/services/api/internal/sources/espn"
)

// summaryClient serves body for every summary request.
func summaryClient(t *testing.T, body string) *espn.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return espn.NewClientWith(srv.URL, srv.Client())
}

const kickingKeys = `["fieldGoalsMade/fieldGoalAttempts","fieldGoalPct","longFieldGoalMade",` +
	`"extraPointsMade/extraPointAttempts","totalKickingPoints"]`

func TestLongFieldGoal(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		points  int
		want    int
		wantErr string
	}{
		{
			name: "longest make",
			body: `{"boxscore":{"players":[{"team":{"id":"14"},"statistics":[{"name":"kicking",
				"keys":` + kickingKeys + `,"athletes":[{"stats":["2/3","66.7","56","3/3","9"]}]}]}]}}`,
			points: 30, want: 56,
		},
		{
			name: "kicker made no field goal",
			body: `{"boxscore":{"players":[{"team":{"id":"14"},"statistics":[{"name":"kicking",
				"keys":` + kickingKeys + `,"athletes":[{"stats":["0/1","0.0","0","3/3","3"]}]}]}]}}`,
			points: 21, want: 0,
		},
		{
			name: "shut out team with no kicker row",
			body: `{"boxscore":{"players":[{"team":{"id":"14"},"statistics":[{"name":"passing","keys":[],"athletes":[]}]},
				{"team":{"id":"25"},"statistics":[]}]}}`,
			points: 0, want: 0,
		},
		{
			name:    "team missing from the player box score",
			body:    `{"boxscore":{"players":[{"team":{"id":"25"},"statistics":[{"name":"kicking","keys":[],"athletes":[]}]}]}}`,
			points:  0,
			wantErr: "no player box score for team 14",
		},
		{
			name: "scoring team with no kicking category",
			body: `{"boxscore":{"players":[{"team":{"id":"14"},"statistics":[{"name":"passing","keys":[],"athletes":[]}]},
				{"team":{"id":"25"},"statistics":[]}]}}`,
			points:  17,
			wantErr: "no kicking stats",
		},
		{
			name:    "no player box score at all",
			body:    `{"boxscore":{"players":[]}}`,
			points:  10,
			wantErr: "no player box score",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := summaryClient(t, tc.body).LongFieldGoal(context.Background(), "football/nfl", "1", "14", tc.points)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// Every free throw counts toward the streak, technical, flagrant and
// clear-path ones included: the sponsors' terms say "miss two free throws in a
// row" with no exception.
func TestFreeThrowMissStreakCountsEveryFreeThrow(t *testing.T) {
	body := `{"plays":[
	 {"period":{"number":4},"type":{"text":"Free Throw - Technical"},"scoringPlay":false,"team":{"id":"99"}},
	 {"period":{"number":4},"type":{"text":"Free Throw - Flagrant 1 of 2"},"scoringPlay":false,"team":{"id":"99"}},
	 {"period":{"number":4},"type":{"text":"Free Throw - Clear Path 1 of 2"},"scoringPlay":false,"team":{"id":"99"}},
	 {"period":{"number":4},"type":{"text":"Free Throw - Clear Path 2 of 2"},"scoringPlay":true,"team":{"id":"99"}},
	 {"period":{"number":4},"type":{"text":"Free Throw - 1 of 2"},"scoringPlay":false,"team":{"id":"99"}}
	]}`
	got, err := summaryClient(t, body).FreeThrowMissStreak(context.Background(), "basketball/nba", "1", "99", 4)
	require.NoError(t, err)
	assert.Equal(t, 3, got, "technical, flagrant and clear-path misses all count")
}
