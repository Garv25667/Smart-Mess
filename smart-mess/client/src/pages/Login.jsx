// Login: college email + one-time code. Same response whether or not the email
// exists on the server (the server enforces that; this screen just relays it).

import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { requestOTP, ApiError } from "../api";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [error, setError] = useState(null);
  const [sent, setSent] = useState(false);
  const [busy, setBusy] = useState(false);
  const errorId = "login-error";

  async function onSubmit(e) {
    e.preventDefault();
    setError(null);
    if (!EMAIL_RE.test(email.trim())) {
      setError("Enter your college email, for example 112400000@cse.iiitp.ac.in.");
      document.getElementById(errorId)?.focus();
      return;
    }
    setBusy(true);
    try {
      await requestOTP(email.trim());
      try {
        sessionStorage.setItem("sm:loginEmail", email.trim());
      } catch {
        // ignore: the code screen then uses the field value for the resend path
      }
      navigate("/verify");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not send the code. Try again.");
      document.getElementById(errorId)?.focus();
    } finally {
      setBusy(false);
    }
  }

  return (
    <section aria-labelledby="login-h" className="mx-auto max-w-xl space-y-3">
      <h1 id="login-h" className="text-2xl font-bold">
        Sign in
      </h1>
      <p className="text-sm text-[var(--text-muted)]">
        Use your IIIT Pune college email. We send a 6-digit one-time code, valid
        for 10 minutes.
      </p>
      {sent && (
        <p role="status" className="rounded-md border-2 border-[var(--accent)] p-3 font-bold">
          Code sent. Check your inbox.
        </p>
      )}
      <form onSubmit={onSubmit} className="rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-4" noValidate>
        <label htmlFor="email" className="block font-bold">
          College email
        </label>
        <input
          id="email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          aria-describedby={error ? errorId : undefined}
          aria-invalid={error ? "true" : "false"}
          className="mt-1 min-h-[44px] w-full rounded border-2 border-[var(--border)] bg-[var(--bg)] p-2"
        />
        <p id={errorId} role="alert" tabIndex={-1} className="outline-none">
          {error ? (
            <span className="mt-1 flex items-center gap-2 font-bold text-[var(--danger)]">
              <WarningTriangle /> {error}
            </span>
          ) : null}
        </p>
        <button
          type="submit"
          disabled={busy}
          className="mt-3 min-h-[48px] w-full rounded border-2 border-[var(--accent)] bg-[var(--surface-muted)] px-4 text-lg font-bold disabled:opacity-60"
        >
          {busy ? "Sending…" : "Send code"}
        </button>
      </form>
    </section>
  );
}

function WarningTriangle() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" focusable="false">
      <path d="M12 3.5 22 20H2L12 3.5Z" />
      <path d="M12 9.5v4.5" stroke="#ffffff" strokeWidth="1.8" />
      <circle cx="12" cy="17" r="1.1" fill="#ffffff" />
    </svg>
  );
}
