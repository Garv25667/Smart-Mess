// Package rules contains pure business-logic functions.
// They take "now" and plain data. No database, no HTTP.
package rules

import "time"

// KolkataLoc loads the Asia/Kolkata timezone.
// Import _ "time/tzdata" in main.go so this works in minimal containers.
var KolkataLoc *time.Location

func init() {
	var err error
	KolkataLoc, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fallback for minimal environments (should not happen with tzdata import).
		KolkataLoc = time.FixedZone("Asia/Kolkata", 5*3600+30*60)
	}
}

// --- Day kind ---

// DayKind returns "weekday" or "weekend" for a date in Asia/Kolkata.
func DayKind(t time.Time) string {
	d := t.In(KolkataLoc).Weekday()
	if d == time.Saturday || d == time.Sunday {
		return "weekend"
	}
	return "weekday"
}

// --- Meal schedules and windows ---

// Schedule represents a meal's timing for one day kind.
type Schedule struct {
	HostelGroup string
	MealType    string
	DayKind     string
	StartTime   TimeOfDay // wall-clock in Asia/Kolkata
	EndTime     TimeOfDay
}

// TimeOfDay is hours and minutes (wall-clock).
type TimeOfDay struct {
	Hour   int
	Minute int
}

// MealWindow returns the absolute start and end instants for a schedule on a given date.
func MealWindow(s Schedule, date time.Time) (start, end time.Time) {
	d := date.In(KolkataLoc)
	y, m, day := d.Date()
	start = time.Date(y, m, day, s.StartTime.Hour, s.StartTime.Minute, 0, 0, KolkataLoc)
	end = time.Date(y, m, day, s.EndTime.Hour, s.EndTime.Minute, 0, 0, KolkataLoc)
	return start, end
}

// MealOrder defines a sort-order for meals within a day.
var MealOrder = map[string]int{
	"breakfast": 0,
	"lunch":     1,
	"snacks":    2,
	"dinner":    3,
}

// CurrentMeal returns the schedule whose window contains `now`, or nil.
func CurrentMeal(schedules []Schedule, now time.Time) *Schedule {
	date := now.In(KolkataLoc)
	dk := DayKind(date)
	for i := range schedules {
		s := &schedules[i]
		if s.DayKind != dk {
			continue
		}
		start, end := MealWindow(*s, date)
		if !now.Before(start) && !now.After(end) {
			return s
		}
	}
	return nil
}

// NextMeal returns the first schedule that starts after `now` today, or nil.
func NextMeal(schedules []Schedule, now time.Time) *Schedule {
	date := now.In(KolkataLoc)
	dk := DayKind(date)

	var best *Schedule
	var bestStart time.Time

	for i := range schedules {
		s := &schedules[i]
		if s.DayKind != dk {
			continue
		}
		start, _ := MealWindow(*s, date)
		if start.After(now) {
			if best == nil || start.Before(bestStart) {
				best = s
				bestStart = start
			}
		}
	}
	return best
}

// --- Review state ---

// ReviewState returns one of: "not_yet_open", "open", "closed", "already_reviewed".
func ReviewState(mealStart, mealEnd time.Time, hasReviewed bool, now time.Time) string {
	if hasReviewed {
		return "already_reviewed"
	}
	if now.Before(mealStart) {
		return "not_yet_open"
	}
	if now.After(mealEnd) {
		return "closed"
	}
	return "open"
}

// --- Crowd state ---

// CrowdState describes the display state of a mess's crowd level.
type CrowdState struct {
	State       string  // "fresh", "stale", "no_update", "closed"
	Level       string  // "low", "medium", "high" or ""
	UpdatedAt   *time.Time
	AgeMinutes  *int
	NextOpening *time.Time
}

// ComputeCrowdState calculates the crowd display state for a single mess.
// lastUpdate may be nil if there's no crowd_status row since the meal started.
// mealStart/mealEnd are for the current meal window (zero values if no meal is active).
func ComputeCrowdState(
	now time.Time,
	mealStart, mealEnd time.Time,
	lastUpdate *time.Time,
	lastLevel string,
	staleMinutes int,
	nextOpening *time.Time,
) CrowdState {
	insideMeal := !mealStart.IsZero() && !mealEnd.IsZero() &&
		!now.Before(mealStart) && !now.After(mealEnd)

	if !insideMeal {
		cs := CrowdState{State: "closed"}
		if nextOpening != nil {
			cs.NextOpening = nextOpening
		}
		return cs
	}

	// Inside a meal window.
	if lastUpdate == nil || lastUpdate.Before(mealStart) {
		return CrowdState{State: "no_update"}
	}

	age := int(now.Sub(*lastUpdate).Minutes())
	cs := CrowdState{
		Level:      lastLevel,
		UpdatedAt:  lastUpdate,
		AgeMinutes: &age,
	}
	if age > staleMinutes {
		cs.State = "stale"
	} else {
		cs.State = "fresh"
	}
	return cs
}

// NextOpening returns when the next meal window opens: the next meal today,
// or, if none is left, tomorrow's first meal start. Returns nil when no
// schedules exist for today or tomorrow.
func NextOpening(schedules []Schedule, now time.Time) *time.Time {
	if s := NextMeal(schedules, now); s != nil {
		start, _ := MealWindow(*s, now)
		return &start
	}
	tomorrow := now.In(KolkataLoc).AddDate(0, 0, 1)
	dayStart := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, KolkataLoc)
	if s := NextMeal(schedules, dayStart); s != nil {
		start, _ := MealWindow(*s, dayStart)
		return &start
	}
	return nil
}
