// Pure business rules. Every function takes `now` and plain data.
// No React, no storage, no device clock access.

export const MEALS_BY_ORDER = ["breakfast", "lunch", "snacks", "dinner"];
export const STALE_AFTER_MINUTES = 30;

export function dayKind(date) {
  const dow = new Date(date).getDay();
  return dow === 0 || dow === 6 ? "weekend" : "weekday";
}

function parseHM(hm) {
  const [h, m] = hm.split(":").map(Number);
  return h * 60 + m;
}

// Minutes since midnight.
export function minutesOfDay(now) {
  return now.getHours() * 60 + now.getMinutes();
}

// MealWindow(fullSchedules, mealType, kind) -> {start, end} in minutes.
export function mealWindow(schedules, mealType, kind) {
  const s = schedules[mealType][kind];
  return { start: parseHM(s.start), end: parseHM(s.end) };
}

// The meal whose window contains now (start <= now <= end), else null.
export function currentMeal(schedules, now, date) {
  const kind = dayKind(date || now);
  const mins = minutesOfDay(now);
  for (const type of MEALS_BY_ORDER) {
    const w = mealWindow(schedules, type, kind);
    if (mins >= w.start && mins <= w.end) return type;
  }
  return null;
}

// Next meal strictly after now; null if none left today.
export function nextMeal(schedules, now, date) {
  const kind = dayKind(date || now);
  const mins = minutesOfDay(now);
  for (const type of MEALS_BY_ORDER) {
    const w = mealWindow(schedules, type, kind);
    if (w.start > mins) return type;
  }
  return null;
}

// "HH:MM" of the next opening (next meal today, else tomorrow's first meal).
export function nextOpening(schedules, now, date) {
  const kind = dayKind(date || now);
  const mins = minutesOfDay(now);
  for (const type of MEALS_BY_ORDER) {
    const w = mealWindow(schedules, type, kind);
    if (w.start > mins) return schedules[type][kind].start;
  }
  const tomorrow = new Date(date || now);
  tomorrow.setDate(tomorrow.getDate() + 1);
  const tomKind = dayKind(tomorrow);
  const first = MEALS_BY_ORDER[0];
  return schedules[first][tomKind].start;
}

export function isWindowOpen(window, now) {
  const mins = minutesOfDay(now);
  return mins >= window.start && mins <= window.end;
}

// reviewState -> "not_yet_open" | "open" | "closed" | "already_reviewed".
// Already reviewed wins over the window state (the student cannot review twice).
export function reviewState(window, hasReviewed, now) {
  if (hasReviewed) return "already_reviewed";
  const mins = minutesOfDay(now);
  if (mins < window.start) return "not_yet_open";
  if (mins <= window.end) return "open";
  return "closed";
}

// Crowding state for one mess.
// schedules: full { weekday, weekend } schedules. meal: active meal type or null.
// crowd: { level, updatedAt (Date or null) } | null
// Returns { state, message, level, ageMinutes, nextOpening }.
export function crowdState({ crowd, schedules, meal, now, date }) {
  const kind = dayKind(date || now);
  const mins = minutesOfDay(now);
  const activeWindow = meal ? mealWindow(schedules, meal, kind) : null;
  if (!activeWindow || mins < activeWindow.start || mins > activeWindow.end) {
    const opensAt = nextOpening(schedules, now, date);
    return { state: "closed", message: `Mess closed. Opens at ${opensAt}`, level: null, nextOpening: opensAt };
  }
  const updated = crowd && crowd.updatedAt ? new Date(crowd.updatedAt) : null;
  if (!crowd || !crowd.level || !updated || isNaN(updated.getTime())) {
    return { state: "no_update", message: "No update yet", level: null, nextOpening: null };
  }
  const ageMinutes = Math.max(0, minutesOfDay(now) - (updated.getHours() * 60 + updated.getMinutes()));
  if (ageMinutes > STALE_AFTER_MINUTES) {
    return {
      state: "stale",
      message: `Updated ${ageMinutes} min ago, may be outdated`,
      level: crowd.level,
      ageMinutes,
      nextOpening: null,
    };
  }
  return {
    state: "fresh",
    message: `Updated ${Math.max(1, ageMinutes)} min ago`,
    level: crowd.level,
    ageMinutes,
    nextOpening: null,
  };
}
