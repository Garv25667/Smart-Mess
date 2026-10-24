package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"smartmess/internal/auth"
	"smartmess/internal/clock"
	"smartmess/internal/config"
	"smartmess/internal/rules"
	"smartmess/internal/store"
)

// Test date far in the future, a weekday, so nothing collides with real data.
// Clock is fixed to 14:00 (inside the lunch window 12:30–14:30). 2062-06-12 is a
// Monday. 14:00 keeps both a fresh (5 min old) and a stale (45 min old) crowd
// row inside the window, so all crowd states are testable.
var (
	testDate = time.Date(2062, 6, 12, 0, 0, 0, 0, rules.KolkataLoc)
	testNow  = time.Date(2062, 6, 12, 14, 0, 0, 0, rules.KolkataLoc)
	testTomorrow = time.Date(2062, 6, 13, 0, 0, 0, 0, rules.KolkataLoc)
)

type testEnv struct {
	pool   *pgxpool.Pool
	router chi.Router

	student, studentGirl, studentFresh, mainBoy, mainGirl, messKrishna, messGodavri uuid.UUID

	mealID    uuid.UUID // boys lunch on testDate, published
	girlMeal  uuid.UUID // girls lunch on testDate, published
	draftMeal uuid.UUID // boys snacks on testDate, draft
}

var env *testEnv

func setupEnv(t *testing.T) *testEnv {
	t.Helper()
	if env != nil {
		return env
	}
	dsn := os.Getenv("DATABASE_URL_TEST")
	if dsn == "" {
		t.Skip("DATABASE_URL_TEST not set; API tests need a separate database")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("test db connect: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("test db ping: %v", err)
	}
	// No pool.Close cleanup: env is a package-level singleton shared by all tests.

	sqldb, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("stdlib open: %v", err)
	}
	defer sqldb.Close()
	if err := goose.UpContext(ctx, sqldb, "../../db/migrations"); err != nil {
		t.Fatalf("test migrations: %v", err)
	}

	var messes int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM messes").Scan(&messes); err != nil {
		t.Fatalf("check messes: %v", err)
	}
	if messes < 5 {
		t.Fatalf("test database not seeded (messes=%d): run the seed script against DATABASE_URL_TEST", messes)
	}

	env = &testEnv{pool: pool}
	env.buildRouter()
	env.resetData(ctx, t)
	env.setupUsers(ctx, t)
	env.setupMeals(ctx, t)
	return env
}

func (e *testEnv) buildRouter() {
	app := &App{
		Cfg: &config.Config{
			JWTSecret:          "test-jwt-secret",
			OTPSecret:          "test-otp-secret",
			ClientOrigin:       "http://localhost:5173",
			AllowedEmailDomain: "iiitp.ac.in",
			AppEnv:             "development",
		},
		Pool:   e.pool,
		Q:      store.New(e.pool),
		Mailer: nopSender{},
		Clock:  clock.Fixed(testNow),
	}
	r := chi.NewRouter()
	Register(r, app)
	e.router = r
}

type nopSender struct{}

func (nopSender) SendOTP(to, code string) error { return nil }

// resetData removes rows the tests create so reruns are clean.
func (e *testEnv) resetData(ctx context.Context, t *testing.T) {
	t.Helper()
	stmts := []string{
		`DELETE FROM reviews WHERE meal_id IN (SELECT id FROM meals WHERE service_date IN ('2062-06-12','2062-06-13'))`,
		`DELETE FROM crowd_status WHERE mess_id IN (SELECT id FROM messes WHERE code IN ('krishna','brahmaputra'))`,
		`DELETE FROM meal_items WHERE meal_id IN (SELECT id FROM meals WHERE service_date IN ('2062-06-12','2062-06-13'))`,
		`DELETE FROM meals WHERE service_date IN ('2062-06-12','2062-06-13')`,
	}
	for _, s := range stmts {
		if _, err := e.pool.Exec(ctx, s); err != nil {
			t.Fatalf("reset: %v", err)
		}
	}
}

func (e *testEnv) setupUsers(ctx context.Context, t *testing.T) {
	t.Helper()
	rows := []struct {
		email string
		role  string
		group string
	}{
		{"apitest-student@iiitp.ac.in", "student", "boys"},
		{"apitest-student-girls@iiitp.ac.in", "student", "girls"},
		{"apitest-fresh@iiitp.ac.in", "student", ""}, // group not chosen yet
		{"apitest-main-boys@iiitp.ac.in", "main_manager", "boys"},
		{"apitest-main-girls@iiitp.ac.in", "main_manager", "girls"},
		{"apitest-mess-krishna@iiitp.ac.in", "mess_manager", ""},
		{"apitest-mess-godavri@iiitp.ac.in", "mess_manager", ""},
	}
	for _, u := range rows {
		var sql string
		var args []any
		if u.group != "" {
			sql = `INSERT INTO users (email, role, hostel_group) VALUES ($1, $2, $3)
				ON CONFLICT (email) DO UPDATE SET role = $2, hostel_group = $3, diet_pref = 'regular', cvd_mode = 'off', text_scale = 1.0, theme = 'system', reduced_motion = false`
			args = []any{u.email, u.role, u.group}
		} else {
			sql = `INSERT INTO users (email, role, hostel_group) VALUES ($1, $2, NULL)
				ON CONFLICT (email) DO UPDATE SET role = $2, hostel_group = NULL, diet_pref = 'regular', cvd_mode = 'off', text_scale = 1.0, theme = 'system', reduced_motion = false`
			args = []any{u.email, u.role}
		}
		if _, err := e.pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("seed user %s: %v", u.email, err)
		}
	}
	getID := func(email string) uuid.UUID {
		var id uuid.UUID
		if err := e.pool.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&id); err != nil {
			t.Fatalf("get user %s: %v", email, err)
		}
		return id
	}
	e.student = getID("apitest-student@iiitp.ac.in")
	e.studentGirl = getID("apitest-student-girls@iiitp.ac.in")
	e.studentFresh = getID("apitest-fresh@iiitp.ac.in")
	e.mainBoy = getID("apitest-main-boys@iiitp.ac.in")
	e.mainGirl = getID("apitest-main-girls@iiitp.ac.in")
	e.messKrishna = getID("apitest-mess-krishna@iiitp.ac.in")
	e.messGodavri = getID("apitest-mess-godavri@iiitp.ac.in")

	// Sub manager assignments: krishna manager has krishna only, godavri manager has godavri only.
	messID := func(code string) uuid.UUID {
		var id uuid.UUID
		if err := e.pool.QueryRow(ctx, "SELECT id FROM messes WHERE code = $1", code).Scan(&id); err != nil {
			t.Fatalf("get mess %s: %v", code, err)
		}
		return id
	}
	if _, err := e.pool.Exec(ctx,
		`DELETE FROM manager_messes WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'apitest-mess-%')`); err != nil {
		t.Fatal(err)
	}
	assign := func(user, mess uuid.UUID) {
		if _, err := e.pool.Exec(ctx,
			`INSERT INTO manager_messes (user_id, mess_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, user, mess); err != nil {
			t.Fatal(err)
		}
	}
	assign(e.messKrishna, messID("krishna"))
	assign(e.messGodavri, messID("godavri"))
}

func (e *testEnv) setupMeals(ctx context.Context, t *testing.T) {
	t.Helper()
	meal := func(group, mealType, status string) uuid.UUID {
		var id uuid.UUID
		pub := any(nil)
		if status == "published" {
			pub = &testNow
		}
		if err := e.pool.QueryRow(ctx, `
			INSERT INTO meals (hostel_group, service_date, meal_type, status, published_at)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			group, testDate.Format("2006-01-02"), mealType, status, pub,
		).Scan(&id); err != nil {
			t.Fatalf("meal %s/%s: %v", group, mealType, err)
		}
		return id
	}
	e.mealID = meal("boys", "lunch", "published")
	e.girlMeal = meal("girls", "lunch", "published")
	e.draftMeal = meal("boys", "snacks", "draft")

	// One dish on the boys lunch so the student menu is not empty.
	var dishID uuid.UUID
	if err := e.pool.QueryRow(ctx,
		`INSERT INTO dishes (hostel_group, name) VALUES ('boys', 'APITEST Dal')
		 ON CONFLICT (hostel_group, lower(name)) DO UPDATE SET name = 'APITEST Dal' RETURNING id`,
	).Scan(&dishID); err != nil {
		t.Fatalf("dish: %v", err)
	}
	if _, err := e.pool.Exec(ctx,
		`INSERT INTO meal_items (meal_id, dish_id, diet_menu, sort_order)
		 SELECT $1, $2, 'both', 0 WHERE NOT EXISTS (SELECT 1 FROM meal_items WHERE meal_id = $1 AND dish_id = $2)`,
		e.mealID, dishID); err != nil {
		t.Fatalf("meal item: %v", err)
	}

	// One existing review on the boys lunch for moderation/list determinism.
	if _, err := e.pool.Exec(ctx, `
		INSERT INTO reviews (meal_id, user_id, variant, stars, comment)
		VALUES ($1, $2, 'regular', 2, 'APITEST comment')
		ON CONFLICT (meal_id, user_id) DO NOTHING`, e.mealID, e.studentGirl); err != nil {
		t.Fatalf("seed review: %v", err)
	}
}

// call performs an authenticated API call with the given user and an optional JSON body.
func (e *testEnv) call(t *testing.T, userID uuid.UUID, method, path string, body any) (int, map[string]any) {
	t.Helper()
	rd := strings.NewReader("")
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode body: %v", err)
		}
		rd = strings.NewReader(string(b))
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Content-Type", "application/json")
	if userID != uuid.Nil {
		token, err := auth.CreateJWT(userID.String(), "student", "apitest@iiitp.ac.in", "test-jwt-secret", time.Hour)
		if err != nil {
			t.Fatalf("jwt: %v", err)
		}
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func errCode(t *testing.T, body map[string]any) string {
	t.Helper()
	if body == nil {
		return ""
	}
	inner, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %v", body)
	}
	code, _ := inner["code"].(string)
	return code
}

func (e *testEnv) messCode(t *testing.T, code string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := e.pool.QueryRow(context.Background(), "SELECT id FROM messes WHERE code = $1", code).Scan(&id); err != nil {
		t.Fatalf("mess %s: %v", code, err)
	}
	return id
}

// --- Tests ---

func TestHealthAndUnauthenticated(t *testing.T) {
	env := setupEnv(t)
	code, _ := env.call(t, uuid.Nil, "GET", "/api/health", nil)
	if code != http.StatusOK {
		t.Fatalf("health: got %d", code)
	}
	code, body := env.call(t, uuid.Nil, "GET", "/api/me", nil)
	if code != http.StatusUnauthorized || errCode(t, body) != "UNAUTHENTICATED" {
		t.Fatalf("unauthenticated: got %d %v", code, body)
	}
}

func TestMeShowsPreferences(t *testing.T) {
	env := setupEnv(t)
	code, body := env.call(t, env.student, "GET", "/api/me", nil)
	if code != http.StatusOK {
		t.Fatalf("/me: %d %v", code, body["error"])
	}
	user := body["user"].(map[string]any)
	if user["role"] != "student" || user["hostelGroup"] != "boys" {
		t.Fatalf("/me user: %v", user)
	}

	code, body = env.call(t, env.student, "PATCH", "/api/me/preferences",
		map[string]any{"cvdMode": "deutan", "textScale": 1.5, "reducedMotion": true, "dietPref": "jain", "theme": "dark"})
	if code != http.StatusOK {
		t.Fatalf("preferences: %d %v", code, body["error"])
	}
	code, body = env.call(t, env.student, "GET", "/api/me", nil)
	user = body["user"].(map[string]any)
	if user["cvdMode"] != "deutan" || user["textScale"] != 1.5 || user["reducedMotion"] != true || user["dietPref"] != "jain" {
		t.Fatalf("preferences not persisted: %v", user)
	}

	code, body = env.call(t, env.student, "PATCH", "/api/me/preferences",
		map[string]any{"cvdMode": "rainbow"})
	if code != http.StatusBadRequest || errCode(t, body) != "VALIDATION_ERROR" {
		t.Fatalf("invalid cvd: %d %v", code, body)
	}
	code, body = env.call(t, env.student, "PATCH", "/api/me/preferences",
		map[string]any{"textScale": 2.5})
	if code != http.StatusBadRequest || errCode(t, body) != "VALIDATION_ERROR" {
		t.Fatalf("invalid scale: %d %v", code, body)
	}

	env.call(t, env.student, "PATCH", "/api/me/preferences",
		map[string]any{"cvdMode": "off", "textScale": 1.0, "reducedMotion": false, "dietPref": "regular", "theme": "system"})
}

func TestHostelGroupSetOnce(t *testing.T) {
	env := setupEnv(t)
	code, body := env.call(t, env.studentFresh, "GET", "/api/meals", nil)
	if code != http.StatusForbidden {
		t.Fatalf("no-group meals: %d %v", code, body)
	}
	code, body = env.call(t, env.studentFresh, "POST", "/api/me/hostel-group", map[string]any{"hostelGroup": "girls"})
	if code != http.StatusOK {
		t.Fatalf("set group: %d %v", code, body["error"])
	}
	code, body = env.call(t, env.studentFresh, "POST", "/api/me/hostel-group", map[string]any{"hostelGroup": "boys"})
	if code != http.StatusConflict || errCode(t, body) != "HOSTEL_GROUP_ALREADY_SET" {
		t.Fatalf("second set: %d %v", code, body)
	}
}

func TestMessesCrowdStates(t *testing.T) {
	env := setupEnv(t)
	// No crowd rows for the fixed now: every boys mess is "no_update" inside lunch.
	code, body := env.call(t, env.student, "GET", "/api/messes", nil)
	if code != http.StatusOK {
		t.Fatalf("messes: %d %v", code, body["error"])
	}
	messes := body["messes"].([]any)
	if len(messes) != 3 {
		t.Fatalf("expected 3 boys messes, got %d", len(messes))
	}
	for _, m := range messes {
		card := m.(map[string]any)
		crowd := card["crowd"].(map[string]any)
		if crowd["state"] != "no_update" {
			t.Fatalf("mess %v: expected no_update, got %v", card["code"], crowd["state"])
		}
	}

	// Fresh: insert a row 5 minutes before the fixed now directly (the server
	// computes state from the newest row, which is what this exercises).
	krishna := env.messCode(t, "krishna")
	if _, err := env.pool.Exec(context.Background(),
		`INSERT INTO crowd_status (mess_id, level, updated_by, updated_at) VALUES ($1, 'medium', $2, $3)`,
		krishna, env.messKrishna, testNow.Add(-5*time.Minute)); err != nil {
		t.Fatal(err)
	}
	code, body = env.call(t, env.student, "GET", "/api/messes", nil)
	for _, m := range body["messes"].([]any) {
		card := m.(map[string]any)
		if card["code"] == "krishna" {
			crowd := card["crowd"].(map[string]any)
			if crowd["state"] != "fresh" || crowd["level"] != "medium" {
				t.Fatalf("krishna: expected fresh/medium, got %v", crowd)
			}
		}
	}

	// Stale: a 45-minute-old row.
	brah := env.messCode(t, "brahmaputra")
	if _, err := env.pool.Exec(context.Background(),
		`INSERT INTO crowd_status (mess_id, level, updated_by, updated_at) VALUES ($1, 'high', $2, $3)`,
		brah, env.messGodavri, testNow.Add(-45*time.Minute)); err != nil {
		t.Fatal(err)
	}
	code, body = env.call(t, env.student, "GET", "/api/messes", nil)
	for _, m := range body["messes"].([]any) {
		card := m.(map[string]any)
		if card["code"] == "brahmaputra" {
			crowd := card["crowd"].(map[string]any)
			if crowd["state"] != "stale" {
				t.Fatalf("brahmaputra: expected stale, got %v", crowd)
			}
		}
	}
}

func TestReviewRules(t *testing.T) {
	env := setupEnv(t)
	path := fmt.Sprintf("/api/meals/%s/reviews", env.mealID)

	// Validation errors (section 6 checks).
	for _, body := range []map[string]any{
		{"variant": "regular", "stars": 0},
		{"variant": "regular", "stars": 6},
		{"variant": "regular", "stars": 3.5},
		{"variant": "veg", "stars": 4},
		{"variant": "regular", "stars": 4, "comment": strings.Repeat("x", 501)},
	} {
		code, resp := env.call(t, env.student, "POST", path, body)
		if code != http.StatusBadRequest || errCode(t, resp) != "VALIDATION_ERROR" {
			t.Fatalf("expected VALIDATION_ERROR for %v, got %d %v", body, code, resp)
		}
	}

	// The girls student cannot review a boys meal.
	code, resp := env.call(t, env.studentGirl, "POST", path, map[string]any{"variant": "regular", "stars": 3})
	if code != http.StatusForbidden {
		t.Fatalf("girl on boys meal: %d %v", code, resp)
	}
	// Managers cannot review.
	code, resp = env.call(t, env.mainBoy, "POST", path, map[string]any{"variant": "regular", "stars": 3})
	if code != http.StatusForbidden {
		t.Fatalf("manager review: %d %v", code, resp)
	}
	// Draft meal -> MEAL_NOT_PUBLISHED (even though it is lunch time, snacks is draft).
	draftPath := fmt.Sprintf("/api/meals/%s/reviews", env.draftMeal)
	code, resp = env.call(t, env.student, "POST", draftPath, map[string]any{"variant": "regular", "stars": 3})
	if code != http.StatusConflict || errCode(t, resp) != "MEAL_NOT_PUBLISHED" {
		t.Fatalf("draft meal: %d %v", code, resp)
	}

	// Valid submission at 13:00 inside the lunch window.
	code, resp = env.call(t, env.student, "POST", path, map[string]any{"variant": "regular", "stars": 4, "comment": "Nice"})
	if code != http.StatusCreated {
		t.Fatalf("submit: %d %v", code, resp["error"])
	}
	// Second submission -> REVIEW_ALREADY_SUBMITTED (unique constraint is the last guard).
	code, resp = env.call(t, env.student, "POST", path, map[string]any{"variant": "jain", "stars": 2})
	if code != http.StatusConflict || errCode(t, resp) != "REVIEW_ALREADY_SUBMITTED" {
		t.Fatalf("second submission: %d %v", code, resp)
	}
	// The review now shows in the visible list.
	code, resp = env.call(t, env.student, "GET", path+"?variant=regular", nil)
	if code != http.StatusOK {
		t.Fatalf("student list: %d %v", code, resp["error"])
	}
	reviews := resp["reviews"].([]any)
	if len(reviews) != 2 {
		t.Fatalf("expected 2 visible reviews (seeded girls + new), got %d", len(reviews))
	}
}

func TestStudentReadsScopedToGroup(t *testing.T) {
	env := setupEnv(t)
	// The boys student sees exactly one published meal (lunch; snacks is draft).
	code, body := env.call(t, env.student, "GET", "/api/meals?date=2062-06-12", nil)
	if code != http.StatusOK {
		t.Fatalf("meals: %d %v", code, body["error"])
	}
	meals := body["meals"].([]any)
	if len(meals) != 1 || meals[0].(map[string]any)["mealType"] != "lunch" {
		t.Fatalf("expected only published boys lunch, got %v", meals)
	}
	// Inside the window: "open", or "already_reviewed" when another test already
	// submitted a review for this student.
	insideWindow := []string{"open", "already_reviewed"}
	state, _ := meals[0].(map[string]any)["reviewState"].(string)
	if state != insideWindow[0] && state != insideWindow[1] {
		t.Fatalf("reviewState inside lunch: %v", meals[0])
	}

	// The girls student gets only the girls meal.
	code, body = env.call(t, env.studentGirl, "GET", "/api/meals?date=2062-06-12", nil)
	meals = body["meals"].([]any)
	if len(meals) != 1 || meals[0].(map[string]any)["mealType"] != "lunch" {
		t.Fatalf("girls meals: %v", meals)
	}

	// The girls student cannot read a boys meal's reviews.
	boysPath := fmt.Sprintf("/api/meals/%s/reviews?variant=regular", env.mealID)
	code, body = env.call(t, env.studentGirl, "GET", boysPath, nil)
	if code != http.StatusForbidden {
		t.Fatalf("girl reads boys reviews: %d %v", code, body)
	}
	// Her own group's reviews are readable.
	girlsPath := fmt.Sprintf("/api/meals/%s/reviews?variant=regular", env.girlMeal)
	code, body = env.call(t, env.studentGirl, "GET", girlsPath, nil)
	if code != http.StatusOK {
		t.Fatalf("girls reviews: %d %v", code, body["error"])
	}
}

func TestCrowdScoping(t *testing.T) {
	env := setupEnv(t)
	krishna := env.messCode(t, "krishna")
	godavri := env.messCode(t, "godavri")

	// Assigned mess: allowed.
	code, body := env.call(t, env.messKrishna, "PUT", fmt.Sprintf("/api/manager/messes/%s/crowd", krishna), map[string]any{"level": "high"})
	if code != http.StatusOK {
		t.Fatalf("assigned mess: %d %v", code, body["error"])
	}
	if c := body["crowd"].(map[string]any); c["level"] != "high" {
		t.Fatalf("crowd body: %v", body)
	}
	// Unassigned mess: FORBIDDEN.
	code, body = env.call(t, env.messKrishna, "PUT", fmt.Sprintf("/api/manager/messes/%s/crowd", godavri), map[string]any{"level": "low"})
	if code != http.StatusForbidden || errCode(t, body) != "FORBIDDEN" {
		t.Fatalf("unassigned mess: %d %v", code, body)
	}
	// Invalid level.
	code, body = env.call(t, env.messKrishna, "PUT", fmt.Sprintf("/api/manager/messes/%s/crowd", krishna), map[string]any{"level": "full"})
	if code != http.StatusBadRequest || errCode(t, body) != "VALIDATION_ERROR" {
		t.Fatalf("invalid level: %d %v", code, body)
	}
	// Students and main managers cannot update crowd.
	for _, u := range []uuid.UUID{env.student, env.mainBoy} {
		code, body = env.call(t, u, "PUT", fmt.Sprintf("/api/manager/messes/%s/crowd", krishna), map[string]any{"level": "low"})
		if code != http.StatusForbidden {
			t.Fatalf("non mess_manager crowd: %d %v", code, body)
		}
	}
}

func TestMenuEditorAndPublish(t *testing.T) {
	env := setupEnv(t)
	tomorrow := "2062-06-13"

	// Default items are seeded.
	code, body := env.call(t, env.mainBoy, "GET", "/api/manager/default-items?mealType=lunch", nil)
	if code != http.StatusOK {
		t.Fatalf("default items: %d %v", code, body["error"])
	}
	if items := body["items"].([]any); len(items) == 0 {
		t.Fatalf("expected seeded default items, got none")
	}

	// Save a draft for tomorrow.
	code, body = env.call(t, env.mainBoy, "PUT", "/api/manager/meals/"+tomorrow+"/lunch",
		map[string]any{"items": []map[string]any{
			{"name": "APITEST Rice", "dietMenu": "both", "sortOrder": 0},
			{"name": "APITEST Jain Curry", "dietMenu": "jain", "sortOrder": 1},
		}})
	if code != http.StatusOK {
		t.Fatalf("save draft: %d %v", code, body["error"])
	}
	saved := body["meal"].(map[string]any)
	if saved["status"] != "draft" || saved["hostelGroup"] != "boys" {
		t.Fatalf("saved meal: %v", saved)
	}

	// Round-trip: manager list shows the meal with both items.
	code, body = env.call(t, env.mainBoy, "GET", "/api/manager/meals?date="+tomorrow, nil)
	meals := body["meals"].([]any)
	if len(meals) != 1 || len(meals[0].(map[string]any)["items"].([]any)) != 2 {
		t.Fatalf("manager meals after save: %v", meals)
	}

	// Publish it.
	code, body = env.call(t, env.mainBoy, "POST", "/api/manager/meals/"+tomorrow+"/lunch/publish", nil)
	if code != http.StatusOK || body["meal"].(map[string]any)["status"] != "published" {
		t.Fatalf("publish: %d %v", code, body["error"])
	}
	// Students now see it and jain-only items are hidden from regular students.
	code, body = env.call(t, env.student, "GET", "/api/meals?date="+tomorrow, nil)
	if code != http.StatusOK || len(body["meals"].([]any)) != 1 {
		t.Fatalf("student sees published meal: %d %v", code, body)
	}
	meal := body["meals"].([]any)[0].(map[string]any)
	items := meal["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("diet filter failed: %v", items)
	}
	if items[0].(map[string]any)["name"] != "APITEST Rice" {
		t.Fatalf("wrong item visible: %v", items)
	}

	// Past date menu edits are rejected.
	code, body = env.call(t, env.mainBoy, "PUT", "/api/manager/meals/2062-06-01/lunch", map[string]any{"items": []any{}})
	if code != http.StatusConflict || errCode(t, body) != "MENU_DATE_IN_PAST" {
		t.Fatalf("past date: %d %v", code, body)
	}

	// Unknown meal type.
	code, body = env.call(t, env.mainBoy, "PUT", "/api/manager/meals/"+tomorrow+"/brunch", map[string]any{"items": []any{}})
	if code != http.StatusBadRequest {
		t.Fatalf("bad meal type: %d %v", code, body)
	}
	// Invalid item.
	code, body = env.call(t, env.mainBoy, "PUT", "/api/manager/meals/"+tomorrow+"/dinner",
		map[string]any{"items": []map[string]any{{"name": "", "dietMenu": "salads", "sortOrder": 0}}})
	if code != http.StatusBadRequest || errCode(t, body) != "VALIDATION_ERROR" {
		t.Fatalf("invalid item: %d %v", code, body)
	}

	// saveMeal always writes to the manager's own group: the girls main manager
	// never touches boys data.
	code, body = env.call(t, env.mainGirl, "PUT", "/api/manager/meals/"+tomorrow+"/dinner",
		map[string]any{"items": []map[string]any{{"name": "APITEST Girls Dish", "dietMenu": "both", "sortOrder": 0}}})
	if code != http.StatusOK || body["meal"].(map[string]any)["hostelGroup"] != "girls" {
		t.Fatalf("girls manager save: %d %v", code, body["error"])
	}
}

func TestModeration(t *testing.T) {
	env := setupEnv(t)
	path := fmt.Sprintf("/api/manager/reviews?mealId=%s", env.mealID)

	// Snapshot counts first (earlier tests may have added reviews; this test
	// asserts on the delta, not absolute numbers).
	code, body := env.call(t, env.student, "GET", fmt.Sprintf("/api/meals/%s/reviews?variant=regular", env.mealID), nil)
	if code != http.StatusOK {
		t.Fatalf("student list: %d %v", code, body["error"])
	}
	visibleBefore := len(body["reviews"].([]any))

	// The boys main manager lists reviews including hidden ones.
	code, body = env.call(t, env.mainBoy, "GET", path, nil)
	if code != http.StatusOK {
		t.Fatalf("manager reviews: %d %v", code, body["error"])
	}
	managerBefore := len(body["reviews"].([]any))
	if managerBefore < visibleBefore {
		t.Fatalf("manager list must include hidden reviews: manager=%d students=%d", managerBefore, visibleBefore)
	}
	var reviewID string
	for _, rv := range body["reviews"].([]any) {
		if r, _ := rv.(map[string]any); r["comment"] == "APITEST comment" {
			reviewID, _ = r["id"].(string)
		}
	}
	if reviewID == "" {
		t.Fatalf("seeded review missing: %v", body)
	}

	// Hide it.
	code, body = env.call(t, env.mainBoy, "POST", fmt.Sprintf("/api/manager/reviews/%s/hide", reviewID), nil)
	if code != http.StatusOK {
		t.Fatalf("hide: %d %v", code, body["error"])
	}
	// The manager list still shows it, marked hidden.
	code, body = env.call(t, env.mainBoy, "GET", path, nil)
	if got := len(body["reviews"].([]any)); got != managerBefore {
		t.Fatalf("manager list changed after hide: %d vs %d", got, managerBefore)
	}
	var hiddenFlag bool
	for _, rv := range body["reviews"].([]any) {
		if r, _ := rv.(map[string]any); r["id"].(string) == reviewID {
			hiddenFlag, _ = r["hidden"].(bool)
		}
	}
	if !hiddenFlag {
		t.Fatalf("hidden flag not set: %v", body)
	}
	// Students no longer see the comment.
	code, body = env.call(t, env.student, "GET", fmt.Sprintf("/api/meals/%s/reviews?variant=regular", env.mealID), nil)
	if code != http.StatusOK {
		t.Fatalf("student list after hide: %d %v", code, body["error"])
	}
	if got := len(body["reviews"].([]any)); got != visibleBefore-1 {
		t.Fatalf("hidden comment leaked to students: %d vs %d-1", got, visibleBefore)
	}
	// Stars still count in averages: the rating count is unchanged by hiding.
	var countBefore float64
	code, body = env.call(t, env.student, "GET", "/api/meals?date=2062-06-12", nil)
	countBefore, _ = body["meals"].([]any)[0].(map[string]any)["ratings"].(map[string]any)["regular"].(map[string]any)["count"].(float64)
	// Unhide it back.
	code, body = env.call(t, env.mainBoy, "POST", fmt.Sprintf("/api/manager/reviews/%s/unhide", reviewID), nil)
	if code != http.StatusOK {
		t.Fatalf("unhide: %d %v", code, body["error"])
	}
	code, body = env.call(t, env.student, "GET", fmt.Sprintf("/api/meals/%s/reviews?variant=regular", env.mealID), nil)
	if got := len(body["reviews"].([]any)); got != visibleBefore {
		t.Fatalf("unhidden comment should show for students: %d vs %d", got, visibleBefore)
	}
	code, body = env.call(t, env.student, "GET", "/api/meals?date=2062-06-12", nil)
	countAfter, _ := body["meals"].([]any)[0].(map[string]any)["ratings"].(map[string]any)["regular"].(map[string]any)["count"].(float64)
	if countAfter != countBefore {
		t.Fatalf("stars count must not change with hide/unhide: %v vs %v", countAfter, countBefore)
	}

	// The girls main manager cannot moderate or list boys reviews.
	code, body = env.call(t, env.mainGirl, "POST", fmt.Sprintf("/api/manager/reviews/%s/hide", reviewID), nil)
	if code != http.StatusForbidden {
		t.Fatalf("cross-group hide: %d %v", code, body)
	}
	code, body = env.call(t, env.mainGirl, "GET", path, nil)
	if code != http.StatusForbidden {
		t.Fatalf("cross-group list: %d %v", code, body)
	}
	// Mess managers cannot reach moderation endpoints.
	code, body = env.call(t, env.messKrishna, "GET", path, nil)
	if code != http.StatusForbidden || errCode(t, body) != "FORBIDDEN" {
		t.Fatalf("mess manager moderation: %d %v", code, body)
	}
	// Students cannot moderate.
	code, body = env.call(t, env.student, "POST", fmt.Sprintf("/api/manager/reviews/%s/hide", reviewID), nil)
	if code != http.StatusForbidden {
		t.Fatalf("student moderation: %d %v", code, body)
	}
}

func TestNotFoundPaths(t *testing.T) {
	env := setupEnv(t)
	unknown := uuid.New()
	code, body := env.call(t, env.student, "GET", fmt.Sprintf("/api/meals/%s/reviews?variant=regular", unknown), nil)
	if code != http.StatusNotFound {
		t.Fatalf("unknown meal: %d %v", code, body)
	}
	code, body = env.call(t, env.messKrishna, "PUT", fmt.Sprintf("/api/manager/messes/%s/crowd", unknown), map[string]any{"level": "low"})
	if code != http.StatusNotFound {
		t.Fatalf("unknown mess: %d %v", code, body)
	}
}
