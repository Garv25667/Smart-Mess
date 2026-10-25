import { StarIcon, PlateIcon, LeafIcon } from "./icons.jsx";

export function StarRow({ stars, outOf = 5 }) {
  const cells = Array.from({ length: outOf }, (_, i) => i < stars);
  return (
    <span className="inline-flex" aria-hidden="true">
      {cells.map((filled, i) => (
        <StarIcon key={i} filled={filled} />
      ))}
    </span>
  );
}

// Per-variant rating summary: number + stars + count (never color alone).
export function RatingSummary({ meal }) {
  const variants = [
    { id: "regular", label: "Regular", Icon: PlateIcon },
    { id: "jain", label: "Jain", Icon: LeafIcon },
  ];
  return (
    <section aria-label="Ratings" className="rounded-md border-2 border-[var(--border)] bg-[var(--surface)] p-3">
      <h3 className="text-base font-bold">Ratings ({meal.mealType})</h3>
      <ul className="mt-1 space-y-1">
        {variants.map(({ id, label, Icon }) => {
          const r = meal.ratings?.[id] || { avg: null, count: 0 };
          const hasAvg = r.avg != null;
          return (
            <li key={id} className="flex flex-wrap items-center gap-2">
              <span className="flex min-h-[36px] items-center gap-1 rounded bg-[var(--surface-muted)] px-2 text-sm font-bold">
                <Icon />
                {label}
              </span>
              {hasAvg ? (
                <>
                  <StarRow stars={Math.round(r.avg)} />
                  <span className="font-bold">{r.avg} / 5</span>
                  <span className="text-sm text-[var(--text-muted)]">({r.count} reviews)</span>
                </>
              ) : (
                <span className="font-bold">No ratings yet</span>
              )}
            </li>
          );
        })}
      </ul>
    </section>
  );
}

// Rating input: 5 radios, each label has the star icon and a visible number.
export function StarInput({ value, onChange }) {
  const descriptions = ["Terrible", "Poor", "Okay", "Good", "Excellent"];
  return (
    <fieldset className="mt-3">
      <legend className="font-bold">Your rating</legend>
      <div className="flex gap-1" role="radiogroup" aria-label="Rating out of 5">
        {[1, 2, 3, 4, 5].map((n) => (
          <label
            key={n}
            className={`flex min-h-[44px] min-w-[44px] cursor-pointer flex-col items-center justify-center rounded border-2 px-2 ${
              value === n ? "border-[var(--accent)] bg-[var(--surface-muted)]" : "border-[var(--border)] bg-[var(--surface)]"
            }`}
          >
            <input
              type="radio"
              name="stars"
              value={n}
              aria-label={String(n)}
              checked={value === n}
              onChange={() => onChange(n)}
              className="absolute h-px w-px opacity-0"
            />
            <span className="flex items-center gap-0.5">
              <StarIcon filled={value >= n} />
              <span className="text-xs font-bold">{n}</span>
            </span>
            <span className="rating-caption">{descriptions[n - 1]}</span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}
