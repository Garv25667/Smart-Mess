import React from "react";
import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, cleanup } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import axe from "axe-core";
import { MemoryRouter } from "react-router-dom";
import App from "./App.jsx";
import { AuthProvider } from "./auth/AuthContext.jsx";
import { stub, messes, meals as mealsFix, user } from "./test/apiMock.js";

let restore;
function makeClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0, staleTime: 0 } },
  });
}

function mount(initial = "/") {
  const qc = makeClient();
  return render(
    <QueryClientProvider client={qc}>
      <AuthProvider>
        <MemoryRouter initialEntries={[initial]}>
          <App />
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>
  );
}

const serious = (res) => res.violations.filter((v) => ["serious", "critical"].includes(v.impact));

describe("Student home", () => {
  beforeEach(() => {
    document.documentElement.lang = "en";
    try {
      sessionStorage.clear();
    } catch {}
  });
  afterEach(() => { if (restore) restore(); cleanup(); });

  it("renders crowd cards with text label, person icons and the 3-segment bar", async () => {
    const state = stub({
      me: user(),
      messes: () =>
        messes("boys", {
          krishna: { state: "fresh", level: "low", ageMinutes: 5 },
          brahmaputra: { state: "fresh", level: "medium", ageMinutes: 10 },
          godavri: { state: "fresh", level: "high", ageMinutes: 2 },
        }),
    });
    restore = state.restore;
    mount("/");
    await screen.findByText("Krishna");
    expect(screen.getByText("Low crowd")).toBeTruthy();
    expect(screen.getByText("Medium crowd")).toBeTruthy();
    expect(screen.getByText("High crowd")).toBeTruthy();
    expect(screen.getAllByRole("img", { name: /Crowd bar/ }).length).toBe(3);
  });

  it("shows the stale note and dashed state via text", async () => {
    const state = stub({
      me: user(),
      messes: () => messes("boys", { godavri: { state: "stale", level: "medium", ageMinutes: 45 } }),
    });
    restore = state.restore;
    mount("/");
    await screen.findByText("Godavri");
    expect(screen.getByText(/may be outdated/)).toBeTruthy();
    expect(screen.getByText("Medium crowd")).toBeTruthy();
  });

  it("shows closed cards and no-update cards", async () => {
    const state = stub({
      me: user(),
      messes: () =>
        messes("boys", {
          krishna: { state: "closed" },
          brahmaputra: null, // no_update
        }),
    });
    restore = state.restore;
    mount("/");
    await screen.findByText("Krishna");
    expect(screen.getByText(/Mess closed\. Opens at/)).toBeTruthy();
    expect(screen.getAllByText("No update yet").length).toBeGreaterThan(0);
  });

  it("shows the menu with variant tabs and rating summary from the server", async () => {
    const state = stub({ me: user(), messes: () => [], meals: mealsFix("boys") });
    restore = state.restore;
    mount("/");
    await screen.findByText("SAMPLE Rice");
    expect(screen.getByText("Same in both menus")).toBeTruthy(); // 'both' item tag
    expect(screen.getByText("4.2 / 5")).toBeTruthy();
    expect(screen.getByText("(31 reviews)")).toBeTruthy();
    expect(screen.getByText("No ratings yet")).toBeTruthy(); // jain variant has no ratings
  });

  it("shows the 'not published' placeholder from server status (draft meals excluded)", async () => {
    // Server returns only published meals; if none exist show the empty note.
    const state = stub({ me: user(), messes: () => [], meals: [] });
    restore = state.restore;
    mount("/");
    await screen.findByText("No published menu for today yet.");
  });

  it("passes axe with no serious or critical violations", async () => {
    const state = stub({
      me: user(),
      messes: () =>
        messes("boys", {
          krishna: { state: "fresh", level: "high", ageMinutes: 2 },
          brahmaputra: { state: "stale", level: "low", ageMinutes: 42 },
          godavri: null,
        }),
      meals: mealsFix("boys"),
    });
    restore = state.restore;
    document.title = "Smart Mess — IIIT Pune";
    mount("/");
    await screen.findByText("Review this meal");
    await waitFor(() => {
      expect(screen.getAllByText("Crowd right now").length).toBeGreaterThan(0);
    });
    const res = await axe.run(document, { resultTypes: ["violations"] });
    const bad = serious(res);
    if (bad.length) console.error("AXE:", JSON.stringify(bad, null, 2));
    expect(bad).toEqual([]);
  });
});
