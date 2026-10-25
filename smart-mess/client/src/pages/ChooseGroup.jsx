// One-time hostel group choice (C2). The server rejects a second set with
// HOSTEL_GROUP_ALREADY_SET; this screen never lets a chosen student return.

import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { setHostelGroup, ApiError } from "../api";
import { useAuth } from "../auth/AuthContext.jsx";
import { PlateIcon, LeafIcon, CheckIcon, CrossIcon, WarningTriangleIcon } from "../components/icons.jsx";

export default function ChooseGroup() {
  const navigate = useNavigate();
  const { refresh } = useAuth();
  const [picked, setPicked] = useState(null);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);
  const errorId = "choose-error";

  async function confirm(e) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await setHostelGroup(picked); // server 409s if already chosen
      await refresh();
      navigate("/", { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save your choice. Try again.");
      document.getElementById(errorId)?.focus();
    } finally {
      setBusy(false);
    }
  }

  return (
    <section aria-labelledby="choose-group-h" className="mx-auto max-w-xl choose-group-screen">
      <ol className="flow-progress" aria-label="Sign in progress">
        <li className="complete"><span>1</span>Sign in</li>
        <li className="complete"><span>2</span>Verify</li>
        <li className="current" aria-current="step"><span>3</span>Choose group</li>
      </ol>
      <h1 id="choose-group-h" className="text-2xl font-bold">
        Which hostel group are you in?
      </h1>
      <p className="mt-1 text-[var(--text-muted)]">
        This sets which messes you see. Choose once.
      </p>

      <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
        <button
          type="button"
          onClick={() => setPicked("girls")}
          aria-pressed={picked === "girls"}
          className={`hostel-choice flex min-h-[88px] items-center justify-center gap-3 rounded-lg border-2 px-4 text-xl font-bold ${
            picked === "girls"
              ? "border-[var(--accent)] bg-[var(--surface-muted)]"
              : "border-[var(--border)] bg-[var(--surface)]"
          }`}
        >
          <LeafIcon />
          Girls
        </button>
        <button
          type="button"
          onClick={() => setPicked("boys")}
          aria-pressed={picked === "boys"}
          className={`hostel-choice flex min-h-[88px] items-center justify-center gap-3 rounded-lg border-2 px-4 text-xl font-bold ${
            picked === "boys"
              ? "border-[var(--accent)] bg-[var(--surface-muted)]"
              : "border-[var(--border)] bg-[var(--surface)]"
          }`}
        >
          <PlateIcon />
          Boys
        </button>
      </div>

      {picked && (
        <div role="alert" className="mt-4 rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-4">
          <p className="flex items-center gap-2 font-bold">
            <CheckIcon />
            You picked {picked === "girls" ? "Girls" : "Boys"}
          </p>
          <p className="mt-1 flex items-start gap-2">
            <WarningTriangleIcon />
            This cannot be changed later.
          </p>
          <form onSubmit={confirm} className="mt-3 flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => setPicked(null)}
              className="min-h-[44px] rounded border-2 border-[var(--border)] bg-[var(--surface)] px-4 font-bold"
            >
              <CrossIcon />
              <span className="ml-1">Go back</span>
            </button>
            <button
              type="submit"
              disabled={busy}
              className="primary-action min-h-[44px] rounded border-2 border-[var(--accent)] bg-[var(--surface-muted)] px-4 font-bold disabled:opacity-60"
            >
              <CheckIcon />
              <span className="ml-1">Continue</span>
            </button>
          </form>
        </div>
      )}

      <p id={errorId} role="alert" tabIndex={-1} className="outline-none">
        {error ? (
          <span className="mt-2 flex items-center gap-2 font-bold text-[var(--danger)]">
            <WarningTriangleIcon /> {error}
          </span>
        ) : null}
      </p>
    </section>
  );
}
