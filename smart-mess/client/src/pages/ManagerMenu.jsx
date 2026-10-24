// Main manager menu editor (route /manager/menu). Date picker, four meal tabs,
// dish editor, default items prefill, Save draft and Publish. The server
// scopes everything to the manager's own hostel group.

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getManagerMeals,
  getDefaultItems,
  saveMeal,
  publishMeal,
  todayStr,
  ApiError,
} from "../api";
import { MealTabs } from "../components/Tabs.jsx";
import { CheckIcon, ClockIcon, PlusIcon, CrossIcon, WarningTriangleIcon } from "../components/icons.jsx";

const cap = (s) => s.charAt(0).toUpperCase() + s.slice(1);
const DIETS = ["both", "regular", "jain"];
const DIET_LABEL = { both: "Both", regular: "Regular", jain: "Jain" };
const MEALS_BY_ORDER = ["breakfast", "lunch", "snacks", "dinner"];

function blankItem() {
  return { name: "New dish name", dietMenu: "both", sortOrder: -1 };
}

export default function ManagerMenu() {
  const [date, setDate] = useState(todayStr());
  const [mealType, setMealType] = useState("lunch");

  const mealsQuery = useQuery({
    queryKey: ["managerMeals", date],
    queryFn: () => getManagerMeals(date),
    enabled: Boolean(date),
  });
  const defaultsQuery = useQuery({
    queryKey: ["defaultItems", mealType],
    queryFn: () => getDefaultItems(mealType),
    enabled: Boolean(mealType),
  });

  // Local editor state, reseeded from the server on tab switch.
  const [items, setItems] = useState([]);
  const [notice, setNotice] = useState(null);
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);

  const loadedItems = items.length === 0 && mealsQuery.data ? serverItems(mealsQuery.data, mealType) : items;

  const current = mealsQuery.data?.meals?.find((m) => m.mealType === mealType);
  const status = current?.status || "none";
  const publishedAt = current?.publishedAt;

  function switchMeal(type) {
    setMealType(type);
    setItems([]);
    setNotice(null);
    setError(null);
  }

  function switchDate(d) {
    setDate(d);
    setItems([]);
    setNotice(null);
    setError(null);
  }

  function updateItem(i, patch) {
    setLoaded();
    const base = loadedItems;
    setItems(base.map((it, idx) => (idx === i ? { ...it, ...patch } : it)));
  }

  function setLoaded() {
    if (items.length === 0 && mealsQuery.data) {
      setItems(serverItems(mealsQuery.data, mealType));
    }
  }

  function removeItem(i) {
    setLoaded();
    setItems(loadedItems.filter((_, idx) => idx !== i));
  }

  function addItem() {
    setLoaded();
    setItems([...loadedItems, blankItem()]);
  }

  function prefill() {
    const defaults = (defaultsQuery.data?.items || []).map((it, i) => ({
      name: it.name,
      dietMenu: "both",
      sortOrder: it.sortOrder ?? i,
    }));
    setLoaded();
    setItems([...loadedItems, ...defaults]);
  }

  async function onSave(publish) {
    setLoaded();
    const list = loadedItems;
    if (list.some((it) => !it.name.trim())) {
      setNotice({ ok: false, text: "Every dish needs a name before saving." });
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const ordered = list.map((it, i) => ({ name: it.name.trim(), dietMenu: it.dietMenu, sortOrder: i }));
      await saveMeal(date, mealType, ordered);
      if (publish) await publishMeal(date, mealType);
      await mealsQuery.refetch();
      setItems([]); // reseed from the saved server state
      setNotice({
        ok: true,
        text: publish
          ? `Published the ${cap(mealType.toLowerCase())} menu for ${date}. Students now see it.`
          : `Draft saved for ${cap(mealType.toLowerCase())} on ${date}.`,
      });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save the menu. Try again.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section aria-labelledby="menu-mgr-h" className="space-y-4">
      <h1 id="menu-mgr-h" className="text-2xl font-bold">
        Edit the menu
      </h1>
      <p className="text-sm text-[var(--text-muted)]">
        All messes in your hostel group serve the same menu. The food changes
        daily, so pick the service date and set what is served that day.
      </p>

      <label className="flex items-center gap-2">
        <span className="font-bold">Service date</span>
        <input
          type="date"
          value={date}
          onChange={(e) => switchDate(e.target.value)}
          className="min-h-[44px] rounded border-2 border-[var(--border)] bg-[var(--surface)] px-2"
        />
      </label>

      {mealsQuery.isLoading && <p role="status" className="font-bold">Loading the menu…</p>}
      {mealsQuery.isError && (
        <p role="alert" className="font-bold text-[var(--danger)]">
          Could not load the menu. <button type="button" className="underline" onClick={() => mealsQuery.refetch()}>Retry</button>
        </p>
      )}

      {mealsQuery.data && (
        <div id="menu-editor-panel" role="tabpanel" aria-label={`${mealType} editor`} className="space-y-3">
          <MealTabs
            meals={MEALS_BY_ORDER.map((t) => ({ mealType: t, scheduleLabel: "" }))}
            selected={mealType}
            onSelect={switchMeal}
            panelId="menu-editor-panel"
          />
          <p className="text-sm text-[var(--text-muted)]">
            <span className="inline-flex items-center gap-1">
              <ClockIcon /> Status: <strong>{status}</strong>
              {publishedAt ? ` · published ${String(publishedAt).slice(0, 16).replace("T", " ")}Z` : ""}
            </span>{" "}
            · Menus can be edited for today and future dates.
          </p>

          <ul className="space-y-2">
            {loadedItems.map((it, i) => (
              <li key={i} className="flex flex-wrap items-center gap-2 rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-2">
                <label className="flex-1 min-w-[220px]">
                  <span className="block text-sm font-bold">Dish name</span>
                  <input
                    type="text"
                    value={it.name}
                    onChange={(e) => updateItem(i, { name: e.target.value })}
                    aria-label={`Dish ${i + 1} name`}
                    className="mt-0.5 min-h-[44px] w-full rounded border-2 border-[var(--border)] bg-[var(--bg)] p-2"
                  />
                </label>
                <label>
                  <span className="block text-sm font-bold">Variant</span>
                  <select
                    value={it.dietMenu}
                    onChange={(e) => updateItem(i, { dietMenu: e.target.value })}
                    aria-label={`Dish ${i + 1} variant`}
                    className="mt-0.5 min-h-[44px] rounded border-2 border-[var(--border)] bg-[var(--surface)] px-2"
                  >
                    {DIETS.map((d) => (
                      <option key={d} value={d}>
                        {DIET_LABEL[d]}
                      </option>
                    ))}
                  </select>
                </label>
                <button
                  type="button"
                  onClick={() => removeItem(i)}
                  className="mt-4 flex min-h-[44px] items-center gap-1 rounded border-2 border-[var(--border)] bg-[var(--surface)] px-3"
                >
                  <CrossIcon /> Remove
                </button>
              </li>
            ))}
          </ul>

          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={addItem}
              className="flex min-h-[44px] items-center gap-1 rounded border-2 border-[var(--border)] bg-[var(--surface)] px-3 font-bold"
            >
              <PlusIcon /> Add dish
            </button>
            <button
              type="button"
              onClick={prefill}
              className="flex min-h-[44px] items-center gap-1 rounded border-2 border-[var(--border)] bg-[var(--surface)] px-3 font-bold"
            >
              <PlusIcon /> Prefill daily items
            </button>
          </div>

          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              disabled={busy}
              onClick={() => onSave(false)}
              className="min-h-[48px] flex-1 rounded border-2 border-[var(--border)] bg-[var(--surface)] px-4 font-bold disabled:opacity-60"
            >
              <CheckIcon /> <span className="ml-1">Save draft</span>
            </button>
            <button
              type="button"
              disabled={busy}
              onClick={() => onSave(true)}
              className="min-h-[48px] flex-1 rounded border-2 border-[var(--accent)] bg-[var(--surface-muted)] px-4 font-bold disabled:opacity-60"
            >
              <CheckIcon /> <span className="ml-1">{busy ? "Saving…" : "Save and Publish"}</span>
            </button>
          </div>
        </div>
      )}

      {notice && (
        <p
          role="status"
          className={`flex items-center gap-2 rounded-md border-2 p-3 font-bold ${
            notice.ok
              ? "border-[var(--accent)] text-[var(--ok)]"
              : "border-[var(--danger)] text-[var(--danger)]"
          }`}
        >
          {notice.ok ? <CheckIcon /> : <WarningTriangleIcon />} {notice.text}
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

function serverItems(data, mealType) {
  const m = data.meals?.find((x) => x.mealType === mealType);
  return (m?.items || []).map((it) => ({ name: it.name, dietMenu: it.dietMenu, sortOrder: it.sortOrder }));
}
