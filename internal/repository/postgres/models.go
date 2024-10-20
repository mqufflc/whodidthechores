// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package postgres

import (
	"time"

	"github.com/google/uuid"
)

type Chore struct {
	ID                int32
	Name              string
	Description       string
	DefaultDurationMn int32
}

type Task struct {
	ID         uuid.UUID
	UserID     int32
	ChoreID    int32
	StartedAt  time.Time
	DurationMn int32
}

type User struct {
	ID   int32
	Name string
}
