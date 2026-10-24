package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"smartmess/internal/clock"
	"smartmess/internal/config"
	"smartmess/internal/httpapi"
	"smartmess/internal/mail"
	"smartmess/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connect error: %v", err)
	}
	defer pool.Close()

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	if err := pool.Ping(dbCtx); err != nil {
		log.Fatalf("database ping error: %v", err)
	}

	var sender mail.Sender
	if cfg.SMTPConfigured() {
		sender = &mail.SMTPSender{
			Host: cfg.SMTPHost, Port: cfg.SMTPPort,
			User: cfg.SMTPUser, Pass: cfg.SMTPPass, MailFrom: cfg.MailFrom,
		}
	} else {
		if cfg.IsDev() {
			log.Println("SMTP not configured: OTP codes will be printed to the console (development mode).")
		}
		sender = mail.ConsoleSender{}
	}

	// DEV_CLOCK (dev only): run against a fixed wall clock, RFC3339, for
	// reproducible e2e smoke flows (e.g. a lunch-window scenario).
	appClock := clock.Clock(clock.Real())
	if cfg.IsDev() {
		if v := os.Getenv("DEV_CLOCK"); v != "" {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				log.Printf("DEV_CLOCK override active: %s", v)
				appClock = clock.Fixed(t)
			} else {
				log.Printf("DEV_CLOCK ignored (not RFC3339): %s", v)
			}
		}
	}

	app := &httpapi.App{
		Cfg:    cfg,
		Pool:   pool,
		Q:      store.New(pool),
		Mailer: sender,
		Clock:  appClock,
	}

	r := chi.NewRouter()
	httpapi.Register(r, app)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
