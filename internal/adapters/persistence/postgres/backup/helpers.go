package postgres_backup

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func parseDatabaseID(raw string) (int64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("database id is required")
	}

	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse database id %q: %w", raw, err)
	}

	return parsed, nil
}

func parsePublicID(raw string) (pgtype.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return pgtype.UUID{}, nil
	}

	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("parse public id %q: %w", raw, err)
	}

	var bytes [16]byte
	copy(bytes[:], parsed[:])
	return pgtype.UUID{Bytes: bytes, Valid: true}, nil
}

func publicIDToString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}
