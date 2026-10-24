package httpapi

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// uid converts a google/uuid into sqlc's pgtype.UUID.
func uid(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// fromUUID converts sqlc's pgtype.UUID back to google/uuid (zero when null).
func fromUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.UUID{}
	}
	return u.Bytes
}

// tstamp converts a time.Time into pgtype.Timestamptz.
func tstamp(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// fromTstamp converts pgtype.Timestamptz to time.Time (zero when null).
func fromTstamp(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

// fromDate converts pgtype.Date to time.Time (zero when null).
func fromDate(d pgtype.Date) time.Time {
	if !d.Valid {
		return time.Time{}
	}
	return d.Time
}

// toPgDate converts a time.Time to pgtype.Date (UTC midnight).
func toPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

// maskedEmail shows only the first characters before the domain (e.g. 112•••••••@cse.iiitp.ac.in).
func maskedEmail(email string) string {
	at := strings.IndexByte(email, '@')
	if at < 0 {
		return "student"
	}
	local := email[:at]
	if len(local) > 3 {
		local = local[:3] + "••••••••"
	}
	return local + email[at:]
}

// derefBool returns *bool value or false when nil.
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// derefInt32 returns *int32 value or 0 when nil.
func derefInt32(n *int32) int32 {
	if n == nil {
		return 0
	}
	return *n
}
