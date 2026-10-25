// Settings (route /settings): color vision mode, text size, reduced motion and
// menu preference. Every change PATCHes /api/me/preferences immediately.

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { updatePreferences, ApiError } from "../api";
import { useAuth } from "../auth/AuthContext.jsx";
import { WarningTriangleIcon } from "../components/icons.jsx";

const CVD_OPTIONS = [
  { id: "off", label: "Color vision mode off (default)" },
  { id: "unsure", label: "Not sure which type I have" },
  { id: "protan", label: "Red-weak or red-blind (protanopia / protanomaly)" },
  { id: "deutan", label: "Green-weak or green-blind (deuteranopia / deuteranomaly)" },
  { id: "tritan", label: "Blue-weak (tritanopia / tritanomaly)" },
  { id: "mono", label: "No color vision (achromatopsia)" },
];

const THEME_OPTIONS = [
  { id: "light", label: "Light" },
  { id: "dark", label: "Dark" },
  { id: "system", label: "System" },
  { id: "high_contrast", label: "High contrast" },
];

export default function Settings() {
  const { user, refresh } = useAuth();
  const [error, setError] = useState(null);
  const [saved, setSaved] = useState(false);
  const queryClient = useQueryClient();
  const pref = user || {};

  async function save(patch) {
    setError(null);
    setSaved(false);
    try {
      await updatePreferences(patch);
      await refresh();
      queryClient.invalidateQueries(); // screens re-render with the new prefs
      setSaved(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save the setting. Try again.");
    }
  }

  if (!user) {
    return (
      <section aria-labelledby="settings-h">
        <h1 id="settings-h" className="text-2xl font-bold">Settings</h1>
        <p>Loading…</p>
      </section>
    );
  }

  return (
    <section aria-labelledby="settings-h" className="mx-auto max-w-xl space-y-4 settings-screen">
      <h1 id="settings-h" className="text-2xl font-bold">
        Settings
      </h1>
      <p className="settings-intro">Customize the appearance and accessibility options.</p>

      <fieldset className="settings-group rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-3">
        <legend className="px-1 font-bold">Appearance</legend>
        <p className="text-sm text-[var(--text-muted)]">Choose how Smart Mess looks.</p>
        <div className="theme-options">
          {THEME_OPTIONS.map((option) => (
            <label key={option.id} className="theme-option">
              <input
                type="radio"
                name="theme"
                value={option.id}
                checked={(pref.theme || "system") === option.id}
                onChange={() => save({ theme: option.id })}
              />
              <span>{option.label}</span>
            </label>
          ))}
        </div>
      </fieldset>

      <fieldset className="settings-group rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-3">
        <legend className="px-1 font-bold">Do you have trouble telling colors apart?</legend>
        <p className="text-sm text-[var(--text-muted)]">
          Color vision mode changes colors only. Text, icons and patterns are
          identical in every mode.
        </p>
        <div className="mt-2 space-y-1 cvd-option-list">
          {CVD_OPTIONS.map((o) => (
            <label key={o.id} className="flex min-h-[44px] items-center gap-2">
              <input
                type="radio"
                name="cvd"
                value={o.id}
                checked={pref.cvdMode === o.id}
                onChange={() => save({ cvdMode: o.id })}
                className="h-5 w-5"
              />
              <span>{o.label}</span>
            </label>
          ))}
        </div>
        <p className="mt-3 text-sm font-bold">Sample status colors:</p>
        <ul className="mt-1 flex gap-2 text-sm">
          <li className="flex items-center gap-1 rounded px-2 py-1 tone-low">Low</li>
          <li className="flex items-center gap-1 rounded px-2 py-1 tone-medium">Medium</li>
          <li className="flex items-center gap-1 rounded px-2 py-1 tone-high">High</li>
        </ul>
        <p className="mt-1 text-xs text-[var(--text-muted)]">
          Statuses never rely on these colors alone; every status also has text,
          an icon and a pattern.
        </p>
      </fieldset>

      <fieldset className="settings-group rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-3">
        <legend className="px-1 font-bold">Text size</legend>
        <label htmlFor="text-scale" className="block">
          <span className="font-bold">Text scale: {Number(pref.textScale || 1).toFixed(2)}×</span>
          <input
            id="text-scale"
            type="range"
            min="1"
            max="2"
            step="0.05"
            value={pref.textScale || 1}
            onChange={(e) => save({ textScale: Number(e.target.value) })}
            className="mt-1 min-h-[44px] w-full"
          />
        </label>
        <p className="text-sm text-[var(--text-muted)]">
          The page keeps all its content at 200% text on a 360 px screen without
          horizontal scrolling.
        </p>
      </fieldset>

      <fieldset className="settings-group rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-3">
        <legend className="px-1 font-bold">Motion</legend>
        <label className="flex min-h-[44px] items-center gap-2">
          <input
            type="checkbox"
            checked={Boolean(pref.reducedMotion)}
            onChange={(e) => save({ reducedMotion: e.target.checked })}
            className="h-5 w-5"
          />
          <span>Reduce motion (no animations or transitions)</span>
        </label>
        <p className="text-sm text-[var(--text-muted)]">
          Also honors your system "reduced motion" preference.
        </p>
      </fieldset>

      <fieldset className="settings-group rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-3">
        <legend className="px-1 font-bold">Menu preference</legend>
        <div className="flex gap-2">
          {["regular", "jain"].map((id) => (
            <label
              key={id}
              className={`flex min-h-[44px] items-center gap-2 rounded border-2 px-3 ${
                pref.dietPref === id ? "border-[var(--accent)] bg-[var(--surface-muted)]" : "border-[var(--border)]"
              }`}
            >
              <input
                type="radio"
                name="dietPref"
                value={id}
                checked={pref.dietPref === id}
                onChange={() => save({ dietPref: id })}
                className="h-5 w-5"
              />
              <span className="font-bold">{id === "jain" ? "Jain" : "Regular"}</span>
            </label>
          ))}
        </div>
        <p className="text-sm text-[var(--text-muted)]">
          Used as the default tab on the home screen and the default review variant.
        </p>
      </fieldset>

      <div className="settings-actions">
        <p className="settings-autosave">Changes are saved automatically.</p>
        <button
          type="button"
          className="settings-reset"
          onClick={() => save({ dietPref: "regular", cvdMode: "off", textScale: 1, theme: "system", reducedMotion: false })}
        >
          Reset to defaults
        </button>
      </div>

      <p role="alert" tabIndex={-1} className="outline-none">
        {error ? (
          <span className="flex items-center gap-2 font-bold text-[var(--danger)]">
            <WarningTriangleIcon /> {error}
          </span>
        ) : null}
      </p>
      {saved && <p role="status" className="settings-saved">Settings saved.</p>}
    </section>
  );
}
