package httpapi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// parsedID is a parsed URL id.
type parsedID = uuid.UUID

// parseIDParam parses a chi URL parameter as a UUID.
func parseIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, name))
}

// mealIDNull wraps a uuid into a uuid.NullUUID.
func mealIDNull(id uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{UUID: id, Valid: true}
}

// uuidNull is an alias of mealIDNull for readability at call sites.
func uuidNull(id uuid.UUID) uuid.NullUUID {
	return mealIDNull(id)
}

// isUniqueViolation reports whether the error is a Postgres unique constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
