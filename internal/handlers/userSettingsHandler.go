package handlers

import (
	"booklib/internal/middleware"
	"booklib/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type UserSettingsHandler struct {
	DB *sql.DB
}

// GetUserSettings retrieves settings for the authenticated user
func (h *UserSettingsHandler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var settings models.UserSettings
	err := h.DB.QueryRow(`
		SELECT 
			user_id,
			email_reminders_enabled,
			email_upcoming_reminders,
			email_overdue_reminders,
			default_lending_days,
			yearly_reading_goal,
			created_at,
			updated_at
		FROM user_settings
		WHERE user_id = ?
	`, userID).Scan(
		&settings.UserID,
		&settings.EmailRemindersEnabled,
		&settings.EmailUpcomingReminders,
		&settings.EmailOverdueReminders,
		&settings.DefaultLendingDays,
		&settings.YearlyReadingGoal,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default settings for user
		settings = h.createDefaultSettings(userID)
	} else if err != nil {
		http.Error(w, `{"error":"Failed to fetch settings"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateUserSettings updates settings for the authenticated user
func (h *UserSettingsHandler) UpdateUserSettings(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req models.UpdateUserSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// First ensure settings exist
	var exists bool
	err := h.DB.QueryRow("SELECT 1 FROM user_settings WHERE user_id = ?", userID).Scan(&exists)
	if err == sql.ErrNoRows {
		// Create default settings
		h.createDefaultSettings(userID)
	}

	// Build dynamic update query
	query := "UPDATE user_settings SET updated_at = ?"
	args := []interface{}{time.Now()}

	if req.EmailRemindersEnabled != nil {
		query += ", email_reminders_enabled = ?"
		args = append(args, *req.EmailRemindersEnabled)
	}
	if req.EmailUpcomingReminders != nil {
		query += ", email_upcoming_reminders = ?"
		args = append(args, *req.EmailUpcomingReminders)
	}
	if req.EmailOverdueReminders != nil {
		query += ", email_overdue_reminders = ?"
		args = append(args, *req.EmailOverdueReminders)
	}
	if req.DefaultLendingDays != nil {
		query += ", default_lending_days = ?"
		args = append(args, *req.DefaultLendingDays)
	}
	if req.YearlyReadingGoal != nil {
		query += ", yearly_reading_goal = ?"
		args = append(args, *req.YearlyReadingGoal)
	}

	query += " WHERE user_id = ?"
	args = append(args, userID)

	_, err = h.DB.Exec(query, args...)
	if err != nil {
		http.Error(w, `{"error":"Failed to update settings"}`, http.StatusInternalServerError)
		return
	}

	// Return updated settings
	h.GetUserSettings(w, r)
}

// createDefaultSettings creates default settings for a user
func (h *UserSettingsHandler) createDefaultSettings(userID int) models.UserSettings {
	now := time.Now()
	settings := models.UserSettings{
		UserID:                 userID,
		EmailRemindersEnabled:  true,
		EmailUpcomingReminders: true,
		EmailOverdueReminders:  true,
		DefaultLendingDays:     14,
		YearlyReadingGoal:      0,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	_, err := h.DB.Exec(`
		INSERT INTO user_settings (
			user_id,
			email_reminders_enabled,
			email_upcoming_reminders,
			email_overdue_reminders,
			default_lending_days,
			yearly_reading_goal,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		settings.UserID,
		settings.EmailRemindersEnabled,
		settings.EmailUpcomingReminders,
		settings.EmailOverdueReminders,
		settings.DefaultLendingDays,
		settings.YearlyReadingGoal,
		settings.CreatedAt,
		settings.UpdatedAt,
	)

	if err != nil {
		// Log error but return defaults anyway
		return settings
	}

	return settings
}
