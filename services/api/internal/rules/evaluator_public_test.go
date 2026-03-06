package rules_test

import (
	"testing"

	"github.com/retr0h/freebie/services/api/internal/rules"
	"github.com/stretchr/testify/suite"
)

type EvaluatorPublicTestSuite struct {
	suite.Suite
}

func TestEvaluatorPublicTestSuite(
	t *testing.T,
) {
	suite.Run(t, new(EvaluatorPublicTestSuite))
}

func (s *EvaluatorPublicTestSuite) TestEvaluate() {
	tests := []struct {
		name      string
		condition map[string]interface{}
		eventData map[string]interface{}
		want      bool
	}{
		// Numeric comparisons - gte
		{
			name: "gte: 7+ strikeouts matches 10",
			condition: map[string]interface{}{
				"team_strikeouts": map[string]interface{}{
					"gte": 7,
				},
			},
			eventData: map[string]interface{}{
				"team_strikeouts": 10,
			},
			want: true,
		},
		{
			name: "gte: 7+ strikeouts does not match 5",
			condition: map[string]interface{}{
				"team_strikeouts": map[string]interface{}{
					"gte": 7,
				},
			},
			eventData: map[string]interface{}{
				"team_strikeouts": 5,
			},
			want: false,
		},
		{
			name: "gte: precipitation >= 0.5",
			condition: map[string]interface{}{
				"precipitation_mm": map[string]interface{}{
					"gte": 0.5,
				},
			},
			eventData: map[string]interface{}{
				"precipitation_mm": 2.3,
			},
			want: true,
		},
		// Numeric comparisons - gt
		{
			name: "gt: temperature > 75",
			condition: map[string]interface{}{
				"temperature_f": map[string]interface{}{
					"gt": 75,
				},
			},
			eventData: map[string]interface{}{
				"temperature_f": 80,
			},
			want: true,
		},
		{
			name: "gt: temperature > 75 does not match 75",
			condition: map[string]interface{}{
				"temperature_f": map[string]interface{}{
					"gt": 75,
				},
			},
			eventData: map[string]interface{}{
				"temperature_f": 75,
			},
			want: false,
		},
		// Numeric comparisons - lte
		{
			name: "lte: temperature <= 60",
			condition: map[string]interface{}{
				"temperature_f": map[string]interface{}{
					"lte": 60,
				},
			},
			eventData: map[string]interface{}{
				"temperature_f": 55,
			},
			want: true,
		},
		// Numeric comparisons - lt
		{
			name: "lt: temperature < 60",
			condition: map[string]interface{}{
				"temperature_f": map[string]interface{}{
					"lt": 60,
				},
			},
			eventData: map[string]interface{}{
				"temperature_f": 55,
			},
			want: true,
		},
		// Numeric comparisons - eq
		{
			name: "eq: runs equals 5",
			condition: map[string]interface{}{
				"team_runs": map[string]interface{}{
					"eq": 5,
				},
			},
			eventData: map[string]interface{}{
				"team_runs": 5,
			},
			want: true,
		},
		// Boolean conditions
		{
			name: "win condition matches true",
			condition: map[string]interface{}{
				"win": true,
			},
			eventData: map[string]interface{}{
				"win": true,
			},
			want: true,
		},
		{
			name: "win condition does not match false",
			condition: map[string]interface{}{
				"win": true,
			},
			eventData: map[string]interface{}{
				"win": false,
			},
			want: false,
		},
		{
			name: "is_home condition matches false",
			condition: map[string]interface{}{
				"is_home": false,
			},
			eventData: map[string]interface{}{
				"is_home": false,
			},
			want: true,
		},
		// String conditions
		{
			name: "contains operator matches",
			condition: map[string]interface{}{
				"condition": map[string]interface{}{
					"contains": "rain",
				},
			},
			eventData: map[string]interface{}{
				"condition": "rainy",
			},
			want: true,
		},
		{
			name: "contains operator does not match",
			condition: map[string]interface{}{
				"condition": map[string]interface{}{
					"contains": "rain",
				},
			},
			eventData: map[string]interface{}{
				"condition": "sunny",
			},
			want: false,
		},
		{
			name: "exact string match",
			condition: map[string]interface{}{
				"category": "community",
			},
			eventData: map[string]interface{}{
				"category": "community",
			},
			want: true,
		},
		{
			name: "exact string does not match",
			condition: map[string]interface{}{
				"category": "community",
			},
			eventData: map[string]interface{}{
				"category": "sports",
			},
			want: false,
		},
		// Multiple conditions
		{
			name: "multiple conditions all match",
			condition: map[string]interface{}{
				"team_strikeouts": map[string]interface{}{
					"gte": 7,
				},
				"win": true,
			},
			eventData: map[string]interface{}{
				"team_strikeouts": 10,
				"win":             true,
			},
			want: true,
		},
		{
			name: "one condition fails",
			condition: map[string]interface{}{
				"team_strikeouts": map[string]interface{}{
					"gte": 7,
				},
				"win": true,
			},
			eventData: map[string]interface{}{
				"team_strikeouts": 10,
				"win":             false,
			},
			want: false,
		},
		{
			name: "weather event - rain and temperature",
			condition: map[string]interface{}{
				"precipitation_mm": map[string]interface{}{
					"gte": 5.0,
				},
				"temperature_f": map[string]interface{}{
					"lt": 60,
				},
			},
			eventData: map[string]interface{}{
				"precipitation_mm": 7.2,
				"temperature_f":    55,
			},
			want: true,
		},
		{
			name: "all numeric comparisons",
			condition: map[string]interface{}{
				"team_runs": map[string]interface{}{
					"gte": 3,
				},
				"team_hits": map[string]interface{}{
					"gte": 8,
				},
				"team_errors": map[string]interface{}{
					"lte": 1,
				},
			},
			eventData: map[string]interface{}{
				"team_runs":   5,
				"team_hits":   10,
				"team_errors": 0,
			},
			want: true,
		},
		// Missing fields
		{
			name: "missing required field returns false",
			condition: map[string]interface{}{
				"team_strikeouts": map[string]interface{}{
					"gte": 7,
				},
			},
			eventData: map[string]interface{}{
				"team_runs": 5,
			},
			want: false,
		},
		{
			name: "empty event data returns false",
			condition: map[string]interface{}{
				"win": true,
			},
			eventData: map[string]interface{}{},
			want:      false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := rules.Evaluate(tt.condition, tt.eventData)
			s.Equal(tt.want, got)
		})
	}
}
