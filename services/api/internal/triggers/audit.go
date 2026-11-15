package triggers

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/retr0h/freebie/services/api/internal/db"
	"github.com/retr0h/freebie/services/api/internal/sources"
)

// AuditStatus says whether a deal can ever fire.
type AuditStatus string

// Audit statuses. Anything but OK on an active event means the deal can never
// fire.
const (
	AuditOK            AuditStatus = "OK"
	AuditNoSource      AuditStatus = "NO_SOURCE"
	AuditUnknownMetric AuditStatus = "UNKNOWN_METRIC"
	AuditUnknownTeam   AuditStatus = "UNKNOWN_TEAM"
	AuditBadRule       AuditStatus = "BAD_RULE"
	AuditInactive      AuditStatus = "INACTIVE"
)

// AuditRow is one event's verdict.
type AuditRow struct {
	EventID string
	League  string
	TeamID  string
	Metric  string
	Active  bool
	Status  AuditStatus
	// Detail explains a non-OK status. For an inactive event it describes
	// what would be wrong if the event were turned on ("" when nothing).
	Detail string
}

// Failing reports whether the row is an active event that can never fire.
func (r AuditRow) Failing() bool { return r.Active && r.Status != AuditOK }

// Audit checks every event against the registered sources. An active event is
// OK only when its league has a source, the source lists the rule's metric,
// the source maps the event's team and the rule parses with a known operator.
// Inactive events are reported INACTIVE, with the problem they would have in
// Detail.
func Audit(events []db.Event) []AuditRow {
	rows := make([]AuditRow, 0, len(events))
	for _, e := range events {
		status, metric, detail := auditEvent(e)
		row := AuditRow{
			EventID: e.ID, League: e.League, TeamID: e.TeamID, Metric: metric,
			Active: e.IsActive != 0, Status: status, Detail: detail,
		}
		if e.IsActive == 0 {
			row.Status = AuditInactive
			if status != AuditOK {
				row.Detail = string(status) + ": " + detail
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// Failures returns the active events that can never fire.
func Failures(rows []AuditRow) []AuditRow {
	var out []AuditRow
	for _, r := range rows {
		if r.Failing() {
			out = append(out, r)
		}
	}
	return out
}

func auditEvent(e db.Event) (AuditStatus, string, string) {
	if !e.TriggerRule.Valid || e.TriggerRule.String == "" {
		return AuditBadRule, "", "no trigger_rule"
	}
	var rule Rule
	if err := json.Unmarshal([]byte(e.TriggerRule.String), &rule); err != nil {
		return AuditBadRule, "", fmt.Sprintf("trigger_rule is not valid JSON: %v", err)
	}
	if !validOperator(rule.Operator) {
		return AuditBadRule, rule.Metric, fmt.Sprintf("operator %q never triggers", rule.Operator)
	}
	if rule.UnlockAfterDays < 0 || rule.ValidDays < 0 {
		return AuditBadRule, rule.Metric, "unlock_after_days and valid_days cannot be negative"
	}
	src, err := sources.Get(e.League)
	if err != nil {
		return AuditNoSource, rule.Metric, fmt.Sprintf("no source registered for league %q", e.League)
	}
	if !slices.Contains(src.Metrics(), rule.Metric) {
		return AuditUnknownMetric, rule.Metric,
			fmt.Sprintf("the %s source does not emit metric %q (it emits %v)",
				e.League, rule.Metric, src.Metrics())
	}
	if tl, ok := src.(sources.TeamLister); ok && !slices.Contains(tl.Teams(), e.TeamID) {
		return AuditUnknownTeam, rule.Metric,
			fmt.Sprintf("the %s source has no mapping for team %q", e.League, e.TeamID)
	}
	return AuditOK, rule.Metric, ""
}

func validOperator(op string) bool {
	switch op {
	case ">=", ">", "<=", "<", "==", "=", "!=":
		return true
	}
	return false
}
