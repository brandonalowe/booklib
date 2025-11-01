package models

import "time"

type Lending struct {
	ID               int        `json:"id"`
	BookID           int        `json:"book_id"`
	UserID           int        `json:"user_id"`
	LentTo           string     `json:"lent_to"`
	LentAt           time.Time  `json:"lent_at"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	ReturnedAt       *time.Time `json:"returned_at,omitempty"`
	LastReminderSent *time.Time `json:"last_reminder_sent,omitempty"`
}

type LendingWithBook struct {
	ID      int        `json:"id"`
	BookID  int        `json:"book_id"`
	UserID  int        `json:"user_id"`
	LentTo  string     `json:"lent_to"`
	LentAt  time.Time  `json:"lent_at"`
	DueDate *time.Time `json:"due_date,omitempty"`
	Book    *Book      `json:"book"`
}

type CreateLendingRequest struct {
	BookID  int     `json:"book_id"`
	LentTo  string  `json:"lent_to"`
	DueDate *string `json:"due_date,omitempty"`
}

// ReminderLending includes user email for sending reminders
type ReminderLending struct {
	LendingID        int
	UserID           int
	UserEmail        string
	BookTitle        string
	BookAuthor       string
	LentTo           string
	DueDate          time.Time
	LentAt           time.Time
	LastReminderSent *time.Time
}
