// Main manager review moderation (route /manager/reviews). Pick a published
// meal, list all its reviews (including hidden), hide/unhide comments.
// Hiding removes only the comment for students; stars keep counting on the
// server.

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { getManagerMeals, getManagerReviews, moderateReview, todayStr, ApiError } from "../api";
import { CheckIcon, CrossIcon, WarningTriangleIcon } from "../components/icons.jsx";

const cap = (s) => s.charAt(0).toUpperCase() + s.slice(1);

export default function ManagerReviews() {
  const [date, setDate] = useState(todayStr());
  const [mealIdx, setMealIdx] = useState(0);

  const mealsQuery = useQuery({
    queryKey: ["managerMeals", date],
    queryFn: () => getManagerMeals(date),
    enabled: Boolean(date),
  });
  const meals = (mealsQuery.data?.meals || []).filter((m) => m.status === "published");
  const meal = meals[Math.min(mealIdx, Math.max(meals.length - 1, 0))];

  const reviewsQuery = useQuery({
    queryKey: ["managerReviews", meal?.id],
    queryFn: () => getManagerReviews(meal.id),
    enabled: Boolean(meal?.id),
  });
  const reviews = reviewsQuery.data?.reviews || [];
  const [error, setError] = useState(null);
  const [busyId, setBusyId] = useState(null);

  async function moderate(review, hide) {
    setBusyId(review.id);
    setError(null);
    try {
      await moderateReview(review.id, hide);
      await reviewsQuery.refetch();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not update the review. Try again.");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <section aria-labelledby="mod-h" className="space-y-4">
      <h1 id="mod-h" className="text-2xl font-bold">
        Moderate reviews
      </h1>
      <p className="text-sm text-[var(--text-muted)]">
        Hiding a review removes only its comment from students. Its stars keep
        counting towards the average.
      </p>

      <div className="flex flex-wrap items-center gap-2">
        <label className="flex items-center gap-2">
          <span className="font-bold">Service date</span>
          <input
            type="date"
            value={date}
            onChange={(e) => { setDate(e.target.value); setMealIdx(0); }}
            className="min-h-[44px] rounded border-2 border-[var(--border)] bg-[var(--surface)] px-2"
          />
        </label>
        {meals.length > 0 ? (
          <label className="flex items-center gap-2">
            <span className="font-bold">Meal</span>
            <select
              value={String(Math.min(mealIdx, meals.length - 1))}
              onChange={(e) => setMealIdx(Number(e.target.value))}
              className="min-h-[44px] rounded border-2 border-[var(--border)] bg-[var(--surface)] px-2"
            >
              {meals.map((m, i) => (
                <option key={m.id} value={i}>
                  {cap(m.mealType)} {m.status}
                </option>
              ))}
            </select>
          </label>
        ) : (
          <span className="font-bold">No published meals on this date.</span>
        )}
      </div>

      {meal && reviewsQuery.isLoading && <p role="status" className="font-bold">Loading reviews…</p>}
      {meal && reviewsQuery.isError && (
        <p role="alert" className="font-bold text-[var(--danger)]">Could not load reviews.</p>
      )}

      {meal && (
        <ul className="space-y-2">
          {reviews.map((rv) => (
            <li key={rv.id} className="rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="font-bold">
                    <span className="rounded bg-[var(--surface-muted)] px-2 py-0.5 text-sm">{cap(rv.variant)}</span>{" "}
                    {rv.stars} / 5
                    {rv.hidden ? (
                      <span className="ml-2 rounded bg-[var(--surface-muted)] px-2 py-0.5 text-sm">
                        comment hidden
                      </span>
                    ) : null}
                  </p>
                  <p className="text-sm text-[var(--text-muted)]">
                    {rv.userEmail || "student"} · {String(rv.createdAt).slice(0, 16).replace("T", " ")}
                  </p>
                  <p className="mt-1">{rv.comment || <em>No comment</em>}</p>
                </div>
                <button
                  type="button"
                  disabled={busyId === rv.id}
                  onClick={() => moderate(rv, !rv.hidden)}
                  className="flex min-h-[44px] items-center gap-1 rounded border-2 border-[var(--border)] bg-[var(--surface)] px-3 font-bold disabled:opacity-60"
                >
                  {rv.hidden ? <CheckIcon /> : <CrossIcon />}
                  <span className="ml-1">{rv.hidden ? "Unhide" : "Hide"}</span>
                </button>
              </div>
            </li>
          ))}
          {reviews.length === 0 && !reviewsQuery.isLoading && (
            <li className="rounded-md border-2 border-dashed border-[var(--border)] p-3 text-[var(--text-muted)]">
              No reviews for this meal yet.
            </li>
          )}
        </ul>
      )}

      <p role="alert" tabIndex={-1} className="outline-none">
        {error ? (
          <span className="flex items-center gap-2 font-bold text-[var(--danger)]">
            <WarningTriangleIcon /> {error}
          </span>
        ) : null}
      </p>
    </section>
  );
}
