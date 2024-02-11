package repository

import "time"

type ValidationError struct {
	s string
}

func (e ValidationError) Error() string {
	return e.s
}

type Chore struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   time.Time
	ModifiedAt  time.Time
}

type ChoreParams struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (p *ChoreParams) validate() error {
	if p.ID == "" {
		return ValidationError{"missing chore ID"}
	}
	if p.Name == "" {
		return ValidationError{"missing chore name"}
	}
	return nil
}

type Task struct {
	Id              int `json:"id"`
	Duration        int `json:"duration"`
	Date            int `json:"date"`
	Done_by_user_id int `json:"done_by_user_id"`
}
