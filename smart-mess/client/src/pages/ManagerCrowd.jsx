// Sub manager crowd screen (route /manager/crowd). Two taps: pick a level,
// then Save. The server checks manager_messes; this screen only lists the
// messes the server already says this manager owns (user.messes from /api/me).

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { getMesses, setCrowd, scheduleLabel, ApiError } from "../api";
import { useAuth } from "../auth/AuthContext.jsx";
import { CheckIcon, ClockIcon, PersonIcon, WarningTriangleIcon, QuestionMarkIcon } from "../components/icons.jsx";

const LEVELS = [
  { id: "low", label: "Low", filled: 1, tone: "tone-low" },
  { id: "medium", label: "Medium", filled: 2, tone: "tone-medium" },
  { id: "high", label: "High", filled: 3, tone: "tone-high", warn: true },
];

const LEVEL_NAME = { low: "Low", medium: "Medium", high: "High" };

export default function ManagerCrowd() {
  const { user } = useAuth();
  const myMessIds = (user?.messes || []).map((m) => m.id);

  const messesQuery = useQuery({
    queryKey: ["messes"],
    queryFn: getMesses,
    refetchInterval: (q) =>
      q.state.error ? 10_000 : document.visibilityState === "visible" ? 30_000 : false,
  });
  const messes = (messesQuery.data?.messes || []).filter((m) => myMessIds.includes(m.id));

  // Multiple assigned messes: a switcher (A2 allows any number of messes).
  const [messIdx, setMessIdx] = useState(0);
  const mess = messes[Math.min(messIdx, Math.max(messes.length - 1, 0))];

  const current = mess?.crowd?.level || null;
  const [selected, setSelected] = useState(current);
  const [savedLevel, setSavedLevel] = useState(current);
  const [savedAt, setSavedAt] = useState(null);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);

  if (!user) return <p>Loading…</p>;
  if (messesQuery.isLoading) return <p role="status" className="p-4 font-bold">Loading crowd data…</p>;
  if (messes.length === 0) {
    return (
      <section aria-labelledby="crowd-mgr-h" className="mx-auto max-w-xl space-y-3">
        <h1 id="crowd-mgr-h" className="text-2xl font-bold">Update crowd level</h1>
        <p className="font-bold text-[var(--danger)]">
          No messes are assigned to your account. Ask the administrator to add
          your assignment in the database.
        </p>
      </section>
    );
  }

  function switchMess(i) {
    setMessIdx(i);
    const next = messes[i];
    setSelected(next?.crowd?.level || null);
    setError(null);
  }

  async function onSave() {
    if (!selected || !mess) return;
    setBusy(true);
    setError(null);
    try {
      await setCrowd(mess.id, selected);
      setSavedLevel(selected);
      setSavedAt(new Date().toISOString());
      await messesQuery.refetch();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save the crowd level. Try again.");
    } finally {
      setBusy(false);
    }
  }

  const crowd = mess.crowd;

  return (
    <section aria-labelledby="crowd-mgr-h" className="mx-auto max-w-xl space-y-4">
      <h1 id="crowd-mgr-h" className="text-2xl font-bold">
        Update crowd level
      </h1>

      {messes.length > 1 && (
        <label className="flex items-center gap-2">
          <span className="font-bold">Mess</span>
          <select
            value={String(messIdx)}
            onChange={(e) => switchMess(Number(e.target.value))}
            className="min-h-[44px] rounded border-2 border-[var(--border)] bg-[var(--surface)] px-2"
          >
            {messes.map((m, i) => (
              <option key={m.id ?? m.code} value={i}>
                {m.name}
              </option>
            ))}
          </select>
        </label>
      )}

      <p className="rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-3">
        <span className="block text-lg font-bold">{mess.name}</span>
        <span className="flex items-center gap-1 text-sm text-[var(--text-muted)]">
          {current ? <ClockIcon /> : <QuestionMarkIcon />}
          {crowd?.state === "no_update" || !crowd
            ? "No update yet"
            : crowd?.updatedAt
              ? `Last updated ${crowd.ageMinutes ?? 0} min ago${crowd.state === "stale" ? ", may be outdated" : ""}`
              : "No update yet"}
        </span>
        <span className="mt-1 block text-sm text-[var(--text-muted)]">
          The crowd level is a manual estimate by you, the mess manager.
        </span>
      </p>

      <fieldset>
        <legend className="font-bold">How crowded is it right now?</legend>
        <div className="mt-2 grid grid-cols-3 gap-2">
          {LEVELS.map((l) => (
            <button
              key={l.id}
              type="button"
              aria-pressed={selected === l.id}
              onClick={() => setSelected(l.id)}
              className={`flex min-h-[88px] flex-col items-center justify-center gap-1 rounded-lg border-2 px-2 text-lg font-bold ${
                selected === l.id ? "border-[var(--accent)] border-4" : "border-[var(--border)]"
              } ${l.tone}`}
            >
              <span className="flex">
                {Array.from({ length: 3 }, (_, i) => (
                  <PersonIcon key={i} filled={i < l.filled} />
                ))}
                {l.warn ? <WarningTriangleIcon /> : null}
              </span>
              {l.label}
            </button>
          ))}
        </div>
      </fieldset>

      <button
        type="button"
        disabled={!selected || busy}
        onClick={onSave}
        className="primary-action min-h-[56px] w-full rounded border-2 border-[var(--accent)] bg-[var(--surface-muted)] px-4 text-xl font-bold disabled:opacity-60"
      >
        <CheckIcon /> <span className="ml-1">{busy ? "Saving…" : "Save"}</span>
      </button>
      {!selected && (
        <p className="text-sm text-[var(--text-muted)]">Tap a level first, then tap Save.</p>
      )}
      {savedLevel && (
        <p role="status" className="flex items-center gap-2 font-bold text-[var(--ok)]">
          <CheckIcon /> Saved: {LEVEL_NAME[savedLevel] || savedLevel} crowd — just updated
        </p>
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
