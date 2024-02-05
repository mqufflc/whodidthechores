package repository

import "time"

type Chore struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   time.Time
	ModifiedAt  time.Time
}

type Task struct {
	Id              int `json:"id"`
	Duration        int `json:"duration"`
	Date            int `json:"date"`
	Done_by_user_id int `json:"done_by_user_id"`
}
