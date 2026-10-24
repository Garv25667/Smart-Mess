// Verify: the 6-digit code from the college email inbox.

import { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { verifyOTP, requestOTP, ApiError } from "../api";
import { useAuth } from "../auth/AuthContext.jsx";

export default function VerifyCode() {
  const navigate = useNavigate();
  const { refresh } = useAuth();
  const [email, setEmail] = useState(() => {
    try {
      return sessionStorage.getItem("sm:loginEmail") || "";
    } catch {
      return "";
    }
  });
  const [code, setCode] = useState("");
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);
  const [resent, setResent] = useState(false);
  const errorId = "verify-error";

  async function onSubmit(e) {
    e.preventDefault();
    setError(null);
    const trimmed = code.trim();
    if (!/^\d{6}$/.test(trimmed)) {
      setError("Enter the 6-digit code from your email.");
      document.getElementById(errorId)?.focus();
      return;
    }
    if (!email) {
      setError("Enter your college email first.");
      document.getElementById(errorId)?.focus();
      return;
    }
    setBusy(true);
    try {
      await verifyOTP(email, trimmed);
      await refresh();
      try {
        sessionStorage.removeItem("sm:loginEmail");
      } catch {
        // ignore
      }
      navigate("/", { replace: true });
    } catch (err) {
      const msg = err instanceof ApiError && err.code === "OTP_RATE_LIMITED"
        ? "Too many attempts. Wait a while and request a new code."
        : err instanceof ApiError
          ? err.message
          : "Could not verify the code. Try again.";
      setError(msg);
      document.getElementById(errorId)?.focus();
    } finally {
      setBusy(false);
    }
  }

  async function resend() {
    setError(null);
    setBusy(true);
    try {
      await requestOTP(email);
      setResent(true);
    } catch {
      setError("Could not send a new code. Check your email address and try again.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section aria-labelledby="verify-h" className="mx-auto max-w-xl space-y-3">
      <h1 id="verify-h" className="text-2xl font-bold">
        Enter the code
      </h1>
      {!email ? (
        <p>
          No email in this session.{" "}
          <Link to="/login" className="underline">
            Back to sign in
          </Link>
        </p>
      ) : (
        <form onSubmit={onSubmit} className="rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-4" noValidate>
          <label htmlFor="code" className="block font-bold">
            6-digit code sent to {email}
          </label>
          <input
            id="code"
            inputMode="numeric"
            autoComplete="one-time-code"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            aria-describedby={error ? errorId : undefined}
            aria-invalid={error ? "true" : "false"}
            className="mt-1 min-h-[44px] w-40 rounded border-2 border-[var(--border)] bg-[var(--bg)] p-2 tracking-[0.3em]"
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
            {busy ? "Verifying…" : "Verify"}
          </button>
          <p className="mt-2">
            <button type="button" onClick={resend} disabled={busy || !email} className="underline disabled:opacity-60">
              Send a new code
            </button>
          </p>
          {resent && (
            <p role="status" className="font-bold">
              A new code was sent.
            </p>
          )}
        </form>
      )}
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
