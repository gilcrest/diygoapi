package uuid

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var Nil UUID // empty UUID, all zeros

// UUID is a type alias for a Google UUID
type UUID uuid.UUID

// New returns a new UUID
func New() UUID {
	return UUID(uuid.New())
}

// PgxUUID converts a google UUID to a pgx UUID
func (u UUID) PgxUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: u, Valid: true}
}
