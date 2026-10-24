// A fake HTTP server for component tests: it intercepts window.fetch and
// answers the same routes/contract as the real Go API (ARCHITECTURE.md
// section 7). State is in-memory per fixture; behavior is exercised through
// the same fetch wrappers the real client uses.

import { vi } from "vitest";

const BASE = "http://localhost:8080/api";

// ---------- fixtures ----------

// lunch widget meal ids must be consistent per group across state calls.
const MEALS = {
  boys: {
    lunch: "meal-boys-lunch",
    snacks: "meal-boys-snacks",
    breakfast: "meal-boys-breakfast",
    dinner: "meal-boys-dinner",
  },
  girls: {
    lunch: "meal-girls-lunch",
    breakfast: "meal-girls-breakfast",
    snacks: "meal-girls-snacks",
    dinner: "meal-girls-dinner",
  },
};

export const user = (over = {}) => ({
  id: "user-1",
  email: "112400000@cse.iiitp.ac.in",
  role: "student",
  hostelGroup: "boys",
  dietPref: "regular",
  cvdMode: "off",
  textScale: 1,
  theme: "system",
  reducedMotion: false,
  ...over,
});

const SCHEDULES = {
  breakfast: { weekday: { start: "08:00", end: "09:45" }, weekend: { start: "08:00", end: "10:15" } },
  lunch: { weekday: { start: "12:30", end: "14:30" }, weekend: { start: "12:30", end: "14:30" } },
  snacks: { weekday: { start: "17:00", end: "18:00" }, weekend: { start: "17:00", end: "18:00" } },
  dinner: { weekday: { start: "19:30", end: "21:30" }, weekend: { start: "19:30", end: "21:30" } },
};

export function messes(group = "boys", crowdByCode = {}, now = new Date("2026-10-20T13:00:00")) {
  const all = group === "boys"
    ? [
        { code: "krishna", name: "Krishna" },
        { code: "brahmaputra", name: "Brahmaputra" },
        { code: "godavri", name: "Godavri" },
      ]
    : [
        { code: "sindhu", name: "Sindhu" },
        { code: "indrayani", name: "Indrayani" },
      ];
  return all.map((m, i) => {
    const crowd = crowdByCode[m.code] || null; // {state, level, ageMinutes}
    return {
      id: `mess-${m.code}`,
      code: m.code,
      name: m.name,
      crowd: crowdFixture(crowd, now),
    };
  });
}

function crowdFixture(spec, now) {
  if (!spec) return { state: "no_update", level: null, updatedAt: null, ageMinutes: null, nextOpening: null, message: null };
  if (spec.state === "closed")
    return {
      state: "closed",
      level: null,
      updatedAt: null,
      ageMinutes: null,
      nextOpening: new Date("2026-10-20T12:30:00").toISOString(),
      message: "Mess closed. Opens at 12:30",
    };
  const age = spec.ageMinutes ?? 5;
  return {
    state: spec.state,
    level: spec.level || null,
    updatedAt: new Date(now.getTime() - age * 60000).toISOString(),
    ageMinutes: age,
    nextOpening: null,
    // The server writes this sentence; the client renders it verbatim.
    message:
      spec.state === "stale"
        ? `Updated ${age} min ago, may be outdated`
        : `Updated ${age} min ago`,
  };
}

export function meals(group = "boys") {
  const windowLabel = (type) => `${SCHEDULES[type].weekday.start}–${SCHEDULES[type].weekday.end}`;
  return ["lunch", "snacks"].map((type, i) => ({
    id: MEALS[group][type],
    mealType: type,
    status: "published",
    serviceDate: "2026-10-20",
    startsAt: new Date(`2026-10-20T${SCHEDULES[type].weekday.start}:00+05:30`).toISOString(),
    endsAt: new Date(`2026-10-20T${SCHEDULES[type].weekday.end}:00+05:30`).toISOString(),
    items:
      i === 0
        ? [
            { id: "i1", name: "SAMPLE Rice", dietMenu: "both", sortOrder: 0 },
            { id: "i2", name: "SAMPLE Jain Sabzi", dietMenu: "jain", sortOrder: 1 },
          ]
        : [{ id: "i3", name: "SAMPLE Samosa", dietMenu: "both", sortOrder: 0 }],
    ratings: {
      regular: { avg: i === 0 ? 4.2 : null, count: i === 0 ? 31 : 0 },
      jain: { avg: null, count: 0 },
    },
    hasReviewed: false,
    reviewState: "open",
    scheduleLabel: windowLabel(type),
  }));
}

// ---------- fake server ----------

function json(res, status, body) {
  return new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });
}

export function stub({ me, messes: messesFix, meals: mealsFix, managerMeals, defaultItems = [], reviews = [], now = new Date("2026-10-20T13:00:00") } = {}) {
  const state = {
    reviews: [...reviews],
    calls: [],
    myReviews: {}, // mealId -> review
    crowdSaved: [],
    menuSaved: [],
    published: [],
    moderated: [],
  };

  const handler = async (url, method, body) => {
    // Base-independent: take the path after the "/api" prefix, whatever the base is.
    const idx = url.indexOf("/api/");
    const path = (idx >= 0 ? url.slice(idx + 4) : url).split("?")[0];
    const query = new URLSearchParams(url.split("?")[1] || "");

    if (method === "GET" && path === "/me") return { status: 200, body: { user: me || user() } };
    if (method === "POST" && path === "/auth/logout") return { status: 204, body: {} };
    if (method === "GET" && path === "/messes") return { status: 200, body: { messes: typeof messesFix === "function" ? messesFix() : messesFix || [], serverTime: now } };
    if (method === "GET" && path === "/meals") return mealsFix != null ? { status: 200, body: { meals: typeof mealsFix === "function" ? mealsFix(state) : mealsFix, serverTime: now, date: "2026-10-20" } } : { status: 200, body: { meals: meals(state), serverTime: now, date: "2026-10-20" } };
    if (method === "POST" && /^\/meals\/[^/]+\/reviews$/.test(path)) {
      const mealId = path.split("/")[2];
      if (state.myReviews[mealId]) return { status: 409, body: { error: { code: "REVIEW_ALREADY_SUBMITTED", message: "You already reviewed this meal." } } };
      const review = { id: "rev-1", mealId, variant: body.variant, stars: body.stars, comment: body.comment || "", createdAt: now.toISOString() };
      state.myReviews[mealId] = review;
      state.reviews.push({ ...review, userEmail: me?.email });
      return { status: 201, body: { review } };
    }
    if (method === "PUT" && /^\/manager\/messes\/[^/]+\/crowd$/.test(path)) {
      state.crowdSaved.push({ messId: path.split("/")[3], level: body.level });
      return { status: 200, body: { crowd: { level: body.level } } };
    }
    if (method === "PUT" && /^\/manager\/meals\/\d{4}-\d{2}-\d{2}\/[^/]+$/.test(path)) {
      state.menuSaved.push({ url: path, items: body.items });
      state.editorMeals = state.editorMeals || {};
      state.editorMeals[path] = body.items;
      return { status: 200, body: { meal: { mealType: path.split("/")[3], status: "draft" } } };
    }
    if (method === "POST" && /\/publish$/.test(path)) {
      state.published.push(path);
      return { status: 200, body: { meal: { mealType: "lunch", status: "published" } } };
    }
    if (method === "GET" && path === "/manager/reviews") return { status: 200, body: { reviews: state.reviews, meal: {} } };
    if (method === "POST" && /^\/manager\/reviews\/[^/]+\/(hide|unhide)$/.test(path)) {
      const id = path.split("/")[3];
      const hide = path.endsWith("/hide");
      state.moderated.push({ id, hide });
      const rv = state.reviews.find((r) => r.id === id);
      if (rv) rv.hidden = hide;
      return { status: 200, body: { review: { id, hidden: hide } } };
    }
    if (method === "PATCH" && path === "/me/preferences") {
      return { status: 200, body: { user: { ...(me || user()), ...body } } };
    }
    if (method === "GET" && path === "/manager/meals") return { status: 200, body: { meals: typeof managerMeals === "function" ? managerMeals(state) : managerMeals || [], date: query.get("date") || "2026-10-20", serverTime: now } };
    if (method === "GET" && path === "/manager/default-items") return { status: 200, body: { items: typeof defaultItems === "function" ? defaultItems() : defaultItems, serverTime: now } };
    return { status: 404, body: { error: { code: "NOT_FOUND", message: "Unknown route (test server)." } } };
  };

  const fetchImpl = vi.fn(async (input, init = {}) => {
    const url = typeof input === "string" ? input : input.url;
    const method = (init.method || (typeof input === "string" ? "GET" : input.method) || "GET").toUpperCase();
    state.calls.push({ url, method, body: init.body ? String(init.body) : undefined });
    let parsed;
    if (init.body && typeof init.body === "string") {
      try {
        parsed = JSON.parse(init.body);
      } catch {
        parsed = undefined;
      }
    }
  const out = await handler(url, method, parsed);
  const text = out.status === 204 ? "" : JSON.stringify(out.body ?? {});
  return new Response(text, {
    status: out.status,
    headers: { "Content-Type": "application/json" },
  });
  });

  vi.stubGlobal("fetch", fetchImpl);

  return state;
}
