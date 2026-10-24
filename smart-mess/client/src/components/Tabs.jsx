import { useRef } from "react";
import { PlateIcon, LeafIcon } from "./icons.jsx";

// Tabs follow WAI-ARIA tab pattern: roving tabindex, arrow keys move focus
// and selection (keyboard-only must work).
function useArrowKeys(count, current, onChange) {
  const refs = useRef([]);
  const onKeyDown = (e) => {
    if (e.key !== "ArrowRight" && e.key !== "ArrowLeft") return;
    e.preventDefault();
    const dir = e.key === "ArrowRight" ? 1 : -1;
    const next = (current + dir + count) % count;
    onChange(next);
    refs.current[next]?.focus();
  };
  return { onKeyDown, refs };
}

export function MealTabs({ meals, selected, onSelect, panelId }) {
  const index = meals.findIndex((m) => m.mealType === selected);
  const { onKeyDown, refs } = useArrowKeys(meals.length, index, (i) => onSelect(meals[i].mealType));

  return (
    <div role="tablist" aria-label="Meals today" onKeyDown={onKeyDown} className="flex flex-wrap gap-2">
      {meals.map((m, i) => (
        <button
          key={m.mealType}
          ref={(el) => (refs.current[i] = el)}
          role="tab"
          type="button"
          aria-selected={selected === m.mealType}
          aria-controls={panelId}
          tabIndex={selected === m.mealType ? 0 : -1}
          onClick={() => onSelect(m.mealType)}
          className={`flex min-h-[44px] flex-col items-start rounded border-2 px-3 py-1 ${
            selected === m.mealType
              ? "border-[var(--text)] bg-[var(--text)] text-white"
              : "border-[var(--border)] bg-[var(--surface)]"
          }`}
        >
          <span className="font-bold">{m.mealType.charAt(0).toUpperCase() + m.mealType.slice(1)}</span>
          <span className="text-xs">{m.scheduleLabel}</span>
        </button>
      ))}
    </div>
  );
}

export function VariantTabs({ selected, onSelect, panelId }) {
  const variants = [
    { id: "regular", label: "Regular", Icon: PlateIcon },
    { id: "jain", label: "Jain", Icon: LeafIcon },
  ];
  const index = variants.findIndex((v) => v.id === selected);
  const { onKeyDown, refs } = useArrowKeys(variants.length, index, (i) => onSelect(variants[i].id));

  return (
    <div role="tablist" aria-label="Menu variant" onKeyDown={onKeyDown} className="flex gap-2">
      {variants.map(({ id, label, Icon }, i) => (
        <button
          key={id}
          ref={(el) => (refs.current[i] = el)}
          role="tab"
          type="button"
          aria-selected={selected === id}
          aria-controls={panelId}
          tabIndex={selected === id ? 0 : -1}
          onClick={() => onSelect(id)}
          className={`flex min-h-[44px] items-center gap-2 rounded border-2 px-3 py-1 font-bold ${
            selected === id
              ? "border-[var(--text)] bg-[var(--text)] text-white"
              : "border-[var(--border)] bg-[var(--surface)]"
          }`}
        >
          <Icon />
          {label}
        </button>
      ))}
    </div>
  );
}

export function VariantNote({ variant }) {
  if (variant === "jain") {
    return (
      <p className="flex items-center gap-2 text-sm text-[var(--text-muted)]">
        <LeafIcon />
        Prepared without onion and garlic
      </p>
    );
  }
  return null;
}
