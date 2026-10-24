// Icon-decoration tests: every decorative icon must be hidden from AT (the
// accessible information always lives in adjacent text, ARCHITECTURE.md 8),
// and the person/star "filled" variants must be distinguishable as data.
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import {
  PersonIcon,
  WarningTriangleIcon,
  ClockIcon,
  MoonIcon,
  QuestionMarkIcon,
  CheckIcon,
  LockIcon,
  PlateIcon,
  LeafIcon,
  StarIcon,
  CrossIcon,
  PlusIcon,
} from "./components/icons.jsx";

const ALL = { PersonIcon, WarningTriangleIcon, ClockIcon, MoonIcon, QuestionMarkIcon, CheckIcon, LockIcon, PlateIcon, LeafIcon, StarIcon, CrossIcon, PlusIcon };

describe("icons are decorative and not focusable", () => {
  for (const [name, Icon] of Object.entries(ALL)) {
    it(`${name} is aria-hidden and not focusable`, () => {
      render(<Icon />);
      const svg = document.querySelector("svg");
      expect(svg.getAttribute("aria-hidden")).toBe("true");
      expect(svg.getAttribute("focusable")).toBe("false");
    });
  }
});

describe("filled-icon data variants", () => {
  it("PersonIcon filled and outline are distinguishable", () => {
    const { container: a } = render(<PersonIcon filled />);
    const { container: b } = render(<PersonIcon filled={false} />);
    expect(a.innerHTML).not.toBe(b.innerHTML);
  });
  it("StarIcon filled and outline are distinguishable", () => {
    const { container: a } = render(<StarIcon filled />);
    const { container: b } = render(<StarIcon filled={false} />);
    expect(a.innerHTML).not.toBe(b.innerHTML);
  });
  it("status icons for fresh/stale (Clock), closed (Moon), no-update (Question) are distinct glyph shapes", () => {
    const clip = (Icon) => {
      const { container } = render(<Icon />);
      return container.querySelector("svg").innerHTML.replace(/\s+/g, " ").trim();
    };
    const shapes = new Set([clip(ClockIcon), clip(MoonIcon), clip(QuestionMarkIcon)]);
    expect(shapes.size).toBe(3);
  });
});
