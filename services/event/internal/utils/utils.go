package utils

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ParsedUUID(id string) pgtype.UUID {
	return pgtype.UUID{
		Bytes: uuid.MustParse(id),
		Valid: true,
	}
}
