package models

import "time"

type UserSettings struct {
	UserID                 int       `json:"user_id"`
	EmailRemindersEnabled  bool      `json:"email_reminders_enabled"`
	EmailUpcomingReminders bool      `json:"email_upcoming_reminders"`
	EmailOverdueReminders  bool      `json:"email_overdue_reminders"`
	DefaultLendingDays     int       `json:"default_lending_days"`
	YearlyReadingGoal      int       `json:"yearly_reading_goal"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type UpdateUserSettingsRequest struct {
	EmailRemindersEnabled  *bool `json:"email_reminders_enabled,omitempty"`
	EmailUpcomingReminders *bool `json:"email_upcoming_reminders,omitempty"`
	EmailOverdueReminders  *bool `json:"email_overdue_reminders,omitempty"`
	DefaultLendingDays     *int  `json:"default_lending_days,omitempty"`
	YearlyReadingGoal      *int  `json:"yearly_reading_goal,omitempty"`
}
