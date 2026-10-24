package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Manager struct {
	Email       string   `json:"email"`
	HostelGroup string   `json:"hostelGroup,omitempty"`
	MessCodes   []string `json:"messCodes,omitempty"`
}

type ManagersFile struct {
	MainManagers []Manager `json:"mainManagers"`
	MessManagers []Manager `json:"messManagers"`
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("cannot ping database: %v", err)
	}
	log.Println("Connected to database")

	// Load managers JSON. Accepts --config <path> (or a positional path).
	path := "managers.json"
	if len(os.Args) > 2 && os.Args[1] == "--config" {
		path = os.Args[2]
	} else if len(os.Args) > 1 && os.Args[1] != "--config" {
		path = os.Args[1]
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("cannot read managers file %s: %v", path, err)
	}
	var mf ManagersFile
	if err := json.Unmarshal(data, &mf); err != nil {
		log.Fatalf("json parse error: %v", err)
	}

	// Validate emails
	for _, m := range mf.MainManagers {
		if strings.TrimSpace(m.Email) == "" {
			log.Fatalf(`empty main manager email in %s. Fill in mainManagers[].email in the managers file (and messes for mess managers), then run the seed again.`, path)
		}
	}
	for _, m := range mf.MessManagers {
		if strings.TrimSpace(m.Email) == "" {
			log.Fatalf(`empty mess manager email in %s. Fill in messManagers[].email and messCodes in the managers file, then run the seed again.`, path)
		}
	}

	// --- Seed messes ---
	messes := []struct {
		Code        string
		Name        string
		HostelGroup string
	}{
		{"sindhu", "Sindhu", "girls"},
		{"indrayani", "Indrayani", "girls"},
		{"krishna", "Krishna", "boys"},
		{"brahmaputra", "Brahmaputra", "boys"},
		{"godavri", "Godavri", "boys"},
	}

	for _, m := range messes {
		_, err := pool.Exec(ctx,
			`INSERT INTO messes (code, name, hostel_group)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (code) DO NOTHING`,
			m.Code, m.Name, m.HostelGroup)
		if err != nil {
			log.Fatalf("insert mess %s: %v", m.Code, err)
		}
	}
	log.Println("Seeded messes")

	// --- Seed meal schedules ---
	type sched struct {
		Group    string
		Meal     string
		DayKind  string
		Start    string
		End      string
	}
	schedules := []sched{
		// Boys weekday
		{"boys", "breakfast", "weekday", "08:00", "09:45"},
		{"boys", "lunch", "weekday", "12:30", "14:30"},
		{"boys", "snacks", "weekday", "17:00", "18:00"},
		{"boys", "dinner", "weekday", "19:30", "21:30"},
		// Boys weekend
		{"boys", "breakfast", "weekend", "08:00", "10:15"},
		{"boys", "lunch", "weekend", "12:30", "14:30"},
		{"boys", "snacks", "weekend", "17:00", "18:00"},
		{"boys", "dinner", "weekend", "19:30", "21:30"},
		// Girls weekday (PLACEHOLDER: same as boys, A3)
		{"girls", "breakfast", "weekday", "08:00", "09:45"},
		{"girls", "lunch", "weekday", "12:30", "14:30"},
		{"girls", "snacks", "weekday", "17:00", "18:00"},
		{"girls", "dinner", "weekday", "19:30", "21:30"},
		// Girls weekend
		{"girls", "breakfast", "weekend", "08:00", "10:15"},
		{"girls", "lunch", "weekend", "12:30", "14:30"},
		{"girls", "snacks", "weekend", "17:00", "18:00"},
		{"girls", "dinner", "weekend", "19:30", "21:30"},
	}

	for _, s := range schedules {
		_, err := pool.Exec(ctx,
			`INSERT INTO meal_schedules (hostel_group, meal_type, day_kind, start_time, end_time)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (hostel_group, meal_type, day_kind) DO NOTHING`,
			s.Group, s.Meal, s.DayKind, s.Start, s.End)
		if err != nil {
			log.Fatalf("insert schedule %s/%s/%s: %v", s.Group, s.Meal, s.DayKind, err)
		}
	}
	log.Println("Seeded meal schedules")

	// --- Seed default items (PLACEHOLDER) ---
	type defItem struct {
		Group string
		Meal  string
		Name  string
		Order int
	}
	defaultItems := []defItem{
		// Breakfast defaults (both groups)
		{"boys", "breakfast", "Bread", 1},
		{"boys", "breakfast", "Butter", 2},
		{"boys", "breakfast", "Jam", 3},
		{"boys", "breakfast", "Tea", 4},
		{"boys", "breakfast", "Coffee", 5},
		{"boys", "breakfast", "Milk", 6},
		{"boys", "breakfast", "Bournvita", 7},
		{"boys", "breakfast", "Cornflakes", 8},
		{"boys", "breakfast", "Seasonal fruits", 9},
		{"girls", "breakfast", "Bread", 1},
		{"girls", "breakfast", "Butter", 2},
		{"girls", "breakfast", "Jam", 3},
		{"girls", "breakfast", "Tea", 4},
		{"girls", "breakfast", "Coffee", 5},
		{"girls", "breakfast", "Milk", 6},
		{"girls", "breakfast", "Bournvita", 7},
		{"girls", "breakfast", "Cornflakes", 8},
		{"girls", "breakfast", "Seasonal fruits", 9},
		// Lunch/dinner defaults
		{"boys", "lunch", "Aachar", 1},
		{"boys", "lunch", "Salad", 2},
		{"boys", "dinner", "Aachar", 1},
		{"boys", "dinner", "Salad", 2},
		{"girls", "lunch", "Aachar", 1},
		{"girls", "lunch", "Salad", 2},
		{"girls", "dinner", "Aachar", 1},
		{"girls", "dinner", "Salad", 2},
	}

	for _, d := range defaultItems {
		_, err := pool.Exec(ctx,
			`INSERT INTO default_items (hostel_group, meal_type, name, sort_order)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT DO NOTHING`,
			d.Group, d.Meal, d.Name, d.Order)
		if err != nil {
			log.Fatalf("insert default item: %v", err)
		}
	}
	log.Println("Seeded default items")

	// --- Seed managers ---
	for _, m := range mf.MainManagers {
		email := strings.TrimSpace(strings.ToLower(m.Email))
		_, err := pool.Exec(ctx,
			`INSERT INTO users (email, role, hostel_group)
			 VALUES ($1, 'main_manager', $2)
			 ON CONFLICT (email) DO UPDATE SET role = 'main_manager', hostel_group = $2`,
			email, m.HostelGroup)
		if err != nil {
			log.Fatalf("insert main manager %s: %v", email, err)
		}
		log.Printf("Seeded main manager: %s (%s)", email, m.HostelGroup)
	}

	for _, m := range mf.MessManagers {
		email := strings.TrimSpace(strings.ToLower(m.Email))
		var userID string
		err := pool.QueryRow(ctx,
			`INSERT INTO users (email, role)
			 VALUES ($1, 'mess_manager')
			 ON CONFLICT (email) DO UPDATE SET role = 'mess_manager'
			 RETURNING id`,
			email).Scan(&userID)
		if err != nil {
			log.Fatalf("insert mess manager %s: %v", email, err)
		}
		// Assign messes
		for _, code := range m.MessCodes {
			_, err := pool.Exec(ctx,
				`INSERT INTO manager_messes (user_id, mess_id)
				 SELECT $1, id FROM messes WHERE code = $2
				 ON CONFLICT DO NOTHING`,
				userID, code)
			if err != nil {
				log.Fatalf("assign mess %s to %s: %v", code, email, err)
			}
		}
		log.Printf("Seeded mess manager: %s (messes: %v)", email, m.MessCodes)
	}

	// --- Seed sample dishes (for tests only) ---
	sampleDishes := []struct {
		Group string
		Name  string
	}{
		{"boys", "SAMPLE Dal"},
		{"boys", "SAMPLE Rice"},
		{"boys", "SAMPLE Roti"},
		{"boys", "SAMPLE Paneer"},
		{"girls", "SAMPLE Dal"},
		{"girls", "SAMPLE Rice"},
		{"girls", "SAMPLE Roti"},
		{"girls", "SAMPLE Paneer"},
	}

	for _, d := range sampleDishes {
		_, err := pool.Exec(ctx,
			`INSERT INTO dishes (hostel_group, name)
			 VALUES ($1, $2)
			 ON CONFLICT (hostel_group, lower(name)) DO NOTHING`,
			d.Group, d.Name)
		if err != nil {
			log.Fatalf("insert sample dish: %v", err)
		}
	}
	log.Println("Seeded sample dishes")

	fmt.Println("\n✓ Seed completed successfully!")
}
