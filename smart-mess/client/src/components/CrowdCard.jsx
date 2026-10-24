// Crowd status card. Every status is carried by text + icon shape + pattern;
// color only reinforces (ARCHITECTURE.md section 8).

import {
  PersonIcon,
  WarningTriangleIcon,
  ClockIcon,
  MoonIcon,
  QuestionMarkIcon,
} from "./icons.jsx";

const LEVEL = {
  low: { text: "Low crowd", filled: 1, tone: "tone-low" },
  medium: { text: "Medium crowd", filled: 2, tone: "tone-medium" },
  high: { text: "High crowd", filled: 3, tone: "tone-high" },
};

function LevelBar({ filled, colorVar, levelText }) {
  // 3-segment position cue; programmatic name comes from the adjacent text.
  return (
    <div
      className="flex gap-1"
      role="img"
      aria-label={`Crowd bar: ${levelText} — segment ${filled} of 3 filled`}
    >
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          style={i < filled ? { backgroundColor: `var(--${colorVar})` } : undefined}
          className={`h-3 w-8 rounded-sm ${i < filled ? "" : "seg-empty"}`}
        />
      ))}
    </div>
  );
}

export default function CrowdCard({ mess }) {
  const { name, code, crowd } = mess;
  const stale = crowd.state === "stale";
  const closed = crowd.state === "closed";
  const noUpdate = crowd.state === "no_update";
  const levelInfo = crowd.level ? LEVEL[crowd.level] : null;

  const cardBase =
    "relative rounded-md border-2 p-3 bg-[var(--surface)] " +
    (stale || noUpdate ? "border-dashed border-[var(--status-stale)]" : "border-[var(--border)]");

  let statusRow = null;
  let extra = null;

  if (levelInfo && !closed) {
    statusRow = (
      <div className="mt-2 flex flex-wrap items-center gap-2">
        <span
          className={`flex min-h-[36px] items-center gap-1 rounded px-2 py-1 text-sm font-bold ${levelInfo.tone}`}
        >
          {crowd.level === "high" ? <WarningTriangleIcon /> : null}
          {Array.from({ length: 3 }, (_, i) => (
            <PersonIcon key={i} filled={i < levelInfo.filled} />
          ))}
        </span>
        <span className="text-base font-bold">{levelInfo.text}</span>
      </div>
    );
    extra = (
      <div className="mt-2 flex items-center gap-2 text-sm">
        <LevelBar
          filled={levelInfo.filled}
          colorVar={`status-${crowd.level}`}
          levelText={levelInfo.text}
        />
        <span className="flex items-center gap-1 text-[var(--text-muted)]">
          <ClockIcon />
          {crowd.message}
        </span>
      </div>
    );
    if (stale) {
      extra = (
        <div className="mt-2 flex items-center gap-2 text-sm">
          <LevelBar
            filled={levelInfo.filled}
            colorVar={`status-${crowd.level}`}
            levelText={levelInfo.text}
          />
          <span className="flex items-center gap-1 font-medium text-[var(--status-stale)]">
            <ClockIcon />
            {crowd.message}
          </span>
        </div>
      );
    }
  } else if (noUpdate) {
    statusRow = (
      <p className="mt-2 flex items-center gap-2 font-medium text-[var(--status-stale)]">
        <QuestionMarkIcon />
        No update yet
      </p>
    );
  } else if (closed) {
    statusRow = (
      <p className="mt-2 flex items-center gap-2 text-[var(--text-muted)]">
        <MoonIcon />
        {crowd.message}
      </p>
    );
  }

  return (
    <article className={cardBase} data-testid={`crowd-${code}`}>
      <h3 className="text-lg font-bold">{name}</h3>
      <span className="sr-only">{`Crowd status for ${name}.`}</span>
      {statusRow}
      {extra}
    </article>
  );
}
