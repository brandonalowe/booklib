package models

import "time"

type Book struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Author    string     `json:"author"`
	ISBN      string     `json:"isbn"`
	Genre     string     `json:"genre"`
	Read      bool       `json:"read"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}
