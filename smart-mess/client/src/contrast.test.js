import { describe, it, expect } from "vitest";
import { readFileSync } from "node:fs";
import { join } from "node:path";

// Automated contrast check (ARCHITECTURE.md section 8, rule 2): tokens are
// verified with a WCAG relative-luminance contrast function, not by eye.

const tokensCss = readFileSync(join(process.cwd(), "src", "theme", "tokens.css"), "utf8");
const indexCss = readFileSync(join(process.cwd(), "src", "index.css"), "utf8");

function hex(v) {
  return /#[0-9a-fA-F]{6}/.exec(v)?.[0];
}

const MODES = ["off", "unsure", "protan", "deutan", "tritan", "mono"];
const VARS = ["status-low", "status-medium", "status-high", "status-stale", "accent", "focus", "bg"];

function varsFor(mode) {
  const vars = {};
  const collect = (modeName) => {
    for (const m of tokensCss.matchAll(/([^{}]+)\{([^{}]*)\}/g)) {
      const selector = m[1];
      const body = m[2];
      const quoted = [...selector.matchAll(/data-cvd="(\w+)"/g)].map((x) => x[1]);
      const isRoot = selector.includes(":root") && quoted.length === 0;
      const applies = quoted.includes(modeName) || (isRoot && (modeName === "off" || modeName === "unsure"));
      if (!applies) continue;
      for (const v of body.matchAll(/--([\w-]+): ([^;]+);/g)) {
        if (VARS.includes(v[1])) vars[v[1]] = v[2].trim();
      }
    }
  };
  void mode;
  collect("off"); // :root base values first…
  if (mode !== "off" && mode !== "unsure") collect(mode); // …then mode overrides
  return vars;
}

function luminance(h) {
  const [r, g, b] = [1, 3, 5].map((i) => parseInt(h.slice(i, i + 2), 16) / 255).map((c) =>
    c <= 0.04045 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4)
  );
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a, b) {
  const [l1, l2] = [luminance(hex(a)), luminance(hex(b))].sort((x, y) => y - x);
  return (l1 + 0.05) / (l2 + 0.05);
}

// .tone-* text colors per mode; per-mode overrides live in index.css and the
// exact values asserted here keep that file honest.
const toneText = (selectorRe) => {
  const m = selectorRe.exec(indexCss);
  return m ? hex(m[1]) : null;
};
const unsureHighText = toneText(/html\[data-cvd="(?:off|unsure)"\] \.tone(?:-high)?[^{]*\{\s*color: (#[0-9a-fA-F]+)/);
const monoMediumText = toneText(/html\[data-cvd="mono"\] \.tone-medium\s*\{\s*color: (#[0-9a-fA-F]+)/);
const monoLowText = toneText(/html\[data-cvd="mono"\] \.tone-low\s*\{\s*color: (#[0-9a-fA-F]+)/);

const BLACK = "#1c1c1e";
const WHITE = "#ffffff";
const PAIRS = {
  off: { "status-low": WHITE, "status-medium": BLACK, "status-high": unsureHighText },
  unsure: { "status-low": WHITE, "status-medium": BLACK, "status-high": unsureHighText },
  protan: { "status-low": WHITE, "status-medium": BLACK, "status-high": WHITE },
  deutan: { "status-low": WHITE, "status-medium": BLACK, "status-high": WHITE },
  tritan: { "status-low": WHITE, "status-medium": BLACK, "status-high": WHITE },
  mono: { "status-low": monoLowText, "status-medium": monoMediumText, "status-high": WHITE },
};

describe("Token contrast (automated)", () => {
  for (const mode of MODES) {
    it(`${mode}: status chip text pairs meet 4.5:1`, () => {
      const vars = varsFor(mode);
      for (const [name, text] of Object.entries(PAIRS[mode])) {
        const ratio = contrast(vars[name], text);
        expect(ratio, `${mode}/${name} ${vars[name]} on ${text}`).toBeGreaterThanOrEqual(4.5);
      }
      expect(contrast(vars["status-stale"], vars.bg.startsWith("#") ? vars.bg : "#f8f8f7")).toBeGreaterThanOrEqual(4.5);
      expect(contrast(vars.focus, "#f8f8f7")).toBeGreaterThanOrEqual(3);
    });
  }

  it("index.css defines the per-mode chip text overrides", () => {
    expect(unsureHighText).toBe("#000000");
    expect(monoMediumText).toBe(WHITE);
    expect(monoLowText).toBe("#000000");
  });
});
