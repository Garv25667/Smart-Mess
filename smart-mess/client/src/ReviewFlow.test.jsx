import React from "react";
import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import App from "./App.jsx";
import { AuthProvider } from "./auth/AuthContext.jsx";
import { stub, meals as mealsFix, user, messes } from "./test/apiMock.js";

let restore;

function makeClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0, staleTime: 0 } },
  });
}

function mount(initial = "/review") {
  const qc = makeClient();
  const utils = render(
    <QueryClientProvider client={qc}>
      <AuthProvider>
        <MemoryRouter initialEntries={[initial]}>
          <App />
        </MemoryRouter>
      </AuthProvider>
    </QueryClientProvider>
  );
  return utils;
}

function openState(meals) {
  return {
    me: user(),
    messes: () => [],
    // review window: still simulated by reviewState — the server computes it;
    // here the fixture just reflects what the server would return.
    meals: (state) => {
      const list = mealsFix("boys");
      const has = (mealId) => Boolean(state.myReviews[mealId]);
      return list.map((m) => ({
        ...m,
        hasReviewed: has(m.id),
        reviewState: has(m.id) ? "already_reviewed" : m.reviewState,
      }));
    },
  };
}

function closedState() {
  const list = mealsFix("boys").map((m, i) =>
    i === 0
      ? { ...m, reviewState: "closed", hasReviewed: false }
      : { ...m, reviewState: "not_yet_open", hasReviewed: false }
  );
  return { me: user(), messes: () => [], meals: list };
}

function notOpenState() {
  const lunch = mealsFix("boys")[0];
  return {
    me: user(),
    messes: () => [],
    meals: [{ ...lunch, reviewState: "not_yet_open", hasReviewed: false }],
  };
}

describe("Review screen states (server-driven)", () => {
  beforeEach(() => {
    try {
      sessionStorage.clear();
    } catch {}
  });
  afterEach(() => { if (restore) restore(); cleanup(); });

  it("open: submit a review and the meals refetch shows already reviewed", async () => {
    restore = stub(openState()).restore;
    mount("/review");
    // pick stars 4
    const radios = await screen.findAllByRole("radio", { name: "4" });
    fireEvent.click(radios[radios.length - 1]);
    fireEvent.click(screen.getByRole("button", { name: "Submit review" }));
    expect(await screen.findByText("You reviewed this meal")).toBeTruthy();
    // react-query invalidation refetches /api/meals; now the meal state is already_reviewed
    await waitFor(() => {
      expect(screen.queryByRole("button", { name: "Submit review" })).toBeNull();
    });
  });

  it("already reviewed: no form (one review per meal, from the server)", async () => {
    const st = stub(openState());
    st.myReviews = { "meal-boys-lunch": { id: "rev-0", variant: "jain", stars: 3 } };
    restore = st.restore;
    mount("/review");
    expect(await screen.findByText("You already reviewed this meal")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Submit review" })).toBeNull();
  });

  it("closed after the meal window", async () => {
    restore = stub(closedState()).restore;
    mount("/review");
    expect(await screen.findByText("Reviews for this meal have closed")).toBeTruthy();
  });

  it("not open yet before the meal window", async () => {
    restore = stub(notOpenState()).restore;
    mount("/review");
    expect(await screen.findByText(/Reviews open at/)).toBeTruthy();
  });

  it("validation: submitting without stars shows an error and sends no review POST", async () => {
    const st = stub(openState());
    restore = st.restore;
    mount("/review");
    fireEvent.click(await screen.findByRole("button", { name: "Submit review" }));
    expect(await screen.findByText("Choose a star rating from 1 to 5.")).toBeTruthy();
    const reviewPosts = st.calls.filter((c) => c.url.includes("reviews") && c.method === "POST");
    expect(reviewPosts.length).toBe(0);
  });

  it("comment over 500 characters is rejected client-side before any POST", async () => {
    const st = stub(openState());
    restore = st.restore;
    mount("/review");
    const star = await screen.findByRole("radio", { name: "5" });
    fireEvent.click(star);
    const box = await screen.findByLabelText("Comment (optional)");
    fireEvent.change(box, { target: { value: "x".repeat(501) } });
    expect(st.calls.filter((c) => c.url.includes("reviews") && c.method === "POST").length).toBe(0);
    expect(await screen.findByText("Comment must be 500 characters or fewer.")).toBeTruthy();
  });
});

describe("Manager screens", () => {
  beforeEach(() => {
    try {
      sessionStorage.clear();
    } catch {}
  });
  afterEach(() => { if (restore) restore(); cleanup(); });

  it("sub manager: two taps — pick High then Save — sends PUT and confirms", async () => {
    const me = user({ role: "mess_manager", messes: [{ id: "mess-krishna", code: "krishna", name: "Krishna" }] });
    const st = stub({ me, messes: () => messes("boys", { krishna: { state: "fresh", level: "low", ageMinutes: 8 } }) });
    restore = st.restore;
    mount("/manager/crowd");
    await screen.findByText("Update crowd level");
    fireEvent.click(screen.getByRole("button", { name: /^High/ }));
    fireEvent.click(screen.getByRole("button", { name: /Save/ }));
    expect(await screen.findByText(/Saved: High crowd/)).toBeTruthy();
    expect(st.crowdSaved).toEqual([{ messId: "mess-krishna", level: "high" }]);
  });

  it("main manager: Save and Publish saves the draft then publishes", async () => {
    const me = user({ role: "main_manager", hostelGroup: "boys" });
    const st = stub({
      me,
      messes: () => [],
      managerMeals: (state) => {
        // After saving+publishing, the editor refetches and shows published.
        const saved = state.menuSaved[0];
        return [
          {
            id: "meal-boys-lunch",
            mealType: "lunch",
            status: state.published.length > 0 ? "published" : "draft",
            items: saved ? saved.items.map((it, i) => ({ ...it, sortOrder: i })) : [],
          },
          { id: "meal-boys-snacks", mealType: "snacks", status: "draft", items: [] },
        ];
      },
      defaultItems: [{ name: "Aachar", sortOrder: 0 }, { name: "Salad", sortOrder: 1 }],
    });
    restore = st.restore;
    mount("/manager/menu");
    await screen.findByText("Prefill daily items");
    // seed the editor from the daily defaults, then rename dish 1
    fireEvent.click(screen.getByRole("button", { name: /Prefill daily items/ }));
    const nameField = await screen.findByLabelText("Dish 1 name");
    fireEvent.change(nameField, { target: { value: "SAMPLE Veg Biryani" } });
    fireEvent.click(screen.getByRole("button", { name: /Save and Publish/ }));
    expect(await screen.findByText(/Published the Lunch menu/)).toBeTruthy();
    // one PUT to save the items, one POST to publish
    expect(st.menuSaved.length).toBe(1);
    expect(st.menuSaved[0].items[0].name).toBe("SAMPLE Veg Biryani");
    expect(st.published.length).toBe(1);
  });

  it("moderation: hide marks hidden and calls the endpoint", async () => {
    const me = user({ role: "main_manager", hostelGroup: "boys" });
    const st = stub({
      me,
      messes: () => [],
      managerMeals: () => [
        { id: "meal-boys-lunch", mealType: "lunch", status: "published", items: [] },
      ],
    });
    st.reviews.push({
      id: "rev-1",
      variant: "regular",
      stars: 2,
      comment: "APITEST comment",
      hidden: false,
      createdAt: "2026-10-20T13:00:00Z",
      userEmail: "112400000@cse.iiitp.ac.in",
    });
    restore = st.restore;
    mount("/manager/reviews");
    await screen.findByText("Moderate reviews");
    const hideBtn = await screen.findByRole("button", { name: /Hide/ });
    fireEvent.click(hideBtn);
    await waitFor(() => {
      expect(st.moderated).toEqual([{ id: "rev-1", hide: true }]);
    });
    expect(await screen.findByText("comment hidden")).toBeTruthy();
  });

  it("student cannot see manager routes (client-side guard matches server rules)", async () => {
    const st = stub({
      me: user({ role: "student" }),
      messes: () => [],
      meals: (state) => mealsFix("boys"),
    });
    restore = st.restore;
    mount("/manager/crowd");
    await waitFor(() => {
      expect(screen.queryByText("Update crowd level")).toBeNull();
    });
  });
});
