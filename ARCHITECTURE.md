# Smart Mess: Architecture and Build Spec

Single source of truth for building Smart Mess. Read this whole file before writing any code. If something is not specified here, do not invent it: add it to a "Questions" list at the end of your response and use the simplest option that does not touch anything else.

Deadline: the owner submits in 5 days. Section 12 is a 5-day plan with a cut list. Follow it.

---

## 0. Rules for the coding agent

1. Use only the stack and libraries in section 3. Do not add other libraries without asking.
2. Do not change table names, columns, endpoints or enum values from this file. If a change is truly needed, update this file in the same change and say so.
3. **ASSUMPTION** = a default the owner has not confirmed. Implement it exactly as written and keep it in one place so it is easy to change.
4. **PLACEHOLDER** = fake or unknown data to be replaced by the owner. Keep it in `server/seed/` or `server/internal/config/` only, never hard-coded in handlers or components.
5. Never communicate any status by color alone (section 8). This is the main point of the project.
6. Build in the order in section 12. After each day, report what works and what does not.
7. Write tests for the business rules in section 6 before building the screens that use them.
8. Do not claim a feature works unless you ran it. Do not claim an accessibility result unless you ran the check.
9. **Do not seed or hard-code dishes from the printed mess timetables.** The messes do not follow the printed food menu. Only the meal timings from the timetable are used. Dishes are uploaded daily by the main managers.
10. **Do not claim user testing.** No tests with color-blind people are being done. Reports and docs must say the evaluation used simulators and automated checks only.

---

## 1. Product summary

Smart Mess is a mobile-first PWA for **IIIT Pune students only**. It shows each student what is being served today, how crowded each mess in their hostel group is right now, and lets them rate meals. It is designed so students with color vision deficiency (CVD) can use every feature without relying on color.

It is mainly an HCI project: the accessibility design and its evaluation matter as much as the features.

### Hostel groups and messes

There are 5 messes in two hostel groups. All messes in a group serve the same food, at the same times, but each mess has its own crowd level.

| Mess code | Mess name | Hostel group |
| --- | --- | --- |
| `sindhu` | Sindhu | `girls` |
| `indrayani` | Indrayani | `girls` |
| `krishna` | Krishna | `boys` |
| `brahmaputra` | Brahmaputra | `boys` |
| `godavri` | Godavri | `boys` |

### Actors

| Role (`user_role`) | Count | What they do |
| --- | --- | --- |
| `student` | many | View menus and crowd levels, submit one review per meal, change display settings. Picks their hostel group (girls or boys) once. |
| `main_manager` | 2 (one for `girls`, one for `boys`) | Upload the daily menu for their whole group. Moderate reviews for their group. |
| `mess_manager` ("sub manager") | 3 | Update the crowd level of the mess or messes assigned to them. |

Everyone logs in the same way: IIIT Pune college email plus a one-time code. Managers have college emails too; their accounts are created in advance by the seed script and the role is read from the database, never from the request.

There is no admin role in version 1.

### Meals and menus

- Meals per day: `breakfast`, `lunch`, `snacks`, `dinner`.
- Each meal has a Regular and a Jain variant. Jain food is the same food prepared without onion and garlic. Most dishes are therefore `both`; a dish that exists in only one variant is marked `regular` or `jain`.
- The menu is **uploaded daily** per hostel group (all messes of a group share it).

---

## 2. Owner decisions

### Confirmed by the owner

| # | Decision |
| --- | --- |
| C1 | Allowed email domain is `iiitp.ac.in`. Real addresses look like `112400000@cse.iiitp.ac.in`, so department subdomains must be accepted. |
| C2 | A student picks `girls` or `boys` once. It cannot be changed in the app. |
| C3 | No admin above the managers. |
| C4 | 5 logins for managers: 2 main managers (menu upload) and 3 sub managers (crowd levels). All have college emails. |
| C5 | Reviews are **per meal**, not per dish. One review per student per meal. |
| C6 | A student can review only while the meal is being served: from the meal's start time to its end time. After submitting, they cannot review that meal again. |
| C7 | Regular and Jain are reviewed **separately** (the Jain food tastes different without onion and garlic). Averages are shown per variant. |
| C8 | One crowd level per mess, shared by Regular and Jain. |
| C9 | Crowd level is only `low`, `medium` or `high`. **No wait times.** |
| C10 | Menu is uploaded daily. |
| C11 | No user testing with color-blind people. |
| C12 | Submission deadline in 5 days. |

### Assumptions (build these, keep them easy to change)

| # | Assumption | Where it lives |
| --- | --- | --- |
| A1 | A student sees the crowd level of **all messes in their group** (3 for boys, 2 for girls) so they can choose the least crowded one. | `GET /api/messes` |
| A2 | Which sub manager handles which mess is unknown. A sub manager can be assigned any number of messes through `manager_messes`. | `manager_messes`, `seed/managers.json` |
| A3 | The girls' meal timings are the same as the boys'. The girls' printed timetable shows no timings. | `meal_schedules` seed |
| A4 | Weekend means Saturday and Sunday. Only breakfast timing differs on weekends. | `rules.DayKind` |
| A5 | Main managers moderate (hide/unhide) review comments for their group. | `/api/manager/reviews` |
| A6 | A crowd level older than 30 minutes during a meal is shown as stale. | `config.StaleAfterMinutes` |
| A7 | Backend is Go, frontend is React in plain JavaScript. Chosen because the owner knows Go, PostgreSQL, sqlc and goose and must explain the code. | whole repo |
| A8 | Runs locally with Docker Compose for the submission. No hosting target yet. | `docker-compose.yml` |
| A9 | The review form's variant defaults to the student's menu preference (`diet_pref`) and the student can switch it before submitting. | `ReviewForm` |
| A10 | Menu edits are allowed for today and future dates only. Published meals can still be edited (the food changes), and ratings are unaffected. | `PUT /api/manager/meals/...` |
| A11 | Session cookie lasts 14 days. | `config` |

### Meal timings (from the boys' printed timetable; A3 applies to girls)

| Meal | Weekdays | Weekends (Sat, Sun) |
| --- | --- | --- |
| `breakfast` | 08:00 to 09:45 | 08:00 to 10:15 |
| `lunch` | 12:30 to 14:30 | 12:30 to 14:30 |
| `snacks` | 17:00 to 18:00 | 17:00 to 18:00 |
| `dinner` | 19:30 to 21:30 | 19:30 to 21:30 |

All times are `Asia/Kolkata`.

Default items served every day, to prefill the menu editor (managers can remove them): breakfast: Bread, Butter, Jam, Tea, Coffee, Milk, Bournvita, Cornflakes, Seasonal fruits. Lunch and dinner: Aachar, Salad. **PLACEHOLDER:** owner to confirm the exact list.

---

## 3. Tech stack (fixed)

| Layer | Choice |
| --- | --- |
| Backend language | Go (latest stable at build time) |
| HTTP router | `github.com/go-chi/chi/v5`, plus `go-chi/cors` and `go-chi/httprate` |
| Database | PostgreSQL |
| DB access | `github.com/jackc/pgx/v5` with **sqlc** generated queries |
| Migrations | **goose** (SQL files) |
| Auth | `github.com/golang-jwt/jwt/v5` in an httpOnly cookie |
| OTP email | standard library `net/smtp` |
| Validation | hand-written validation functions plus `net/mail` for emails |
| Tests | standard library `testing`, `net/http/httptest` |
| Frontend | React (JavaScript, not TypeScript) with Vite, React Router, Tailwind CSS |
| Client data | `@tanstack/react-query` |
| PWA | `vite-plugin-pwa` |
| Accessibility checks | `@axe-core/playwright`, Lighthouse |
| Local run | Docker Compose: Postgres, server, client |

Use the latest stable version of each at build time and pin exact versions in `go.mod` and `package.json`. Do not guess version numbers from memory. Import `_ "time/tzdata"` in the server so `Asia/Kolkata` loads inside minimal containers.

No WebSockets. The client **polls** the crowd endpoint every 30 seconds while the page is visible.

---

## 4. Repository layout

```
smart-mess/
  ARCHITECTURE.md
  README.md
  docker-compose.yml
  server/
    go.mod
    sqlc.yaml
    cmd/api/main.go
    internal/
      config/          # env loading + validation, constants (StaleAfterMinutes, cookie lifetime)
      clock/           # Clock interface so tests can fix "now"
      rules/           # PURE business rules: day kind, meal windows, review eligibility, crowd state
      auth/            # OTP create/verify, JWT, email-domain check
      mail/            # SMTP sender, console sender for development
      store/           # sqlc generated code + hand-written transactions
      httpapi/         # router, handlers, middleware, error responses
    db/
      migrations/      # goose .sql files
      queries/         # sqlc .sql files
    seed/
      main.go
      managers.example.json
  client/
    index.html
    vite.config.js
    tailwind.config.js
    src/
      main.jsx
      App.jsx
      api/             # fetch wrappers, one file per resource
      auth/            # AuthContext, ProtectedRoute
      theme/           # tokens.css, useDisplaySettings.js
      components/      # CrowdCard, StatusBadge, MealTabs, VariantTabs, MealRatingSummary,
                       # StarInput, ReviewForm, CvdSettings, SkipLink
      pages/           # Login, VerifyCode, ChooseGroup, Home, Settings,
                       # ManagerCrowd, ManagerMenuEditor, ManagerReviews
  docs/
    accessibility-report.md
```

All business rules live in `internal/rules` as pure functions that take `now` and plain data. Handlers call them. This keeps them testable.

---

## 5. Data model (PostgreSQL)

All primary keys are `uuid default gen_random_uuid()`. All timestamps are `timestamptz` in UTC. Name columns exactly as shown.

### Enums

```
hostel_group: 'girls' | 'boys'
user_role:    'student' | 'main_manager' | 'mess_manager'
meal_type:    'breakfast' | 'lunch' | 'snacks' | 'dinner'
day_kind:     'weekday' | 'weekend'
diet_menu:    'regular' | 'jain' | 'both'      -- on meal_items
diet_pref:    'regular' | 'jain'                -- on users and reviews.variant
crowd_level:  'low' | 'medium' | 'high'
meal_status:  'draft' | 'published'
cvd_mode:     'off' | 'unsure' | 'protan' | 'deutan' | 'tritan' | 'mono'
theme_pref:   'system' | 'light' | 'dark' | 'high_contrast'
```

### Tables

```
messes
  id, code text unique, name text, hostel_group hostel_group, created_at

users
  id, email text unique (stored lowercase), role user_role,
  hostel_group hostel_group null,       -- students: null until chosen, then never changed by the app
                                        -- main_manager: required
                                        -- mess_manager: null
  diet_pref diet_pref default 'regular',
  cvd_mode cvd_mode default 'off',
  text_scale numeric default 1.0,       -- allowed range 1.0 to 2.0
  theme theme_pref default 'system',
  reduced_motion boolean default false,
  created_at, last_login_at

manager_messes                          -- which messes a mess_manager updates crowd for
  user_id references users, mess_id references messes, primary key (user_id, mess_id)

otp_codes
  id, email text, code_hash text, expires_at timestamptz,
  attempts int default 0, consumed_at timestamptz null, created_at

meal_schedules
  id, hostel_group, meal_type, day_kind, start_time time, end_time time
  unique (hostel_group, meal_type, day_kind)    -- wall-clock in Asia/Kolkata

default_items
  id, hostel_group, meal_type, name text, sort_order int

dishes
  id, hostel_group, name text, created_at
  unique (hostel_group, lower(name))

meals
  id, hostel_group, service_date date, meal_type, status meal_status default 'draft',
  published_at timestamptz null
  unique (hostel_group, service_date, meal_type)

meal_items
  id, meal_id references meals on delete cascade, dish_id references dishes,
  diet_menu diet_menu default 'both', sort_order int default 0
  unique (meal_id, dish_id)

reviews                                  -- one per student per meal
  id, meal_id references meals, user_id references users,
  variant diet_pref,                     -- which version of the meal the student ate
  stars smallint check (stars between 1 and 5),
  comment text null check (char_length(comment) <= 500),
  hidden boolean default false,
  created_at
  unique (meal_id, user_id)

crowd_status                             -- append-only; newest row per mess is current
  id, mess_id references messes, level crowd_level,
  updated_by references users, updated_at timestamptz default now()
  index (mess_id, updated_at desc)
```

### Rules the schema enforces itself

- One review per student per meal: `unique (meal_id, user_id)`. Do not rely on application code alone.
- Hiding a review (`hidden = true`) hides only its comment from students. Its stars still count in averages. **ASSUMPTION.**
- Reviews and averages are separate per `variant`.

### Seed data

- The 5 messes in section 1.
- `meal_schedules` for both groups from the timings table in section 2 (`weekday` and `weekend` rows; only breakfast differs). Girls' rows are **PLACEHOLDER** copies of the boys' rows (A3).
- `default_items` from section 2 (**PLACEHOLDER**).
- Managers from `server/seed/managers.json` (not committed; `managers.example.json` is committed with empty emails):

```json
{
  "mainManagers": [
    { "email": "", "hostelGroup": "girls" },
    { "email": "", "hostelGroup": "boys" }
  ],
  "messManagers": [
    { "email": "", "messCodes": [] },
    { "email": "", "messCodes": [] },
    { "email": "", "messCodes": [] }
  ]
}
```

  The owner fills in the real emails and which mess each sub manager handles. The seed script must fail with a clear message if an email is empty.
- A few clearly named sample dishes (for example "SAMPLE Dal") for tests only. Do not use the printed timetable dishes.

---

## 6. Business rules

### Time

- The app timezone is `Asia/Kolkata`. A "service date" is the calendar date in that timezone.
- The server is the only source of "now". Every time-dependent response includes `serverTime` (ISO string). The client never uses the device clock to decide what is current, open or stale.
- All time logic lives in `internal/rules`:
  - `DayKind(date)` returns `weekend` for Saturday and Sunday, else `weekday`.
  - `MealWindow(schedule, date)` returns start and end instants.
  - `CurrentMeal(schedules, now)` and `NextMeal(schedules, now)`.
  - `ReviewState(meal, hasReviewed, now)` returns one of `not_yet_open`, `open`, `closed`, `already_reviewed`.
- Test all of them, including the weekend breakfast and a time exactly at start and at end. A meal is open when `start <= now <= end`.

### What students see

- Students only see meals with `status = 'published'` for their own `hostel_group`.
- **Messes:** all messes of the student's group, each with its crowd state (below).
- **Menu:** items where `diet_menu` equals the student's `diet_pref` or is `both`. The Jain tab shows the note "Prepared without onion and garlic".
- **Current meal:** the meal whose window contains now. If none, the next meal today. If none left today, show "Mess closed" and tomorrow's first opening time.

### Crowd state (computed on the server)

For each mess, `state` is one of:

| state | When | Shown |
| --- | --- | --- |
| `fresh` | inside a meal window and newest crowd row is 30 minutes old or less | level |
| `stale` | inside a meal window and newest crowd row is older than 30 minutes | level plus "may be outdated" |
| `no_update` | inside a meal window and there is no crowd row since this meal's start | "No update yet" |
| `closed` | outside every meal window | "Mess closed. Opens at HH:MM" and no level |

The response includes `level`, `updatedAt`, `ageMinutes` (when relevant), `state`, `nextOpening`.

### Reviews

A review submission is allowed only if all are true; otherwise return the matching error code from section 9.

1. The user is a student whose `hostel_group` equals the meal's `hostel_group`.
2. The meal is published.
3. Now is inside the meal window (C6).
4. No review exists for `(meal_id, user_id)`.
5. `stars` is an integer from 1 to 5; `variant` is `regular` or `jain`.
6. `comment`, if present, is at most 500 characters after trimming.

The unique constraint is the last guard. Map its violation to `REVIEW_ALREADY_SUBMITTED`.

### Roles and scoping

- `mess_manager` can change crowd only for messes listed in `manager_messes` for their user.
- `main_manager` can read and change menus and reviews only for their own `hostel_group`.
- The group or mess always comes from the database record of the logged-in user or from the URL, checked against it. Never trust a group sent in a request body.
- Add tests that a manager cannot touch another group or an unassigned mess.

---

## 7. API (REST, JSON, prefix `/api`)

Auth is a signed JWT in an httpOnly, `SameSite=Lax` cookie (`Secure` in production). JSON field names are camelCase. All inputs are validated. All errors use the format in section 9.

### Auth (everyone uses OTP)

| Method | Path | Body | Result |
| --- | --- | --- | --- |
| POST | `/auth/request-otp` | `{ email }` | 204. `EMAIL_DOMAIN_NOT_ALLOWED` if the email is outside the allowed domain. Same response whether or not the user exists. |
| POST | `/auth/verify-otp` | `{ email, code }` | Sets cookie, returns `{ user }`. If no user exists for the email, creates a `student`. If a manager row exists, that role is used. |
| POST | `/auth/logout` | none | 204, clears cookie |

**Email domain rule** (put in `internal/auth`, with tests): lowercase and trim. Accept if the part after `@` equals `ALLOWED_EMAIL_DOMAIN` or ends with `"." + ALLOWED_EMAIL_DOMAIN`. With `ALLOWED_EMAIL_DOMAIN=iiitp.ac.in`: `a@cse.iiitp.ac.in` accepted, `a@iiitp.ac.in` accepted, `a@gmail.com` rejected, `a@eviliiitp.ac.in` rejected, `a@iiitp.ac.in.evil.com` rejected.

OTP rules: 6 digits, stored as an HMAC-SHA256 hash, expires in 10 minutes, max 5 wrong attempts per code, max 5 code requests per email per hour, constant-time comparison. In development, if SMTP is not configured, print the code to the server console.

### Current user and preferences

| Method | Path | Body | Result |
| --- | --- | --- | --- |
| GET | `/me` | none | `{ user }` with role, hostelGroup, dietPref, cvdMode, textScale, theme, reducedMotion, and for mess managers `messes: [...]` |
| PATCH | `/me/preferences` | any of `dietPref`, `cvdMode`, `textScale`, `theme`, `reducedMotion` | `{ user }` |
| POST | `/me/hostel-group` | `{ hostelGroup }` | `{ user }`. Students only. `409 HOSTEL_GROUP_ALREADY_SET` if already chosen. |

### Student reads

| Method | Path | Result |
| --- | --- | --- |
| GET | `/messes` | messes of the student's group: `[{ id, code, name, crowd: { state, level, updatedAt, ageMinutes, nextOpening } }]`, plus `serverTime` |
| GET | `/meals?date=YYYY-MM-DD` | published meals for that date for the student's group (default today), each: `{ id, mealType, startsAt, endsAt, items: [{ id, name, dietMenu }], ratings: { regular: { avg, count }, jain: { avg, count } }, reviewState, myReview }`, plus `serverTime` |
| GET | `/meals/{mealId}/reviews?variant=&limit=20&cursor=` | non-hidden comments with stars, newest first |

### Reviews

| Method | Path | Body | Result |
| --- | --- | --- | --- |
| POST | `/meals/{mealId}/reviews` | `{ variant, stars, comment? }` | 201 `{ review }` |

### Managers

| Method | Path | Who | Body / result |
| --- | --- | --- | --- |
| PUT | `/manager/messes/{messId}/crowd` | `mess_manager` | `{ level }` returns the new crowd row |
| GET | `/manager/meals?date=YYYY-MM-DD` | `main_manager` | draft and published meals for their group |
| GET | `/manager/default-items?mealType=` | `main_manager` | default items for that meal |
| PUT | `/manager/meals/{date}/{mealType}` | `main_manager` | `{ items: [{ name, dietMenu, sortOrder }] }` replaces the meal's items, creates dishes by name if missing, keeps status |
| POST | `/manager/meals/{date}/{mealType}/publish` | `main_manager` | sets `published` |
| GET | `/manager/reviews?mealId=` | `main_manager` | all reviews for that meal including hidden |
| POST | `/manager/reviews/{reviewId}/hide` | `main_manager` | hides the comment |
| POST | `/manager/reviews/{reviewId}/unhide` | `main_manager` | shows it again |

---

## 8. Accessibility design (the core requirement)

### Non-negotiable rules

1. **No status uses color alone.** Each status uses three channels at once: a text label, an icon with a distinct shape, and a pattern or position cue. Color is an extra.
2. Text contrast at least 4.5:1; large text and interface components at least 3:1. Check this in code with a contrast function when defining tokens, not by eye.
3. Every interactive element has a visible focus ring that meets 3:1 against its background.
4. Touch targets at least 44 by 44 CSS pixels.
5. Everything works with keyboard only. A "Skip to main content" link is the first focusable element.
6. Text scales to 200% (`text_scale` up to 2.0) without losing content or horizontal scroll at 360 px width.
7. Respect `prefers-reduced-motion` and the `reduced_motion` setting. Respect `prefers-color-scheme` when `theme = 'system'`.
8. Form fields have visible labels. Errors appear as text under the field with an icon, and focus moves to the first error.
9. Buttons and links always have visible text. No icon-only buttons.

### Status encoding table

| Information | Text | Icon / shape | Extra cue |
| --- | --- | --- | --- |
| Crowd low | "Low crowd" | 1 filled person icon | Bar with 1 of 3 segments filled |
| Crowd medium | "Medium crowd" | 2 person icons | Bar with 2 of 3 segments filled |
| Crowd high | "High crowd" | 3 person icons plus a warning triangle | Bar with 3 of 3 segments filled |
| Stale crowd data | "Updated 42 min ago, may be outdated" | clock icon | Dashed outline on the card |
| No update yet | "No update yet" | question-mark icon | Dashed outline, empty bar |
| Mess closed | "Mess closed. Opens at 12:30" | moon icon | Plain gray card, no bar |
| Regular / Jain | "Regular" and "Jain" written on the tabs and on the rating summary | A different icon per variant | Jain tab note: "Prepared without onion and garlic" |
| Meal rating | Number next to the stars: "4.2 / 5 (31 reviews)" | Filled vs outlined stars | Review count |
| Review not open yet | "Reviews open at 12:30" | clock icon | Disabled form, not hidden |
| Review closed | "Reviews for this meal have closed" | lock icon | Disabled form, not hidden |
| Already reviewed | "You reviewed this meal" | check icon | Disabled form showing their stars |
| Form error | Message text under the field | warning icon | Thick border and focus moves to the field |

### Color vision mode

Setting `cvd_mode` on the user. It changes color tokens only. Text, icons and patterns are identical in every mode.

| Value | Label shown to the student |
| --- | --- |
| `off` | Color vision mode off (default) |
| `unsure` | Not sure which type I have |
| `protan` | Red-weak or red-blind (protanopia / protanomaly) |
| `deutan` | Green-weak or green-blind (deuteranopia / deuteranomaly) |
| `tritan` | Blue-weak (tritanopia / tritanomaly) |
| `mono` | No color vision (achromatopsia) |

Implementation: set `data-cvd="<value>"` on `<html>`; define CSS variables per value in `client/src/theme/tokens.css`. Components use only the variables, never raw hex.

Tokens to define in every mode: `--status-low`, `--status-medium`, `--status-high`, `--status-stale`, `--accent`, plus background and text tokens for each theme.

**Starting values (verify, then tune).** Based on the Okabe-Ito palette and lightness steps. Check them in the Chrome DevTools vision-deficiency emulation and with a contrast function. Do not call them "CVD-safe" in code or docs until that check has been run and recorded.

| Mode | `--status-low` | `--status-medium` | `--status-high` | Idea |
| --- | --- | --- | --- | --- |
| `off`, `unsure` | `#0072B2` | `#E69F00` | `#D55E00` | Okabe-Ito blue, orange, vermillion |
| `protan`, `deutan` | `#0072B2` | `#F0E442` | `#3B1F5C` | Blue / yellow / dark purple: differ in hue and lightness, avoid red-green |
| `tritan` | `#2E7D32` | `#FFB300` | `#B71C1C` | Avoid blue-yellow pairs; differ in lightness |
| `mono` | `#D9D9D9` | `#8C8C8C` | `#1A1A1A` | Three gray steps; patterns get thicker |

Text drawn on a status color picks black or white by computed contrast.

### First-run flow

After the first sign-in: (1) choose Girls or Boys, with a confirmation "This cannot be changed later"; (2) a short screen "Do you have trouble telling colors apart?" with "No, continue" and "Yes, set up color vision mode". "Yes" opens the type picker with the labels above and a sample strip of the three status colors beside each choice. The setting is always reachable from Settings.

---

## 9. Error format

```json
{ "error": { "code": "REVIEW_ALREADY_SUBMITTED", "message": "You already reviewed this meal." } }
```

| Code | HTTP | When |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | invalid input (include `fields`) |
| `EMAIL_DOMAIN_NOT_ALLOWED` | 400 | email outside the allowed domain |
| `OTP_INVALID` | 400 | wrong, expired or used code |
| `OTP_RATE_LIMITED` | 429 | too many requests or attempts |
| `UNAUTHENTICATED` | 401 | no or bad cookie |
| `FORBIDDEN` | 403 | wrong role, other group, unassigned mess |
| `NOT_FOUND` | 404 | unknown id |
| `HOSTEL_GROUP_ALREADY_SET` | 409 | student tries to change group |
| `MEAL_NOT_PUBLISHED` | 409 | review on an unpublished meal |
| `REVIEW_WINDOW_CLOSED` | 409 | review outside the meal window |
| `REVIEW_ALREADY_SUBMITTED` | 409 | second review for the same meal |
| `MENU_DATE_IN_PAST` | 409 | menu edit for a past date |

---

## 10. Screens (client routes)

| Route | Who | Content, top to bottom |
| --- | --- | --- |
| `/login` | public | college email field, "Send code" |
| `/verify` | public | 6-digit code field, "Verify", "Send a new code" |
| `/choose-group` | student, first time | two large buttons Girls and Boys, then a confirm step |
| `/` | student | 1. Crowd cards, one per mess in the group (name, level, last updated). 2. Meal tabs: Breakfast, Lunch, Snacks, Dinner (current or next preselected). 3. Regular / Jain tabs. 4. Dish list for the meal. 5. Rating summary for both variants. 6. Review form, or the matching state message. |
| `/settings` | any | color vision mode, text size, theme, reduced motion, menu preference (Regular or Jain), log out |
| `/manager` | `mess_manager` | their mess or messes with big buttons Low / Medium / High and the last updated time; a switcher if more than one mess |
| `/manager/menu` | `main_manager` | date picker, four meal tabs, dish list editor prefilled with default items, Regular / Jain / Both per dish, "Save draft" and "Publish" |
| `/manager/reviews` | `main_manager` | reviews for a chosen meal with hide / unhide |

The crowd update takes two taps: tap the level, then tap "Save". The screen states that the level is a manual estimate.

PWA: installable, cache the app shell and the last fetched menu. Offline, show "Offline. Showing saved menu from <time>" and hide crowd levels (they would be stale).

---

## 11. Security and configuration

Environment variables (validate at startup, fail fast if missing):

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | Postgres connection string |
| `DATABASE_URL_TEST` | separate database for tests |
| `JWT_SECRET` | signs session cookies |
| `OTP_SECRET` | key for hashing OTP codes |
| `CLIENT_ORIGIN` | allowed CORS origin |
| `ALLOWED_EMAIL_DOMAIN` | `iiitp.ac.in` |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `MAIL_FROM` | OTP email; optional in development |
| `APP_ENV` | `development` or `production` |
| `PORT` | server port |

- Parameterized queries only (sqlc already does this).
- Never log OTP codes in production.
- Rate limit auth routes and the review route.
- Generic messages for auth failures.
- Set security headers (no inline scripts; a content security policy for the built client).

---

## 12. Five-day plan

| Day | Work |
| --- | --- |
| 1 | Repo, Docker Compose, goose migrations, sqlc, seed script. OTP login for everyone, `/me`, `/me/hostel-group`, `/me/preferences`. Email-domain tests. |
| 2 | `internal/rules` with tests (day kind, windows, review state, crowd state). Read APIs. Review submission, crowd update, menu upload and publish, moderation. API tests for every rule in section 6. |
| 3 | Client shell, routing, auth, theme tokens for every `cvd_mode`, settings, login, verify, choose-group, student home with crowd cards, meal tabs and dish list. |
| 4 | Rating summary and review form, manager crowd screen, menu editor, review moderation, PWA manifest and app-shell caching. |
| 5 | Accessibility pass and report, README with run steps, demo seed data, bug fixes. |

### Cut list (cut in this order if time runs out)

1. Offline menu caching (keep the manifest and installability).
2. Hide / unhide UI (keep the endpoints).
3. Text-size slider (keep respecting browser zoom).
4. Light/dark/high-contrast theme choice (keep `system`).

Never cut: color-independent status encoding, `cvd_mode`, keyboard support, focus rings, one-review-per-meal rule, review window rule.

### Definition of done

- All Go tests pass.
- Lighthouse accessibility score 95 or higher on the student home and the manager crowd screen.
- axe reports no serious or critical issues.
- Every status in section 8 is still unambiguous in grayscale and under protanopia, deuteranopia, tritanopia and achromatopsia emulation.
- Keyboard-only walkthrough of every screen works; 200% text and 360 px width work.
- `docs/accessibility-report.md` records each check, its result and the date.

---

## 13. How the accessibility work is reported

The evaluation uses simulators and automated checks only: Chrome DevTools vision-deficiency emulation, grayscale view, a contrast function, axe, Lighthouse, a keyboard-only walkthrough, and a 200% text and 360 px width check.

`docs/accessibility-report.md` must state plainly: "Evaluated with simulators and automated tools. No testing was done with color-blind participants." It must not say "WCAG compliant"; it may say "checked against WCAG 2.2 AA criteria" and list which checks were run.

---

## 14. Out of scope for version 1

Do not build: student crowd reports, wait-time estimates, automatic crowd detection or prediction, per-dish ratings, voice input, push notifications, admin dashboard, multiple languages, native apps, payments, WebSockets, social features, or anything not listed above.
