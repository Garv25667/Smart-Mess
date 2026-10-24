package httpapi

import (
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"smartmess/internal/rules"
	"smartmess/internal/store"
)

// kolkataLoc is the app timezone.
var kolkataLoc = rules.KolkataLoc

// parseKolkataDate parses a YYYY-MM-DD string as a service date in Asia/Kolkata.
func parseKolkataDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, kolkataLoc)
}

// todayInKolkata returns today's service date.
func todayInKolkata(now time.Time) time.Time {
	t := now.In(kolkataLoc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, kolkataLoc)
}

// toTimeOfDay extracts wall-clock values from a pgtype.Time column.
func toTimeOfDay(t pgtype.Time) rules.TimeOfDay {
	hour := int(t.Microseconds / 3_600_000_000)
	minute := int((t.Microseconds % 3_600_000_000) / 60_000_000)
	return rules.TimeOfDay{Hour: hour, Minute: minute}
}

// isValidMealType reports whether s is one of the four meals.
func isValidMealType(s string) bool {
	switch s {
	case "breakfast", "lunch", "snacks", "dinner":
		return true
	}
	return false
}

// isValidDietPref reports whether s is regular or jain (students can pick these only).
func isValidDietPref(s string) bool { return s == "regular" || s == "jain" }

// textScaleValue converts the stored numeric text_scale into a float64 (1.0 when unset).
func textScaleValue(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 1.0
	}
	f8, err := n.Float64Value()
	if err != nil || !f8.Valid {
		return 1.0
	}
	f := f8.Float64
	if f < 1.0 || f > 2.0 {
		return 1.0
	}
	return f
}

// floatToNumeric converts a float64 into pgtype.Numeric via its string form.
func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(strconv.FormatFloat(f, 'f', 2, 64))
	return n
}

// toRulesSchedules converts db schedule rows into rules.Schedule values.
func toRulesSchedules(rows []store.MealSchedules) []rules.Schedule {
	out := make([]rules.Schedule, 0, len(rows))
	for _, r := range rows {
		out = append(out, rules.Schedule{
			HostelGroup: string(r.HostelGroup),
			MealType:    string(r.MealType),
			DayKind:     string(r.DayKind),
			StartTime:   toTimeOfDay(r.StartTime),
			EndTime:     toTimeOfDay(r.EndTime),
		})
	}
	return out
}

// derefString returns *string value or "" when nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
