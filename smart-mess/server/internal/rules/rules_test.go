package rules

import (
	"testing"
	"time"
)

func kolkata(year int, month time.Month, day, hour, min int) time.Time {
	return time.Date(year, month, day, hour, min, 0, 0, KolkataLoc)
}

// --- DayKind tests ---

func TestDayKind(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"Monday", kolkata(2026, time.October, 19, 12, 0), "weekday"},
		{"Friday", kolkata(2026, time.October, 23, 12, 0), "weekday"},
		{"Saturday", kolkata(2026, time.October, 24, 12, 0), "weekend"},
		{"Sunday", kolkata(2026, time.October, 25, 12, 0), "weekend"},
		{"Tuesday", kolkata(2026, time.October, 20, 8, 0), "weekday"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DayKind(tt.t)
			if got != tt.want {
				t.Errorf("DayKind(%v) = %q, want %q", tt.t, got, tt.want)
			}
		})
	}
}

// --- MealWindow tests ---

func TestMealWindow(t *testing.T) {
	s := Schedule{
		MealType:  "breakfast",
		DayKind:   "weekday",
		StartTime: TimeOfDay{8, 0},
		EndTime:   TimeOfDay{9, 45},
	}
	date := kolkata(2026, time.October, 19, 0, 0) // Monday
	start, end := MealWindow(s, date)

	wantStart := kolkata(2026, time.October, 19, 8, 0)
	wantEnd := kolkata(2026, time.October, 19, 9, 45)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

func TestWeekendBreakfastWindow(t *testing.T) {
	s := Schedule{
		MealType:  "breakfast",
		DayKind:   "weekend",
		StartTime: TimeOfDay{8, 0},
		EndTime:   TimeOfDay{10, 15},
	}
	date := kolkata(2026, time.October, 24, 0, 0) // Saturday
	start, end := MealWindow(s, date)

	wantStart := kolkata(2026, time.October, 24, 8, 0)
	wantEnd := kolkata(2026, time.October, 24, 10, 15)

	if !start.Equal(wantStart) {
		t.Errorf("start = %v, want %v", start, wantStart)
	}
	if !end.Equal(wantEnd) {
		t.Errorf("end = %v, want %v", end, wantEnd)
	}
}

// --- CurrentMeal / NextMeal tests ---

var testSchedules = []Schedule{
	{"boys", "breakfast", "weekday", TimeOfDay{8, 0}, TimeOfDay{9, 45}},
	{"boys", "lunch", "weekday", TimeOfDay{12, 30}, TimeOfDay{14, 30}},
	{"boys", "snacks", "weekday", TimeOfDay{17, 0}, TimeOfDay{18, 0}},
	{"boys", "dinner", "weekday", TimeOfDay{19, 30}, TimeOfDay{21, 30}},
	{"boys", "breakfast", "weekend", TimeOfDay{8, 0}, TimeOfDay{10, 15}},
	{"boys", "lunch", "weekend", TimeOfDay{12, 30}, TimeOfDay{14, 30}},
	{"boys", "snacks", "weekend", TimeOfDay{17, 0}, TimeOfDay{18, 0}},
	{"boys", "dinner", "weekend", TimeOfDay{19, 30}, TimeOfDay{21, 30}},
}

func TestCurrentMeal(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want string // expected meal_type or "" if nil
	}{
		{"during breakfast weekday", kolkata(2026, 10, 19, 8, 30), "breakfast"},
		{"exactly at breakfast start", kolkata(2026, 10, 19, 8, 0), "breakfast"},
		{"exactly at breakfast end", kolkata(2026, 10, 19, 9, 45), "breakfast"},
		{"between meals", kolkata(2026, 10, 19, 10, 0), ""},
		{"during lunch", kolkata(2026, 10, 19, 13, 0), "lunch"},
		{"during dinner", kolkata(2026, 10, 19, 20, 0), "dinner"},
		{"after all meals", kolkata(2026, 10, 19, 22, 0), ""},
		{"weekend breakfast at 10:00", kolkata(2026, 10, 24, 10, 0), "breakfast"},
		{"weekend breakfast at 10:15", kolkata(2026, 10, 24, 10, 15), "breakfast"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CurrentMeal(testSchedules, tt.now)
			if tt.want == "" {
				if got != nil {
					t.Errorf("expected nil, got %v", got.MealType)
				}
			} else {
				if got == nil {
					t.Errorf("expected %s, got nil", tt.want)
				} else if got.MealType != tt.want {
					t.Errorf("got %s, want %s", got.MealType, tt.want)
				}
			}
		})
	}
}

func TestNextMeal(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want string
	}{
		{"before breakfast", kolkata(2026, 10, 19, 7, 0), "breakfast"},
		{"between breakfast and lunch", kolkata(2026, 10, 19, 10, 0), "lunch"},
		{"between snacks and dinner", kolkata(2026, 10, 19, 18, 30), "dinner"},
		{"after all meals", kolkata(2026, 10, 19, 22, 0), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextMeal(testSchedules, tt.now)
			if tt.want == "" {
				if got != nil {
					t.Errorf("expected nil, got %v", got.MealType)
				}
			} else {
				if got == nil {
					t.Errorf("expected %s, got nil", tt.want)
				} else if got.MealType != tt.want {
					t.Errorf("got %s, want %s", got.MealType, tt.want)
				}
			}
		})
	}
}

// --- ReviewState tests ---

func TestReviewState(t *testing.T) {
	mealStart := kolkata(2026, 10, 19, 12, 30)
	mealEnd := kolkata(2026, 10, 19, 14, 30)

	tests := []struct {
		name        string
		now         time.Time
		hasReviewed bool
		want        string
	}{
		{"before window", kolkata(2026, 10, 19, 11, 0), false, "not_yet_open"},
		{"exactly at start", mealStart, false, "open"},
		{"during window", kolkata(2026, 10, 19, 13, 0), false, "open"},
		{"exactly at end", mealEnd, false, "open"},
		{"after window", kolkata(2026, 10, 19, 15, 0), false, "closed"},
		{"already reviewed", kolkata(2026, 10, 19, 13, 0), true, "already_reviewed"},
		{"already reviewed even before window", kolkata(2026, 10, 19, 11, 0), true, "already_reviewed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReviewState(mealStart, mealEnd, tt.hasReviewed, tt.now)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// --- CrowdState tests ---

func TestComputeCrowdState_Closed(t *testing.T) {
	now := kolkata(2026, 10, 19, 22, 0) // after all meals
	next := kolkata(2026, 10, 20, 8, 0)
	cs := ComputeCrowdState(now, time.Time{}, time.Time{}, nil, "", 30, &next)
	if cs.State != "closed" {
		t.Errorf("state = %q, want closed", cs.State)
	}
	if cs.NextOpening == nil || !cs.NextOpening.Equal(next) {
		t.Error("expected next opening time")
	}
}

func TestComputeCrowdState_NoUpdate(t *testing.T) {
	mealStart := kolkata(2026, 10, 19, 12, 30)
	mealEnd := kolkata(2026, 10, 19, 14, 30)
	now := kolkata(2026, 10, 19, 13, 0)
	cs := ComputeCrowdState(now, mealStart, mealEnd, nil, "", 30, nil)
	if cs.State != "no_update" {
		t.Errorf("state = %q, want no_update", cs.State)
	}
}

func TestComputeCrowdState_NoUpdate_OldUpdate(t *testing.T) {
	mealStart := kolkata(2026, 10, 19, 12, 30)
	mealEnd := kolkata(2026, 10, 19, 14, 30)
	now := kolkata(2026, 10, 19, 13, 0)
	// Last update was before this meal started
	old := kolkata(2026, 10, 19, 9, 0)
	cs := ComputeCrowdState(now, mealStart, mealEnd, &old, "low", 30, nil)
	if cs.State != "no_update" {
		t.Errorf("state = %q, want no_update", cs.State)
	}
}

func TestComputeCrowdState_Fresh(t *testing.T) {
	mealStart := kolkata(2026, 10, 19, 12, 30)
	mealEnd := kolkata(2026, 10, 19, 14, 30)
	now := kolkata(2026, 10, 19, 13, 0)
	upd := kolkata(2026, 10, 19, 12, 45) // 15 min ago
	cs := ComputeCrowdState(now, mealStart, mealEnd, &upd, "medium", 30, nil)
	if cs.State != "fresh" {
		t.Errorf("state = %q, want fresh", cs.State)
	}
	if cs.Level != "medium" {
		t.Errorf("level = %q, want medium", cs.Level)
	}
}

func TestComputeCrowdState_Stale(t *testing.T) {
	mealStart := kolkata(2026, 10, 19, 12, 30)
	mealEnd := kolkata(2026, 10, 19, 14, 30)
	now := kolkata(2026, 10, 19, 14, 0)
	upd := kolkata(2026, 10, 19, 12, 35) // 85 min ago
	cs := ComputeCrowdState(now, mealStart, mealEnd, &upd, "high", 30, nil)
	if cs.State != "stale" {
		t.Errorf("state = %q, want stale", cs.State)
	}
	if cs.Level != "high" {
		t.Errorf("level = %q, want high", cs.Level)
	}
	if cs.AgeMinutes == nil || *cs.AgeMinutes <= 30 {
		t.Errorf("ageMinutes should be > 30, got %v", cs.AgeMinutes)
	}
}

func TestNextOpening(t *testing.T) {
	// Between meals: opens at lunch start.
	if got := NextOpening(testSchedules, kolkata(2026, 10, 19, 10, 0)); got == nil || !got.Equal(kolkata(2026, 10, 19, 12, 30)) {
		t.Errorf("got %v, want lunch start 12:30", got)
	}
	// After all meals: opens at tomorrow's breakfast start.
	if got := NextOpening(testSchedules, kolkata(2026, 10, 19, 22, 0)); got == nil || !got.Equal(kolkata(2026, 10, 20, 8, 0)) {
		t.Errorf("got %v, want tomorrow 08:00", got)
	}
	// Weekend evening: next day is Monday (weekday) breakfast.
	if got := NextOpening(testSchedules, kolkata(2026, 10, 24, 22, 0)); got == nil || !got.Equal(kolkata(2026, 10, 25, 8, 0)) {
		t.Errorf("got %v, want Sunday 08:00", got)
	}
}
