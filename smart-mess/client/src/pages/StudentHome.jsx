// Student home (route /). Crowd cards poll every 30 s while the page is
// visible; the menu and review states come from the published meals endpoint.

import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { getMesses, getMeals, scheduleLabel } from "../api";
import { useAuth } from "../auth/AuthContext.jsx";
import CrowdCard from "../components/CrowdCard.jsx";
import { MealTabs, VariantTabs, VariantNote } from "../components/Tabs.jsx";
import { RatingSummary } from "../components/Ratings.jsx";
import { ClockIcon, MoonIcon, QuestionMarkIcon } from "../components/icons.jsx";

const REFRESH_MS = 30_000;

const cap = (s) => s.charAt(0).toUpperCase() + s.slice(1);

export default function StudentHome() {
  const { user } = useAuth();
  const dietPref = user?.dietPref || "regular";

  // Crowd: poll every 30 s, but only while the tab is visible.
  const messesQuery = useQuery({
    queryKey: ["messes"],
    queryFn: getMesses,
    refetchInterval: (query) => {
      if (query.state.error) return 10_000; // retry sooner on failure
      return document.visibilityState === "visible" ? REFRESH_MS : false;
    },
  });

  const mealsQuery = useQuery({
    queryKey: ["meals"],
    queryFn: () => getMeals(),
    refetchInterval: (query) =>
      query.state.error ? 10_000 : document.visibilityState === "visible" ? REFRESH_MS : false,
  });

  const messes = messesQuery.data?.messes || [];
  const meals = mealsQuery.data?.meals || [];

  const [mealType, setMealType] = useState(null);
  const selectedMeal = resolveSelected(meals, mealType, setMealType);

  return (
    <section aria-labelledby="home-h" className="space-y-6">
      <div>
        <h1 id="home-h" className="text-2xl font-bold">
          Today's menu and crowd
        </h1>
        <p className="text-sm text-[var(--text-muted)]">
          {mealsQuery.data ? `Date: ${mealsQuery.data.date}` : "Loading today's menu…"}
          {messesQuery.isFetching && <span> · Refreshing crowd…</span>}
        </p>
      </div>

      <section aria-labelledby="crowd-h">
        <h2 id="crowd-h" className="text-lg font-bold">
          Crowd right now
        </h2>
        {messesQuery.isLoading ? (
          <p role="status" className="mt-2 font-bold">
            Loading crowd levels…
          </p>
        ) : messesQuery.isError ? (
          <p role="alert" className="mt-2 font-bold text-[var(--danger)]">
            Could not load crowd levels. Retrying in a few seconds.
          </p>
        ) : (
          <ul className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-2">
            {messes.map((m) => (
              <li key={m.code ?? m.id}>
                <CrowdCard mess={m} />
              </li>
            ))}
          </ul>
        )}
      </section>

      <section aria-labelledby="menu-h">
        <h2 id="menu-h" className="text-lg font-bold">
          What is being served
        </h2>
        {mealsQuery.isLoading ? (
          <p role="status" className="mt-2 font-bold">
            Loading the menu…
          </p>
        ) : mealsQuery.isError ? (
          <p role="alert" className="mt-2 font-bold text-[var(--danger)]">
            Could not load the menu. Retrying in a few seconds.
          </p>
        ) : selectedMeal ? (
          <div className="mt-2 space-y-3">
            {/* published meals for the group only; not published yet case */}
            {selectedMeal.status !== "published" ? (
              <p className="rounded-md border-2 border-[var(--border)] bg-[var(--surface-muted)] p-3">
                The {selectedMeal.mealType} menu has not been published yet.
              </p>
            ) : (
              <MealSection meal={selectedMeal} meals={meals} mealType={mealType ?? selectedMeal.mealType} setMealType={setMealType} dietPref={dietPref} />
            )}
          </div>
        ) : (
          <p className="mt-2 text-[var(--text-muted)]">No published menu for today yet.</p>
        )}
      </section>
    </section>
  );
}

// Preselect: already chosen > current meal > next meal > first published meal.
function resolveSelected(meals, chosen, choose) {
  if (chosen && meals.find((m) => m.mealType === chosen)) return meals.find((m) => m.mealType === chosen);
  const pickFirstFallback = () => {
    if (meals.length === 0) return null;
    const current = meals.find((m) => m.reviewState === "open" || m.reviewState === "already_reviewed");
    if (current) return current;
    const first = meals[0];
    choose(first.mealType);
    return first;
  };
  return pickFirstFallback();
}

function MealSection({ meal, meals, mealType, setMealType, dietPref }) {
  const [variant, setVariant] = useState(dietPref);
  const shown = (meal.items || []).filter((it) =>
    variant === "jain" ? it.dietMenu !== "regular" : it.dietMenu !== "jain"
  );
  const openForReview = meal.reviewState === "open";

  return (
    <div id="meal-panel" role="tabpanel" aria-label={`${cap(meal.mealType)} menu`} className="space-y-3">
      <MealTabs
        meals={meals.map((m) => ({ mealType: m.mealType, scheduleLabel: scheduleLabel(m) }))}
        selected={mealType || meal.mealType}
        onSelect={setMealType}
        panelId="meal-panel"
      />
      <p className="text-sm text-[var(--text-muted)]">
        Served {scheduleLabel(meal)} · Review window: the meal's serving time
      </p>
      <VariantTabs selected={variant} onSelect={setVariant} panelId="meal-panel" />
      <VariantNote variant={variant} />

      {shown.length === 0 ? (
        <p className="rounded-md border-2 border-dashed border-[var(--border)] bg-[var(--surface)] p-3 text-[var(--text-muted)]">
          No dishes listed for this variant.
        </p>
      ) : (
        <ul className="rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-2">
          {shown.map((it) => (
            <li
              key={String(it.id) + it.name}
              className="flex flex-wrap items-center justify-between gap-2 border-b border-[var(--border)] px-2 py-2 last:border-b-0"
            >
              <span className="font-medium">{it.name}</span>
              {it.dietMenu === "both" ? (
                <span className="rounded bg-[var(--surface-muted)] px-2 py-0.5 text-xs font-bold">
                  Same in both menus
                </span>
              ) : null}
            </li>
          ))}
        </ul>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <Link
          to={`/review?meal=${meal.mealType}`}
          className={`min-h-[44px] rounded border-2 border-[var(--accent)] bg-[var(--surface-muted)] px-4 py-2 font-bold ${
            openForReview ? "" : "pointer-events-none opacity-60"
          }`}
          aria-disabled={!openForReview}
        >
          Review this meal
        </Link>
        <ReviewStateHint state={meal.reviewState} meal={meal} />
      </div>

      <RatingSummary meal={meal} />
    </div>
  );
}

function ReviewStateHint({ state, meal }) {
  if (state === "open") return null;
  if (state === "not_yet_open")
    return (
      <span className="flex items-center gap-1 text-sm text-[var(--text-muted)]">
        <ClockIcon /> Reviews open at {scheduleLabel(meal).split("–")[0]}
      </span>
    );
  if (state === "closed")
    return (
      <span className="flex items-center gap-1 text-sm text-[var(--text-muted)]">
        <MoonIcon /> Reviews for this meal have closed
      </span>
    );
  if (state === "already_reviewed")
    return (
      <span className="flex items-center gap-1 text-sm text-[var(--text-muted)]">
        <QuestionMarkIcon />
        <span>✎ You already reviewed this meal</span>
      </span>
    );
  return null;
}
