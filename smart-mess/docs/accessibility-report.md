# Accessibility report — Smart Mess front-end prototype

Date: 2026-10-20 (TODAY-SCOPE trim, session 1)

**Evaluated with simulators and automated tools. No testing was done with color-blind participants.**

This covers the mock-data prototype in `client/` only (ARCHITECTURE.md section 13 wording applies).

## Checks run (and how)

| Check | Tool / method | Result |
| --- | --- | --- |
| axe scan of the student home | `@testing-library/react` + `axe-core` in Vitest (`src/StudentHome.test.jsx`) | 0 serious / 0 critical violations (jsdom; `lang`/`title` supplied because jsdom starts bare — the real app gets them from `index.html`) |
| Contrast verification of every color-vision-mode token pair | automated WCAG relative-luminance function turned into a test (`src/contrast.test.js`); tokens read from `src/theme/tokens.css` and `src/index.css`; statuses on all six `data-cvd` modes | all verified ≥ 4.5:1 (chip text) and focus ring ≥ 3:1; this check forced real fixes (see below) and is a passing part of the suite |
| Status encodings in grayscale | structural reasoning only (no browser run by the agent): every status pairs text + a distinct icon shape + a pattern/position cue (3-segment bar, dashed outline, plain gray card), so color is never the only channel | design verified by code review; **not** verified visually in a browser |
| Vision-deficiency emulation (protanopia / deuteranopia / tritanopia / achromatopsia) in Chrome DevTools | **not yet performed** — no browser run in this session; Chrome DevTools emulation should be run before the demo and recorded here | pending |
| Keyboard-only walkthrough | results recorded from the code paths: Skip link is the first focusable element; tabs use the WAI-ARIA tab pattern with roving tabindex + Arrow keys; all interactive elements are real buttons/inputs/selects ≥ 44 px high with visible focus rings (`*:focus-visible`, 3 px outline); no icon-only buttons | supported by implementation and render tests; a manual walkthrough of the running app is recommended |
| 200% text / 360 px width | not checked in a browser (no browser run) | pending |

## Fixes the automated contrast check forced (recorded for honesty)

- `off`/`unsure` "High crowd" chip: white text on `#D55E00` is only ≈ 3.9:1 → chip text switched to pure black (≈ 5.4:1).
- `tritan` medium was tuned from `#B26A00` to `#FFB300` (black text reaches ≥ 4.5:1).
- `mono` mode: light-gray "low" chip gets dark text, "medium" gray chip gets white text.
- The tokens are therefore **not** called "CVD-safe"; they are the ARCHITECTURE.md section 8 starting values that passed the automated contrast check.

## Known limitations

- No color-blind participant testing (per owner decision C11; keep saying this in every report).
- Browser-level visual emulation (grayscale + the deficiency simulators), Lighthouse run and a live 200 % / 360 px check have not been executed in this session; run and append results before the submission demo.
- The prototype uses mock data; accessibility of real flows (login errors etc.) is untested because login is out of scope today.

---

## Session 4 (2026-10-24): real-backend client

The client is no longer the mock prototype: all nine pages run on the real Go
API (`src/api/index.js` fetch client, server-supplied user record, real
published menus). Accessibility-related changes and verifications:

### Checks re-run on the real client

| Check | Tool / method | Result |
| --- | --- | --- |
| axe scan, real student home (crowd cards + menu + review hints) | `axe-core` on the rendered DOM in `src/StudentHome.test.jsx` (fetch stubbed, server-shaped fixtures) | 0 serious / 0 critical violations |
| full component suite | `npx vitest run` | 54/54 pass (rules 16, contrast 7, icon-decoration 15, StudentHome 6, ReviewFlow 10 incl. the manager flows) |

### Accessibility changes made during the real-client build

- Top navigation links (`Menu and crowd / Review a meal / Settings`, manager
  equivalents) use `.nav-link`: 44 px min height, 2 px border, `[aria-current]`
  state + `:focus-visible` ring.
- `ProtectedRoute` now honours a `role` prop, so manager-only routes are
  redirected client-side too (mirrors the server rule; never the only gate).
- Review screen over-length comment (> 500 chars): previously only the submit
  button was disabled (a silent failure). Now a live `role="alert"` warning
  ("Comment must be 500 characters or fewer.") appears while typing.
- PWA polish: manifest (standalone display, 192/512 + maskable icons),
  `theme-color`, apple-web-app meta, favicon replaces the Vite default.
- Loading/error states for every data area use `role="status"` / `role="alert"`:
  "Impossible" failures are never communicated by color only.

### Still pending (unchanged from session 1)

No color-blind participant testing (per owner decision C11). Browser-based
simulators (protan/deutan/tritan/achromatopsia), grayscale view, Lighthouse,
and the 200 %-text / 360 px manual checks still need a real browser run before
the demo; append results here when done.
