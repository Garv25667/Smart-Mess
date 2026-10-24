import { describe, it, expect } from "vitest";
import { dayKind, currentMeal, nextMeal, reviewState, crowdState, mealWindow } from "./rules.js";

// Same windows the prototype (HARNESS.md) and the server seed use.
const MEAL_SCHEDULES = {
  breakfast: { weekday: { start: "08:00", end: "09:45" }, weekend: { start: "08:00", end: "10:15" } },
  lunch: { weekday: { start: "12:30", end: "14:30" }, weekend: { start: "12:30", end: "14:30" } },
  snacks: { weekday: { start: "17:00", end: "18:00" }, weekend: { start: "17:00", end: "18:00" } },
  dinner: { weekday: { start: "19:30", end: "21:30" }, weekend: { start: "19:30", end: "21:30" } },
};

const W = (h, m) => new Date(2026, 9, 20, h, m); // Tue 2026-10-20, local tz
const SAT = (h, m) => new Date(2026, 9, 17, h, m); // Sat 2026-10-17

const lunchWindow = { start: 12 * 60 + 30, end: 14 * 60 + 30 };

describe("dayKind", () => {
  it("weekday for Tue–Fri", () => {
    expect(dayKind(W(10, 0))).toBe("weekday");
  });
  it("weekend for Sat and Sun", () => {
    expect(dayKind(SAT(10, 0))).toBe("weekend");
    expect(dayKind(new Date(2026, 9, 18, 10, 0))).toBe("weekend");
  });
});

describe("currentMeal / windows", () => {
  it("lunch open at 12:30 exactly and at 14:30 exactly", () => {
    expect(currentMeal(MEAL_SCHEDULES, W(12, 30))).toBe("lunch");
    expect(currentMeal(MEAL_SCHEDULES, W(14, 30))).toBe("lunch");
  });
  it("no meal just before breakfast or after dinner", () => {
    expect(currentMeal(MEAL_SCHEDULES, W(7, 59))).toBeNull();
    expect(currentMeal(MEAL_SCHEDULES, W(22, 0))).toBeNull();
  });
  it("not open one minute before start", () => {
    expect(currentMeal(MEAL_SCHEDULES, W(12, 29))).toBeNull();
  });
  it("weekend breakfast ends 10:15, weekday 09:45", () => {
    expect(mealWindow(MEAL_SCHEDULES, "breakfast", "weekend").end).toBe(10 * 60 + 15);
    expect(currentMeal(MEAL_SCHEDULES, SAT(9, 45))).toBe("breakfast");
    expect(currentMeal(MEAL_SCHEDULES, SAT(10, 15))).toBe("breakfast");
    expect(currentMeal(MEAL_SCHEDULES, SAT(10, 16))).toBeNull();
    expect(currentMeal(MEAL_SCHEDULES, W(9, 46))).toBeNull();
  });
  it("nextMeal picks the next meal today, else null", () => {
    expect(nextMeal(MEAL_SCHEDULES, W(11, 0))).toBe("lunch");
    expect(nextMeal(MEAL_SCHEDULES, W(15, 0))).toBe("snacks");
    expect(nextMeal(MEAL_SCHEDULES, W(23, 0))).toBeNull();
  });
});

describe("reviewState (review window)", () => {
  it("not_yet_open before the meal starts", () => {
    expect(reviewState(lunchWindow, false, W(11, 0))).toBe("not_yet_open");
  });
  it("open at start, inside, and at end", () => {
    expect(reviewState(lunchWindow, false, W(12, 30))).toBe("open");
    expect(reviewState(lunchWindow, false, W(13, 0))).toBe("open");
    expect(reviewState(lunchWindow, false, W(14, 30))).toBe("open");
  });
  it("closed after the meal ends", () => {
    expect(reviewState(lunchWindow, false, W(14, 31))).toBe("closed");
  });
  it("one review per student per meal: already reviewed wins at any time", () => {
    expect(reviewState(lunchWindow, true, W(13, 0))).toBe("already_reviewed");
    expect(reviewState(lunchWindow, true, W(15, 0))).toBe("already_reviewed");
    expect(reviewState(lunchWindow, true, W(11, 0))).toBe("already_reviewed");
  });
});

describe("crowdState", () => {
  const schedules = MEAL_SCHEDULES;
  it("closed outside every meal window, with next opening", () => {
    const r = crowdState({ crowd: null, schedules, now: W(11, 0) });
    expect(r.state).toBe("closed");
    expect(r.message).toBe("Mess closed. Opens at 12:30");
    expect(r.level).toBeNull();
  });
  it("no_update inside a window when there is no crowd row", () => {
    const r = crowdState({ crowd: null, schedules, meal: "lunch", now: W(13, 0) });
    expect(r.state).toBe("no_update");
    expect(r.message).toBe("No update yet");
  });
  it("fresh when the newest row is 30 minutes old or less", () => {
    const r = crowdState({ crowd: { level: "low", updatedAt: W(12, 55) }, schedules, meal: "lunch", now: W(13, 25) });
    expect(r.state).toBe("fresh");
    expect(r.level).toBe("low");
  });
  it("stale when the newest row is older than 30 minutes", () => {
    const r = crowdState({ crowd: { level: "high", updatedAt: W(13, 15) }, schedules, meal: "lunch", now: W(14, 0) });
    expect(r.state).toBe("stale");
    expect(r.level).toBe("high");
    expect(r.message).toContain("may be outdated");
  });
  it("crowd row from before the meal's start is stale, not fresh", () => {
    const r = crowdState({ crowd: { level: "medium", updatedAt: W(9, 0) }, schedules, meal: "lunch", now: W(13, 0) });
    expect(r.state).toBe("stale");
  });
});
