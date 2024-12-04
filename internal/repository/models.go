package repository

import (
	"time"

	"github.com/google/uuid"
)

type Chore struct {
	ID                int32  `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	DefaultDurationMn int32  `json:"default_duration_mn"`
}

type Task struct {
	ID          uuid.UUID `json:"id"`
	UserID      int32     `json:"user_id"`
	ChoreID     int32     `json:"chore_id"`
	StartedAt   time.Time `json:"started_at"`
	DurationMn  int32     `json:"duration_mn"`
	Description string    `json:"description"`
}

type User struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}
