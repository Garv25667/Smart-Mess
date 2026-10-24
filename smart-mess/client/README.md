# Smart Mess — client (front-end prototype, mock data)

Front-end prototype with mock data. The full backend (login, database, API) is specified in ARCHITECTURE.md but not built.

## Run

```
npm install
npm run dev
```

Then open http://localhost:5173. The **Demo** bar at the top drives everything: role switcher (Student / Sub manager / Main manager), hostel switcher (Girls / Boys — disabled after the one-time choice), demo clock presets (Before lunch, During lunch, After lunch, Menu stale case) and "Reset demo data".

## Demo walkthrough (under 3 minutes)

1. Student home shows crowd cards with words + person icons + a 3-segment bar, and "Updated X min ago". Switch the demo clock to "Menu stale case" → one card turns stale (dashed outline, "may be outdated" note). "Before lunch" → "Mess closed. Opens at 12:30".
2. Meal tabs (Breakfast…Dinner), Regular/Jain tabs (Jain note: "Prepared without onion and garlic"), dish list with "Same in both menus" tags, per-variant rating summary.
3. Review a meal while the demo clock is inside its window; submit; the already-reviewed state blocks a second review.
4. Settings: color vision mode (`data-cvd` token override, sample strip included) and reduced motion.
5. Role → Sub manager: two taps (Low/Medium/High + Save).
6. Role → Main manager: edit dishes, "Save and Publish"; the student home updates.

## Tests

```
npm test        # Vitest: rules, screens, review rule, contrast function, axe
```

Notes: statuses never use color alone (text + icon + pattern always). Accessibility details: `smart-mess/docs/accessibility-report.md`.
