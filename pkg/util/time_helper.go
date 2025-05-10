package util

import (
	"database/sql"
	"time"
)

// NullTimeToPointer converts sql.NullTime to *time.Time
func NullTimeToPointer(t sql.NullTime) *time.Time {
    if t.Valid {
        v := t.Time
        return &v
    }
    return nil
}

// PointerToNullTime converts *time.Time to sql.NullTime
func PointerToNullTime(t *time.Time) sql.NullTime {
    if t == nil {
        return sql.NullTime{Valid: false}
    }
    return sql.NullTime{Time: *t, Valid: true}
}