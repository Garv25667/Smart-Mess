// Package httpapi holds the HTTP router, handlers, middleware and error responses.
package httpapi

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"smartmess/internal/clock"
	"smartmess/internal/config"
	"smartmess/internal/mail"
	"smartmess/internal/store"
)

// App carries everything the handlers need.
type App struct {
	Cfg    *config.Config
	Pool   *pgxpool.Pool
	Q      *store.Queries
	Mailer mail.Sender
	Clock  clock.Clock
}
