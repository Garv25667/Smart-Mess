// Review screen (route /review?meal=lunch). All four review states come from
// the server's reviewState field; the form posts to the server, which enforces
// the one-review-per-meal and window rules.

import { useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getMeals, submitReview, scheduleLabel, ApiError } from "../api";
import { useAuth } from "../auth/AuthContext.jsx";
import { StarInput, StarRow } from "../components/Ratings.jsx";
import {
  ClockIcon,
  LockIcon,
  CheckIcon,
  CrossIcon,
  WarningTriangleIcon,
  PlateIcon,
  LeafIcon,
} from "../components/icons.jsx";

const cap = (s) => s.charAt(0).toUpperCase() + s.slice(1);

export default function ReviewScreen() {
  const [params, setParams] = useSearchParams();
  const { user } = useAuth();
  const dietPref = user?.dietPref || "regular";

  const mealsQuery = useQuery({
    queryKey: ["meals"],
    queryFn: () => getMeals(),
  });
  const meals = mealsQuery.data?.meals || [];

  const paramMeal = params.get("meal");
  const [mealType, setMealType] = useState(null);
  const chosenType = mealType || (paramMeal && meals.find((m) => m.mealType === paramMeal)?.mealType);

  function pick(e) {
    const next = e.target.value;
    setMealType(next);
    setParams({ meal: next }, { replace: true });
  }

  const meal = meals.find((m) => m.mealType === chosenType) || meals[0];

  return (
    <section aria-labelledby="review-h" className="mx-auto max-w-xl space-y-3">
      <h1 id="review-h" className="text-2xl font-bold">
        Review a meal
      </h1>
      {mealsQuery.isLoading ? (
        <p role="status" className="font-bold">
          Loading the menu…
        </p>
      ) : mealsQuery.isError ? (
        <p role="alert" className="font-bold text-[var(--danger)]">
          Could not load the menu. <Link to="/" className="underline">Back to home</Link>
        </p>
      ) : (
        <>
          <fieldset>
            <legend className="font-bold">Which meal?</legend>
            {meals.map((m) => (
              <label key={m.mealType} className="flex min-h-[44px] items-center gap-2">
                <input
                  type="radio"
                  name="mealPick"
                  value={m.mealType}
                  checked={chosenType === m.mealType}
                  onChange={pick}
                  className="h-5 w-5"
                />
                <span className="font-bold">{cap(m.mealType)}</span>
                <span className="text-sm text-[var(--text-muted)]">{scheduleLabel(m)}</span>
              </label>
            ))}
          </fieldset>
          {meal ? <ReviewPanel meal={meal} dietPref={dietPref} /> : <p>No published meals yet.</p>}
        </>
      )}
    </section>
  );
}

function ReviewForm({ meal, dietPref, onSubmitted }) {
  const [variant, setVariant] = useState(dietPref);
  const [stars, setStars] = useState(0);
  const [comment, setComment] = useState("");
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);
  const [mine, setMine] = useState(null);
  const errorId = "review-error";
  const commentLen = comment.trim().length;

  async function onSubmit(e) {
    e.preventDefault();
    setError(null);
    if (!stars) {
      setError("Choose a star rating from 1 to 5.");
      document.getElementById(errorId)?.focus();
      return;
    }
    if (commentLen > 500) {
      setError("Comment must be 500 characters or fewer.");
      document.getElementById(errorId)?.focus();
      return;
    }
    setBusy(true);
    try {
      const body = { variant, stars };
      if (comment.trim()) body.comment = comment.trim();
      const data = await submitReview(meal.id, body);
      setMine(data.review || { variant, stars });
      onSubmitted();
    } catch (err) {
      if (err instanceof ApiError && err.code === "REVIEW_ALREADY_SUBMITTED") {
        setMine({ variant, stars: stars ?? null });
        onSubmitted();
        return;
      }
      setError(err instanceof ApiError ? err.message : "Could not submit. Try again.");
      document.getElementById(errorId)?.focus();
    } finally {
      setBusy(false);
    }
  }

  if (mine) {
    return (
      <div role="status" className="rounded-md border-2 border-[var(--accent)] bg-[var(--surface)] p-4">
        <p className="flex items-center gap-2 font-bold">
          <CheckIcon /> You reviewed this meal
        </p>
        <p className="mt-1 flex items-center gap-2">
          <StarRow stars={mine.stars} />
          {mine.variant === "jain" ? <LeafIcon /> : <PlateIcon />} {cap(mine.variant)} variant
        </p>
        <Link to="/" className="mt-3 inline-flex min-h-[44px] items-center rounded border-2 border-[var(--border)] bg-[var(--surface)] px-4 font-bold">
          Back to home
        </Link>
      </div>
    );
  }

  return (
    <form onSubmit={onSubmit} className="rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-4" noValidate>
      <p className="font-bold">Rate the {cap(meal.mealType)} you ate</p>
      <fieldset className="mt-2">
        <legend className="font-bold">Which version did you eat?</legend>
        <div className="mt-1 flex gap-2">
          {[
            { id: "regular", label: "Regular", Icon: PlateIcon },
            { id: "jain", label: "Jain", Icon: LeafIcon },
          ].map(({ id, label, Icon }) => (
            <label
              key={id}
              className={`flex min-h-[44px] items-center gap-2 rounded border-2 px-3 ${
                variant === id ? "border-[var(--accent)] bg-[var(--surface-muted)]" : "border-[var(--border)] bg-[var(--surface)]"
              }`}
            >
              <input
                type="radio"
                name="variant"
                value={id}
                checked={variant === id}
                onChange={() => setVariant(id)}
                className="h-5 w-5"
              />
              <Icon />
              <span className="font-bold">{label}</span>
            </label>
          ))}
        </div>
      </fieldset>

      <StarInput value={stars} onChange={setStars} />

      <div className="mt-2">
        <label htmlFor="comment" className="block font-bold">
          Comment (optional)
        </label>
        <textarea
          id="comment"
          rows={3}
          maxLength={600}
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          aria-describedby="comment-count"
          className="mt-1 w-full rounded border-2 border-[var(--border)] bg-[var(--surface)] p-2"
        />
        <p id="comment-count" className="text-sm text-[var(--text-muted)]">
          {commentLen} / 500 characters
        </p>
        {commentLen > 500 && (
          <p role="alert" className="font-bold text-[var(--danger)]">
            Comment must be 500 characters or fewer.
          </p>
        )}
      </div>

      <p id={errorId} role="alert" tabIndex={-1} aria-live="assertive" className="outline-none">
        {error ? (
          <span className="flex items-center gap-2 font-bold text-[var(--danger)]">
            <WarningTriangleIcon /> {error} <span className="sr-only">(Review failed)</span>
          </span>
        ) : null}
      </p>

      <button
        type="submit"
        disabled={busy || commentLen > 500}
        className="mt-2 min-h-[48px] w-full rounded border-2 border-[var(--accent)] bg-[var(--surface-muted)] px-4 py-2 text-lg font-bold disabled:opacity-60"
      >
        {busy ? "Submitting…" : "Submit review"}
      </button>
    </form>
  );
}

const STATES = {
  not_yet_open: {
    icon: ClockIcon,
    title: (meal) => `Reviews open at ${scheduleLabel(meal).split("–")[0]}`,
    body: "The review window opens when the meal starts being served.",
  },
  open: {
    icon: null,
    title: () => "Reviews are open for this meal",
    body: null,
  },
  closed: {
    icon: LockIcon,
    title: () => "Reviews for this meal have closed",
    body: "Reviews are only possible while the meal is being served.",
  },
  already_reviewed: {
    icon: CheckIcon,
    title: () => "You already reviewed this meal",
    body: "Only one review per student per meal is allowed.",
  },
};

export function ReviewPanel({ meal, dietPref }) {
  const info = STATES[meal.reviewState];
  if (!info) return null;
  const Icon = info.icon;
  const queryClient = useQueryClient();

  return (
    <div className="space-y-3">
      <div role="status" aria-live="polite" className="rounded-md border-2 border-dashed border-[var(--border)] bg-[var(--surface)] p-3">
        <p className="flex items-center gap-2 font-bold">
          {Icon ? <Icon /> : null}
          {info.title(meal)}
        </p>
        <p className="mt-1 text-sm text-[var(--text-muted)]">
          {info.body || `Served ${scheduleLabel(meal)} · Review state: ${meal.reviewState}`}
        </p>
      </div>
      {meal.reviewState === "open" ? (
        <ReviewForm
          meal={meal}
          dietPref={dietPref}
          key={`${meal.id}-${meal.reviewState}`}
          onSubmitted={() => queryClient.invalidateQueries({ queryKey: ["meals"] })}
        />
      ) : (
        <fieldset disabled="disabled" aria-hidden="true" className="hidden">
          <legend>Form disabled ({meal.reviewState.replace("_", " ")})</legend>
        </fieldset>
      )}
      <p className="text-sm text-[var(--text-muted)]">
        <Link to="/" className="underline">
          Back to home
        </Link>
      </p>
    </div>
  );
}
