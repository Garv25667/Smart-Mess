// Small status icons. Rule: no icon appears without visible text next to it
// (ARCHITECTURE.md section 8, rule 9); icons carry the *shape* channel of a
// status, never the whole status.

const base = {
  width: 20,
  height: 20,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 2,
  strokeLinecap: "round",
  strokeLinejoin: "round",
  "aria-hidden": "true",
  focusable: "false",
};

export function PersonIcon({ filled = false }) {
  return (
    <svg {...base} fill={filled ? "currentColor" : "none"}>
      <circle cx="12" cy="7" r="3.2" />
      <path d="M5 20c1.5-3.6 4-5.3 7-5.3s5.5 1.7 7 5.3" />
    </svg>
  );
}

export function WarningTriangleIcon() {
  return (
    <svg {...base} fill="currentColor" stroke="none">
      <path d="M12 3.5 22 20H2L12 3.5Z" />
      <path d="M12 9.5v4.5" stroke="#ffffff" strokeWidth="1.8" />
      <circle cx="12" cy="17" r="1.1" fill="#ffffff" />
    </svg>
  );
}

export function ClockIcon() {
  return (
    <svg {...base}>
      <circle cx="12" cy="12" r="8.5" />
      <path d="M12 7.5V12l3 2" />
    </svg>
  );
}

export function MoonIcon() {
  return (
    <svg {...base}>
      <path d="M20 14.5A8.5 8.5 0 0 1 9.5 4 8.5 8.5 0 1 0 20 14.5Z" />
    </svg>
  );
}

export function QuestionMarkIcon() {
  return (
    <svg {...base}>
      <circle cx="12" cy="12" r="8.5" />
      <path d="M9.7 9.5a2.4 2.4 0 1 1 3.5 2.2c-.9.5-1.2 1-1.2 1.8" />
      <circle cx="12" cy="16.4" r="1" fill="currentColor" />
    </svg>
  );
}

export function CheckIcon() {
  return (
    <svg {...base}>
      <path d="M5 12.5 10 17.5 19 7" />
    </svg>
  );
}

export function LockIcon() {
  return (
    <svg {...base}>
      <rect x="5.5" y="10.5" width="13" height="9" rx="1.5" />
      <path d="M8.5 10.5V7.8a3.5 3.5 0 0 1 7 0v2.7" />
    </svg>
  );
}

export function PlateIcon() {
  // "Regular" variant icon: circle plate with fork and knife.
  return (
    <svg {...base}>
      <circle cx="12" cy="12" r="6.5" />
      <circle cx="12" cy="12" r="2.2" />
    </svg>
  );
}

export function LeafIcon() {
  // "Jain" variant icon: leaf.
  return (
    <svg {...base}>
      <path d="M5 19c0-8 4-13 14-14 1 10-5 14-11 13" />
      <path d="M5 19c2-5 5-9 9-11" />
    </svg>
  );
}

export function StarIcon({ filled = false }) {
  return (
    <svg {...base} fill={filled ? "currentColor" : "none"}>
      <path d="M12 3.6 14.6 9l5.8.6-4.3 4 1.2 5.8L12 16.6 6.7 19.4l1.2-5.8-4.3-4L9.4 9 12 3.6Z" />
    </svg>
  );
}

export function CrossIcon() {
  return (
    <svg {...base}>
      <path d="M6 6l12 12M18 6 6 18" />
    </svg>
  );
}

export function PlusIcon() {
  return (
    <svg {...base}>
      <path d="M12 5v14M5 12h14" />
    </svg>
  );
}
