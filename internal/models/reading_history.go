package models

import "time"

type ReadingHistory struct {
	ID          int        `json:"id"`
	BookID      int        `json:"book_id"`
	UserID      int        `json:"user_id"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ReadingHistoryWithBook struct {
	ID          int        `json:"id"`
	BookID      int        `json:"book_id"`
	UserID      int        `json:"user_id"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Book        *Book      `json:"book"`
}

type StartReadingRequest struct {
	BookID int `json:"book_id"`
}

type FinishReadingRequest struct {
	ReadingHistoryID int `json:"reading_history_id"`
}
