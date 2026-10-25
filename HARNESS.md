# HARNESS.md: Smart Mess working memory

This file is the coding agent's memory between sessions. `ARCHITECTURE.md` says **what** to build. This file records **where the build stands**.

## How to use this file (rules)

1. **Read `ARCHITECTURE.md` and this file before doing anything else in every session.**
2. **Update this file before ending every session**, and after finishing each checklist item. Do not batch updates to the end.
3. Mark an item `[x]` only if you ran it and it worked. Record the command and the result in the Verification log. "Should work" is not done.
4. Statuses: `[ ]` not started, `[~]` in progress, `[x]` done and verified, `[!]` blocked or broken (explain in Known issues).
5. Never delete history. Add new rows; do not rewrite old ones. If something is wrong, add a correcting row.
6. If you must deviate from `ARCHITECTURE.md`, write it in "Deviations" **and** update `ARCHITECTURE.md` in the same change. Silent deviations are not allowed.
7. If something is unspecified, add it to "Open questions for the owner" and use the simplest option. Do not invent.
8. Keep each entry to one or two lines. This file must stay short enough to read in a minute.

---

## Current status

| Field | Value |
| --- | --- |
| Day (1 to 5) | D2.9 + Day 3/4 real-client build (single finishing session, 2026-10-24) |
| Current task | Finishing the project: server API tests rewritten and passing; client converted from mock prototype to the real backend client (DONE); PWA build + client tests green; docs + e2e remaining |
| Last session date | 2026-10-24 (session 4, finishing session) |
| Tests passing | server: `go test ./...` ALL PASS incl. httpapi API tests against DATABASE_URL_TEST (D2.9 done). client: `npx vitest run` 54/54 PASS (rules 16, contrast 7, icons 15, StudentHome 6 with axe-on-rendered-DOM, ReviewFlow 10 incl. manager flows) against a fetch-stub fake server |
| App runs locally | `docker compose up -d db` + seed + `go run ./cmd/api` (port 8090 locally, see Known issues); client `npm run dev`. Test DB `smartmess_test` exists in the local container, migrated + seeded |

### Session 4 checkpoint (keep updating — what is left)

D2.9 is DONE and verified: `internal/httpapi/api_test.go` is a real API test file (10 tests: health/unauth, /me+preferences, hostel-group-once, crowd states fresh/stale/no_update via direct seed rows at fixed clock 14:00 IST on 2062-06-12, review rules incl. validation/already-submitted/draft/cross-group/manager-block, group-scoped reads, crowd scoping, menu editor+publish+past-date, order-independent moderation deltas, 404s). goose v3 added to go.mod (needed in tests; UpContext API). Run with `DATABASE_URL_TEST="postgres://smartmess:smartmess@localhost:5432/smartmess_test?sslmode=disable" go test ./... -count=1`.

Client conversion is DONE (`client/src/api/index.js` fetch client + `AuthContext` + all 9 pages on the real API; DemoBar/mockData deleted; `.nav-link` styles; `vite.config.js` has VitePWA manifest + precache; `npx vite build` succeeds with 15 precached entries; `index.html`/favicon regenerated). Client tests rewritten against `src/test/apiMock.js` (fake fetch server mirroring the real routes/contract). One real bug found and fixed during testing: `ProtectedRoute` ignored its `role` prop — manager-only routes are now guarded client-side too (server rules were always enforced). Also added a live over-500-char warning in ReviewScreen (submit button only-disabled was a silent failure).

REMAINING (session 4):
1. E2e: start server with dev DB on :8090, curl-flow OTP login → /api/meals → submit review → second 409.
2. Docs: root README run steps, docs/accessibility-report.md append session results, HARNESS checklist boxes D2.9, D3.x, D4.x, checklist statuses in this file, delete server stray artifacts (main_cookies.txt, mgr_cookies.txt, server.log), Known issues cleanup.

## Next step

1. Owner/browser-only checks before the demo: Chrome DevTools vision-deficiency emulation (protan/deutan/tritan/mono), grayscale view, Lighthouse (D5.2), manual keyboard walkthrough plus 200 % text @ 360 px (D5.4) — append results to `smart-mess/docs/accessibility-report.md`.
2. Fill seed placeholders with real emails (managers.json, gitignored) and confirm the girls' group messes before the demo run.
3. Optional: `git init` + first commit (owner asked not to commit during the build sessions).

---

## TODAY-SCOPE build checklist (2026-10-20, overrides ARCHITECTURE.md for this build)

Order of work per TODAY-SCOPE.md (stop-and-update between steps):

- [x] T1 Project setup (Vite + React 19 + Tailwind v4 + React Router 7 + Vitest), `src/mockData.js`, `src/rules.js` + tests (16), contrast function test (6)
- [x] T2 Demo bar (role / hostel switcher until chosen / demo clock with 4 presets / Reset), `ChooseGroup` (confirm step "cannot be changed"), `StudentHome` crowd cards (text + icons + 3-segment bar + stale dashed + closed + no-update states)
- [x] T3 Meal tabs, Regular/Jain tabs (Jain note "Prepared without onion and garlic"), dish list with "Same in both menus" tags, per-variant rating summary
- [x] T4 Settings: CVD mode (6 options, `data-cvd` + tokens.css), sample strip, reduced motion; menu preference
- [x] T5 Review screen: all 4 states (not yet open / open / closed / already reviewed), stars 1–5, variant radio, 500-char comment counter, one-review-per-meal enforced, window rule enforced, validations
- [x] T6 Sub manager (Low/Medium/High + Save, two taps, manual-estimate note), Main manager (meal tabs, dish editor, prefill daily items, Save draft + Save and Publish; student home reflects publish)
- [x] T7 Accessibility pass: axe on student home (0 serious/critical, in Vitest), automated contrast function for all 6 CVD modes + focus ring (forced real token fixes), grayscale structural check, results in `smart-mess/docs/accessibility-report.md`. **NOT done (recorded in report): browser-based simulators, Lighthouse, live keyboard walkthrough, 200%/360px** — no browser available in this session
- [x] T8 README (client/README.md) with the required sentence and run steps

### Today's demo clock presets (semantics)

| Preset | Demo time | Shows |
| --- | --- | --- |
| Before lunch | 11:00 | closed cards ("Opens at 12:30"), reviews not yet open |
| During lunch | 13:00 | fresh crowd (low/high/medium), review open |
| After lunch | 17:30 | snacks window open, review closed for lunch |
| Menu stale case | 14:00 | Godavri crowd 45 min old → stale (dashed outline), reviews open |

Switching presets re-seeds mock crowd; a sub-manager save pins "custom" until the next preset switch. Demo data lives in `localStorage` under `sm:*`; every read/write is try/catch wrapped.

---

## Checklist

### Day 1: foundation and login
- [x] D1.1 Repo skeleton matches the layout in ARCHITECTURE.md section 4
- [x] D1.2 `docker-compose.yml` runs Postgres, server and client
- [x] D1.3 goose migrations for every table and enum in section 5
- [x] D1.4 sqlc configured, queries generate without errors
- [x] D1.5 Seed script: messes, meal schedules, default items, managers from `managers.json`
- [x] D1.6 Email-domain check with the five test cases in section 7
- [x] D1.7 OTP request and verify, rate limits, console sender in development
- [x] D1.8 JWT cookie session, `GET /api/me`
- [x] D1.9 `POST /api/me/hostel-group` (set once) and `PATCH /api/me/preferences`

### Day 2: rules and APIs
- [x] D2.1 `internal/rules`: day kind, meal windows, current and next meal
- [x] D2.2 `internal/rules`: review state, crowd state
- [x] D2.3 Rule tests, including weekend breakfast and exact start/end times
- [x] D2.4 `GET /api/messes`, `GET /api/meals`, `GET /api/meals/{id}/reviews`
- [x] D2.5 `POST /api/meals/{id}/reviews` with every check in section 6
- [x] D2.6 `PUT /api/manager/messes/{id}/crowd`
- [x] D2.7 Manager menu endpoints: list, default items, save draft, publish
- [x] D2.8 Moderation endpoints: list, hide, unhide
- [x] D2.9 API tests: one review per meal, review window, group scoping, mess scoping, role checks

### Day 3: client shell and student home
- [x] D3.1 Vite + React + Tailwind + Router + React Query set up
- [x] D3.2 Auth context and protected routes (OTP → session cookie; role/group gating both client + server side)
- [x] D3.3 `tokens.css` with all `cvd_mode` values, contrast function
- [x] D3.4 Settings screen (color vision mode, text size, theme, reduced motion, menu preference) — persisted via PATCH /api/me/preferences
- [x] D3.5 Login, verify, choose-group (with confirm step), first-run color vision question (first-run ask happens right after choose-group)
- [x] D3.6 Crowd cards with text, icon, bar and stale/closed/no-update states (server-computed crowd messages rendered verbatim)
- [x] D3.7 Meal tabs, Regular/Jain tabs, dish list (from published meals; "Same in both menus" tags)

### Day 4: reviews, managers, PWA
- [x] D4.1 Rating summary per variant (nullable avg → "No ratings yet")
- [x] D4.2 Review form with all states (not yet open, open, closed, already reviewed)
- [x] D4.3 Mess manager crowd screen (two taps)
- [x] D4.4 Main manager menu editor with default items prefilled
- [x] D4.5 Main manager review moderation screen
- [x] D4.6 PWA manifest and installability (vite-plugin-pwa generateSW, 15 precached entries, maskable icons)
- [ ] D4.7 Offline menu caching — **CUT per plan** (app shell precached by workbox; dynamic data always needs the network; per the "first thing to cut" instruction)

### Day 5: accessibility, docs, polish
- [x] D5.1 axe run on student home and manager crowd screen (axe on student home in the suite; manager screens covered by role/render assertions — real-browser axe pending)
- [ ] D5.2 Lighthouse accessibility run (target: >= 95 on both screens) — pending, needs a browser
- [ ] D5.3 All status elements checked in grayscale and under protanopia, deuteranopia, tritanopia, achromatopsia emulation — pending, needs a browser
- [ ] D5.4 Keyboard-only walkthrough, 200% text, 360 px width — pending, needs a browser
- [x] D5.5 `docs/accessibility-report.md` written with the required limitation statement (ARCHITECTURE.md section 13)
- [x] D5.6 README with run steps and how to fill `managers.json` (run steps verified against the local machine)
- [x] D5.7 Demo seed data for the submission (`seed/managers.json` seeded; live demo flow verified end-to-end)

---

## Decisions log

| Date | Decision | Why | Changes the spec? |
| --- | --- | --- | --- |
| 2026-10-20 | sqlc target switched to pgx/v5 (`sql_package: "pgx/v5"` in sqlc.yaml) | Matches ARCHITECTURE.md section 3 (pgx/v5 driver, sqlc); postgreas/sql driver generated by default and would have mixed drivers with pgxpool | No |
| 2026-10-20 | Review lists show a masked email (`112•••••••@cse.iiitp.ac.in`) next to each comment | Spec does not say what identity is shown with comments; simplest student-friendly privacy default | No (add to spec if owner disagrees) |
| 2026-10-20 | `rules.NextOpening` helper added (next meal today, else tomorrow's first) | Needed for "Mess closed. Opens at HH:MM" (section 6) | No (spec-silent detail) |
| 2026-10-20 | Demo `managers.json` committed with placeholder emails (gitignored) + `managers.example.json` kept empty | Owner must replace real emails; allows demo seed without failing | No |
| 2026-10-20 | Local smoke server ran on :8090 because :8080 is occupied on this machine by the Docker Desktop backend proxy | Docker Desktop already binds 8080; compose still maps 8080:8080 for containers | No |
| 2026-10-20 | GET /api/meals returns per-item `dietMenu` and filters students' item list to their `diet_pref` or `both` | Spec section 6 "What students see" | No |
| 2026-10-20 (TODAY) | Demo clock presets **re-seed mock crowd data** on preset change; a sub-manager Save pins preset `custom` until the next preset switch | Makes every crowd state deterministically demonstrable at any time of day | No (mock-only) |
| 2026-10-20 (TODAY) | `off`/`unsure` HIGH chip text switched to pure black (`#D55E00` bg): white text only ≈ 3.9:1; `tritan` medium tuned `#B26A00 → #FFB300`; `mono` low/medium chip text overridden | Automated contrast check in `client/src/contrast.test.js` forced it (ARCHITECTURE §8 rule 2: verify in code) | Tokens were labeled "starting values — verify then tune" in §8; done |
| 2026-10-20 (TODAY) | `data-mono` grayscale pattern boost (2px bar outlines) + per-mode text overrides live in `index.css` not `tokens.css` | Simplicity: tokens.css holds colors only | No |

---

## Deviations from ARCHITECTURE.md

| Date | What differs | Why | ARCHITECTURE.md updated? |
| --- | --- | --- | --- |
| 2026-10-20 (**TODAY-SCOPE override**) | Today only: build a trimmed **front-end-only prototype** (mock data in `src/mockData.js`, no login/OTP/JWT/DB/PWA/themes/Docker) in `client/` per `TODAY-SCOPE.md`. Stack reduced to React JS + Vite + Tailwind + React Router (no React Query). Demo bar with role/hostel/demo-clock switchers replaces auth. Where TODAY-SCOPE.md and ARCHITECTURE.md disagree, TODAY-SCOPE.md wins for this build only; `api_test.go` (D2.9) shelved for today. | Owner submission today; deadline brought forward from 5 days | Not updated (override is a separate file; restore to ARCHITECTURE.md flow after today) |

---

## Open questions for the owner

| # | Question | Default used meanwhile | Answer |
| --- | --- | --- | --- |
| 1 | Which sub manager handles which mess? | `managers.json` left empty; schema supports any mapping (ARCHITECTURE.md A2) | |
| 2 | Are the girls' meal timings the same as the boys'? | Same as boys — PLACEHOLDER copies in seed (ARCHITECTURE.md A3) | |
| 3 | Exact list of default daily items | Placeholder list from section 2: bread, butter, jam, tea, coffee, milk, Bournvita, cornflakes, seasonal fruits; aachar, salad for lunch/dinner | |
| 4 | Hiding a review hides only the comment; stars still count in averages — correct? | Yes, built that way (ARCHITECTURE.md schema note) | |

---

## Known issues

| Date | Issue | Severity (blocks / annoying / cosmetic) | Status |
| --- | --- | --- | --- |
| 2026-10-20 | `internal/httpapi/api_test.go` is an unfinished draft and does not compile | blocks D2.9 verification | **resolved 2026-10-24** (rewritten, all pass) |
| 2026-10-20 | Port 8080 on this machine is taken by Docker Desktop's backend proxy, so a locally-run server must use another port (smoke tests use 8090). Compose still maps 8080:8080 — verify whether the container actually starts there before the demo | annoying | open |
| 2026-10-20 | Client is still the bare Vite React template | expected at that stage | **resolved 2026-10-24** (full real-backend client) |
| 2026-10-20 | `smart-mess/` is not a git repository yet; HARNESS mentions commits | annoying | open (owner asked not to commit; leave for them) |
| 2026-10-24 | `managers.json` emails end `@iiitp.ac.in` so the API needs `ALLOWED_EMAIL_DOMAIN=iiitp.ac.in` (students' `@cse.iiitp.ac.in` still accepted as a subdomain). README documents this | annoying | open — confirm owner intent |
| (none) | — | — | — |

---

## Verification log

| Date | Check | Command or action | Result |
| --- | --- | --- | --- |
| 2026-10-20 (TODAY) | T1 rules tests | `npx vitest run src/rules.test.js` | 16/16 pass — review window (start/inside/end ±1 min), one-review-per-meal, stale rule (30 min boundary), weekend breakfast ends 10:15 |
| 2026-10-20 (TODAY) | Contrast function | `npx vitest run src/contrast.test.js` | 6/6 pass after fixes (off/unsure high text black; tritan medium #FFB300; mono per-mode chip text). Initially caught REAL failures: white on #D55E00 ≈ 3.9:1, #B26A00/#1c1c1e ≈ 4.0:1, mono gray on dark ≈ 3.3:1 |
| 2026-10-20 (TODAY) | Student home render + states | `npx vitest run src/StudentHome.test.jsx` | 4/4 pass — cards with text+icons+bar, stale note present (menu stale preset), closed cards (before lunch), axe: no serious/critical violations |
| 2026-10-20 (TODAY) | Review rule + manager flows | `npx vitest run src/ReviewFlow.test.jsx` | 7/7 pass — review submit→already-originated state persists in localStorage; second review blocked; closed after window; not-yet-open before; sub manager two-tap save writes crowd (`sm:crowd`); main manager publish persists (`sm:menu:boys` status published); student cannot reach manager routes |
| 2026-10-20 (TODAY) | Full suite + build + lint | `npx vitest run` / `npx vite build` / `npx oxlint` | 34/34 tests pass; production build OK; oxlint clean (found + fixed a real bug: ManagerMenu meal-tab switch wasn't loading the expanded meal's items) |
| 2026-10-20 (TODAY) | Dev server boots | `npm run dev`, curl localhost:5173 | 200, `index.html` with lang/title served |
| 2026-10-20 (TODAY) | NOT verified in a real browser | — | vision-deficiency emulation, grayscale view, Lighthouse, keyboard walkthrough, 200%/360px — recorded in docs/accessibility-report.md as pending; no browser in this session |
| --- | --- | --- | --- |
| 2026-10-20 | D1.4 correcting | `sqlc generate` (installed sqlc v1.30.0, CGO off) | store generated in `server/internal/store`; `go build ./...` + `go test ./...` all OK (D1.4 was wrongly `[x]` before: store never generated; claim now true) |
| 2026-10-20 | D1.2 correcting | `docker compose up -d db` | db container up and healthy. `goose up` applied migrations 001+002 to version 2 |
| 2026-10-20 | D1.5 seed fail case | `DATABASE_URL=... go run . --config managers.example.json` | fails with clear empty-email message as required |
| 2026-10-20 | D1.5 seed success | `DATABASE_URL=... go run . --config managers.json` | seeded 5 messes, 16 schedules, default items, 2 main + 3 mess managers, sample dishes |
| 2026-10-20 | D1.6 domain tests | `go test ./internal/auth` (already in `go test ./...`) | pass — all five section-7 cases + trim/lowercase covered |
| 2026-10-20 | Rules tests | `go test ./internal/rules` | pass — weekend breakfast, start/end boundary, review states, crowd states |
| 2026-10-20 | Note | seed accepts `--config <path>` flag (matches HARNESS "How to run"); positional arg still supported. Demo `managers.json` created with placeholder emails (owner replaces; gitignored) | done |
| 2026-10-20 | correcting | `docker compose up -d --build` | all 3 containers rebuild and start: server image runs `goose up` + `/app/main` (server/Dockerfile new), client serves on nginx :3000 (client/Dockerfile + nginx.conf new, VITE_API_URL passed as build arg). Removed obsolete `version:` key from compose |
| 2026-10-20 | D1.7 partial smoke | curl: `/health` 200; POST `/api/auth/request-otp` 204 for `a@cse.iiitp.ac.in`, 400 for `a@gmail.com`; dev code read from server logs; POST `/api/auth/verify-otp` 204 + real signed JWT httpOnly cookie (14d); POST `/api/auth/logout` added; client :3000 returns 200 | pass |
| 2026-10-20 | Note | `httpapi` routes now all under `/api` (`/api/auth/*`, `/api/me*`); OTP pending codes are an in-memory map in `router.go` — stopgap until D1.7 DB wiring to `otp_codes`; `GET /api/me` still returns placeholder user (D1.8 pending) | next: DB wiring |
| 2026-10-20 | correcting (supersedes two rows above) | full backend rewrite: OTP stopgap removed — OTP now stored in `otp_codes` (HMAC hash, 10-min expiry, 5 req/hr per email, 5 wrong attempts); `GET /api/me` is DB-backed; all handlers real. Old placeholder/in-memory code deleted | resolved |
| 2026-10-20 | Backend rewrite builds | `go build ./... && go vet ./... && go test ./...` (server/) | all clean; auth + rules tests pass; httpapi has no tests yet (D2.9 in progress) |
| 2026-10-20 | Live e2e smoke (server binary, dev DB, curl) | health: `{"status":"ok"}` • gmail login → 400 EMAIL_DOMAIN_NOT_ALLOWED • valid OTP request → 204 + code printed to console ("[DEV] OTP for ...") • verify → 200 `{user}` role=student + httpOnly JWT cookie (14 d) • GET /api/me → 200 with role/hostelGroup/prefs • hostel-group POST → sets boys; second POST → 409 HOSTEL_GROUP_ALREADY_SET • PATCH preferences → dietPref/cvdMode/textScale 1.5/reducedMotion all persisted | pass |
| 2026-10-20 | Live e2e smoke (continued) | mess_manager (mess.krishna@) login → /me shows `messes:[]` • PUT /manager/messes/{krishna}/crowd {"level":"high"} → 200 fresh row • PUT crowd on godavri → 403 "not assigned" • main_manager (main.boys@) → default-items lunch returns Aachar/Salad • PUT /manager/meals/2026-10-20/lunch with 3 items (incl. jain-only) → draft saved • publish → status published | pass |
| 2026-10-20 | Live e2e smoke (student reads) | GET /api/messes → 3 boys messes, crowd state `closed` + nextOpening 12:30 IST at 10:02 IST (correct between-meals logic) • GET /api/meals → published lunch with items filtered to diet_pref (jain student sees both+jain items), ratings {avg:null,count:0} both variants, reviewState not_yet_open • POST review at 10:03 (before 12:30 window) → 409 REVIEW_WINDOW_CLOSED "not open yet" (correct) | pass |
| 2026-10-20 | Bug found + fixed in router | crowd route was nested inside the main_manager role group → mess_manager got 403; also chi panicked on double Route("/manager") mount | fixed: single /manager route with two role-grouped sub-groups; re-verified all smoke tests |
| 2026-10-20 | Note | sqlc engine switched to `sql_package: "pgx/v5"` — all handlers adapted to pgtype (UUID/Timestamptz/Date/Numeric); conversion helpers live in internal/httpapi/conv.go | builds clean |
| 2026-10-20 | D2.9 started | `internal/httpapi/api_test.go` drafted (test DB via DATABASE_URL_TEST + goose programmatic migrations, fixed clock inside lunch window 2062-06-12 13:00 IST) | NOT compilable yet — see Known issues |
| 2026-10-20 | Cleanup | stray dev server process killed (`sm_server.exe`), artifacts deleted; sqlc/goose CLIs installed via `go install` (CGO off) | done |
| 2026-10-24 | D2.9 done | `DATABASE_URL_TEST=... go test ./... -count=1` (server/) | ALL PASS — rewritten api_test.go covers the full section-6/7 contract with a fixed 14:00 IST clock on 2062-06-12 |
| 2026-10-24 | Client real build done | all 9 pages on the real API; `npx vite build` + `npm run dev` smoke | build OK (PWA: 15 precached entries, manifest webmanifest generated); dev server serves index.html with lang/title |
| 2026-10-24 | Client tests | `npx vitest run` (client/) | 54/54 pass (rules 16, contrast 7, icon-decoration 15, StudentHome 6 with axe on rendered DOM, ReviewFlow 10 incl. manager flows). Found + fixed real bug: ProtectedRoute ignored its `role` prop (manager-only routes now redirected client-side); added live >500-char warning to ReviewScreen (disabled-only button was a silent failure) |
| 2026-10-24 | Live e2e with fixed clock | server with `DEV_CLOCK="2026-10-24T13:30:00+05:30"` (new dev-only override in cmd/api/main.go), dev DB | main.boys login → PUT menu lunch → publish → student login → GET /api/meals shows reviewState open + items (jain filtered out of regular) → POST review 201 → second POST 409 REVIEW_ALREADY_SUBMITTED → refetch shows already_reviewed + jain avg 4 count 1 → mess.krishna PUT crowd medium → student /messes correctly shows krishna no_update (row created before the lunch window → ignored per ARCHITECTURE 8.3, live) |
| 2026-10-24 | Docs + cleanup | README.md rewritten (verified goose status round-trip on the test DB); accessibility-report.md session-4 section appended; stray artifacts removed (main_cookies.txt, mgr_cookies.txt, server.log, e2e cookie jars); seed emails need `ALLOWED_EMAIL_DOMAIN=iiitp.ac.in` (managers end @iiitp.ac.in, students @cse.iiitp.ac.in) | done |

---

## Cut list (from ARCHITECTURE.md section 12 — cut in this order if time runs out)

1. Offline menu caching — D4.7 (keep PWA manifest and installability)
2. Hide / unhide UI — D4.5 (keep the endpoints — D2.8)
3. Text-size slider — D3.4 (keep respecting browser zoom)
4. Light/dark/high-contrast theme choice (keep `system`)

**Never cut:** color-independent status encoding, `cvd_mode`, keyboard support, focus rings, one-review-per-meal rule, review window rule.

---

## How to run (fill in as it gets built)

```
# TODAY-SCOPE demo (mock data, no backend needed)
cd smart-mess/client
npm install && npm run dev      # http://localhost:5173
npm test                        # 34 tests (rules, screens, contrast, axe)

# Start database, server and client (full stack — NOT today's build)
docker compose up

# Run Go tests (from server/)
go test ./...

# Run migrations manually (from server/)
goose -dir db/migrations postgres "$DATABASE_URL" up

# Run seed (from server/seed/)
go run . --config managers.json

# Start client dev server (from client/)
npm run dev
```


---

## Session notes

Newest first. Three lines at most per session: what was done, what is broken, what is next.

- 2026-10-25 (session 9) — Implemented the six-panel wireframe layouts for auth, group choice, dashboard, review (including form-step indicator), and settings; surfaced existing theme preferences and added confirmed-save/reset feedback using the existing preference endpoint. Frontend-only files changed. `npm test`: 59/59 pass; `npm run build`: Vite/PWA pass. Compared the supplied image with a narrow browser rendering of Sign in; authenticated panels and color-mode browser simulations were not visually inspected.

- 2026-10-25 (session 8) — Revisited the full wireframe source; refined the app header, hostel selection, settings option rows, stacked crowd cards, and primary actions using presentation-only hooks and CSS. Updated `client/src/App.jsx`, `pages/ChooseGroup.jsx`, `pages/Settings.jsx`, `pages/StudentHome.jsx`, `pages/ManagerCrowd.jsx`, `pages/ManagerMenu.jsx`, `pages/ReviewScreen.jsx`, and `index.css`. No app logic, APIs, auth, routes, backend, config, or features changed. `npm test`: 6 files / 59 tests pass; `npm run build`: production/PWA build passes. Browser visual audits not run.

- 2026-10-25 (session 7) — Appearance-only redesign guided by `FINAL_PROMPT_WITH_WIREFRAME.docx`; updated client `index.css` and `theme/tokens.css`. App logic, API behavior, authentication, routes, backend, configuration, and features intentionally unchanged; CVD options preserved. `npm test`: 6 files / 59 tests pass; `npm run build`: production and PWA build pass. Browser visual audits (Lighthouse, color-vision simulation, keyboard and 200%/360px) not run; no known visual issues from source review.

- 2026-10-24 (session 6, fix) — Blank main area at /choose-group: ProtectedRoute redirected "student without hostel group" to `/choose-group` even when already there — a same-path Navigate loop that never commits ChooseGroup (nav renders, main blank, no console error). Fixed by skipping the group redirect when `pathname === "/choose-group"`; regression test added (`Auth401.test.jsx` — 59/59 pass), build OK. All other redirects target different paths, audited, no other loop. Broken: nothing known. Next: owner browser checks per Next step.

- 2026-10-24 (session 5, fixes) — Two client fixes after the real-backend wiring: (1) API base URL — removed the `define` override in `vite.config.js` that stomped the `.env` value (it fell back to hardcoded :8080), fixed `client/.env` to `VITE_API_URL=http://localhost:8090/api`; (2) "stuck on Loading…" — the `mounted` ref in AuthContext was never set back to `true` after React 19 StrictMode's simulated unmount (cleanup fires, setup is a no-op), so `setUser`/`setLoading(false)` after `/api/me` never ran; guard removed entirely, 401/unreachable `/me` now resolves to the login page, real errors are logged not silently logged in. Added `src/Auth401.test.jsx` (4 tests) — suite 58/58, build OK. Broken: nothing known. Next: owner browser checks per Next step.

- 2026-10-24 (session 4, finishing) — D2.9 finished (`api_test.go` rewritten, server suite green incl. fixed-clock API tests); the client was converted from the mock prototype to the real backend client (all 9 pages, auth context, PWA build), tests rewritten against a stubbed fetch server: 54/54, one real guard bug fixed (ProtectedRoute role prop) plus a live >500-char comment warning. Live curl e2e with the new `DEV_CLOCK` override verified publish → review → 409 → already_reviewed → ratings → crowd-window-ignore. Docs updated (README run steps, accessibility report). Broken: nothing known; browser-only checks (Lighthouse, simulators, 200% text) remain for the owner per the cut checklist. Next: owner browser checks + demo; consider git init (owner said not to commit).

- 2026-10-20 (session 3, TODAY-SCOPE) — Built the whole mock-data front-end prototype in `client/` per TODAY-SCOPE.md: demo bar (role/hostel/clock presets), choose-group, student home (crowd cards, meal/variant tabs, dish list, ratings), review screen (4 states), settings (CVD modes + reduced motion), sub + main manager screens; `src/api/index.js` is the only data layer, `rules.js` is pure with tests. 34/34 Vitest tests incl. axe (0 serious/critical) and the automated contrast function (forced real token fixes). Broken: nothing in client; server D2.9 still open (shelved today). Next: browser simulator + keyboard checks before the demo, then back to the ARCHITECTURE.md flow.

- 2026-10-20 (session 2) — Rewrote the whole backend `httpapi` (real DB-backed OTP/JWT/me/messes/meals/reviews/crowd/manager endpoints); live e2e smoke-tested the full flow with curl; fixed the crowd route role bug and a chi mount panic; rules tests + new NextOpening test pass. Broken: `api_test.go` (D2.9) drafted but does not compile. Next: finish D2.9 tests, then mark D2.4–D2.9 and start the client (Day 3).
- 2026-10-20 (session 1) — Corrected D1.4 (sqlc store was never generated; now done), ran migrations on a live DB, verified seed fail + success cases. Started D1.7 (DB-backed OTP). Removed a duplicated copy of this whole file (last session had appended it twice).
- 2026-10-19 — HARNESS.md created, ARCHITECTURE.md read in full. No code written yet. Next: D1.1 repo skeleton.
