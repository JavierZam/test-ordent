package util

import (
	"database/sql"
	"time"
)

func NullTimeToPointer(t sql.NullTime) *time.Time {
    if t.Valid {
        v := t.Time
        return &v
    }
    return nil
}

func PointerToNullTime(t *time.Time) sql.NullTime {
    if t == nil {
        return sql.NullTime{Valid: false}
    }
    return sql.NullTime{Time: *t, Valid: true}
}