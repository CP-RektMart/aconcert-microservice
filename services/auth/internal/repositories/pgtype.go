package repositories

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ParseUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}
