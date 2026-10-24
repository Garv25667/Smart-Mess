// Auth init bug prevention: GET /api/me answered 401 must resolve to the
// login flow (never a permanent "Loading…"), server/network failures must be
// treated as unauthenticated (never silently "logged in"), and a valid
// session must land where the role allows.
import React from "react";
import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import App from "./App.jsx";
import { AuthProvider } from "./auth/AuthContext.jsx";
import { user } from "./test/apiMock.js";

function mount(initial = "/") {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
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

function useFetch(handler) {
  const prev = window.fetch;
  window.fetch = handler;
  return () => (window.fetch = prev);
}

const json = (status, body) =>
  new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } });

describe("auth initialization", () => {
  afterEach(() => cleanup());

  it("fresh student (no hostel group yet) -> /choose-group renders the chooser, never blank", async () => {
    const restore = useFetch(async (input) => {
      const url = String(input);
      const idx = url.indexOf("/api/");
      const path = idx >= 0 ? url.slice(idx + 4) : url;
      if (path === "/me") return json(200, { user: user({ hostelGroup: null }) });
      return json(200, {});
    });
    mount("/");
    expect(await screen.findByText("Which hostel group are you in?")).toBeTruthy();
    expect(screen.getByRole("button", { name: /Girls/ })).toBeTruthy();
    expect(screen.getByRole("button", { name: /Boys/ })).toBeTruthy();
    restore();
  });

  it("no session: /api/me 401 -> loading finishes, login page shows", async () => {
    const restore = useFetch(async () =>
      json(401, { error: { code: "UNAUTHENTICATED", message: "Sign in to continue." } })
    );
    mount("/review");
    expect(await screen.findByText("Sign in")).toBeTruthy();
    restore();
  });

  it("server unreachable: fetch rejects -> login page, not stuck and not logged in", async () => {
    const restore = useFetch(async () => {
      throw new TypeError("Failed to fetch");
    });
    mount("/");
    expect(await screen.findByText("Sign in")).toBeTruthy();
    restore();
  });

  it("server error on other calls after a good /me: user stays logged in, errors surface as their own states", async () => {
    const restore = useFetch(async (input, init = {}) => {
      const url = String(input);
      const idx = url.indexOf("/api/");
      const path = idx >= 0 ? url.slice(idx + 4) : url;
      if (path.startsWith("/me") && !path.includes("preferences")) {
        return json(200, { user: user() });
      }
      return json(500, { error: { code: "INTERNAL", message: "boom" } });
    });
    mount("/");
    expect(await screen.findByText("Today's menu and crowd")).toBeTruthy();
    restore();
  });

  it("valid session: /api/me 200 -> user lands on the student home", async () => {
    const restore = useFetch(async (input) => {
      const url = String(input);
      const idx = url.indexOf("/api/");
      const path = idx >= 0 ? url.slice(idx + 4) : url;
      if (path.startsWith("/me") && !path.includes("preferences")) return json(200, { user: user() });
      if (path.startsWith("meals")) return json(200, { meals: [], date: "", serverTime: "" });
      if (path.startsWith("messes")) return json(200, { messes: [], serverTime: "" });
      return json(200, {});
    });
    mount("/");
    expect(await screen.findByText("Today's menu and crowd")).toBeTruthy();
    restore();
  });
});
