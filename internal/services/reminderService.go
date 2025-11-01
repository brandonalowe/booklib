package services

import (
	"database/sql"
	"log"
	"time"

	"booklib/internal/models"
)

type ReminderService struct {
	DB           *sql.DB
	EmailService *EmailService
}

func NewReminderService(db *sql.DB, emailService *EmailService) *ReminderService {
	return &ReminderService{
		DB:           db,
		EmailService: emailService,
	}
}

// CheckAndSendReminders runs the daily reminder check
func (r *ReminderService) CheckAndSendReminders() {
	log.Println("Starting reminder check...")

	// Send upcoming due reminders (3 days before due date)
	if err := r.sendUpcomingDueReminders(); err != nil {
		log.Printf("Error sending upcoming due reminders: %v", err)
	}

	// Send overdue reminders
	if err := r.sendOverdueReminders(); err != nil {
		log.Printf("Error sending overdue reminders: %v", err)
	}

	log.Println("Reminder check completed")
}

func (r *ReminderService) sendUpcomingDueReminders() error {
	// Find books due in 3 days that haven't had a reminder sent in the last 24 hours
	// Only include users who have email reminders enabled
	query := `
		SELECT 
			l.id,
			l.user_id,
			u.email,
			b.title,
			b.author,
			l.lent_to,
			l.due_date,
			l.lent_at,
			l.last_reminder_sent
		FROM lending l
		JOIN users u ON l.user_id = u.id
		JOIN books b ON l.book_id = b.id
		LEFT JOIN user_settings us ON u.id = us.user_id
		WHERE l.returned_at IS NULL
		AND l.due_date IS NOT NULL
		AND DATE(l.due_date) = DATE('now', '+3 days')
		AND (l.last_reminder_sent IS NULL OR l.last_reminder_sent < datetime('now', '-24 hours'))
		AND (us.email_reminders_enabled IS NULL OR us.email_reminders_enabled = 1)
		AND (us.email_upcoming_reminders IS NULL OR us.email_upcoming_reminders = 1)
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var lending models.ReminderLending
		err := rows.Scan(
			&lending.LendingID,
			&lending.UserID,
			&lending.UserEmail,
			&lending.BookTitle,
			&lending.BookAuthor,
			&lending.LentTo,
			&lending.DueDate,
			&lending.LentAt,
			&lending.LastReminderSent,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Send email
		emailData := EmailData{
			UserEmail:  lending.UserEmail,
			BookTitle:  lending.BookTitle,
			BookAuthor: lending.BookAuthor,
			LentTo:     lending.LentTo,
			DueDate:    lending.DueDate,
			LentAt:     lending.LentAt,
		}

		if err := r.EmailService.SendUpcomingDueReminder(emailData); err != nil {
			log.Printf("Failed to send upcoming due reminder for lending %d: %v", lending.LendingID, err)
			continue
		}

		// Update last_reminder_sent
		if err := r.updateLastReminderSent(lending.LendingID); err != nil {
			log.Printf("Failed to update last_reminder_sent for lending %d: %v", lending.LendingID, err)
		}

		count++
	}

	log.Printf("Sent %d upcoming due reminders", count)
	return nil
}

func (r *ReminderService) sendOverdueReminders() error {
	// Find overdue books that haven't had a reminder in the last 24 hours
	// Only include users who have email reminders enabled
	query := `
		SELECT 
			l.id,
			l.user_id,
			u.email,
			b.title,
			b.author,
			l.lent_to,
			l.due_date,
			l.lent_at,
			l.last_reminder_sent
		FROM lending l
		JOIN users u ON l.user_id = u.id
		JOIN books b ON l.book_id = b.id
		LEFT JOIN user_settings us ON u.id = us.user_id
		WHERE l.returned_at IS NULL
		AND l.due_date IS NOT NULL
		AND DATE(l.due_date) < DATE('now')
		AND (l.last_reminder_sent IS NULL OR l.last_reminder_sent < datetime('now', '-24 hours'))
		AND (us.email_reminders_enabled IS NULL OR us.email_reminders_enabled = 1)
		AND (us.email_overdue_reminders IS NULL OR us.email_overdue_reminders = 1)
		ORDER BY l.user_id, l.due_date
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Group overdue books by user
	userBooks := make(map[int]*struct {
		Email      string
		Books      []OverdueBook
		LendingIDs []int
	})

	for rows.Next() {
		var lending models.ReminderLending
		err := rows.Scan(
			&lending.LendingID,
			&lending.UserID,
			&lending.UserEmail,
			&lending.BookTitle,
			&lending.BookAuthor,
			&lending.LentTo,
			&lending.DueDate,
			&lending.LentAt,
			&lending.LastReminderSent,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Calculate days overdue
		daysOverdue := int(time.Since(lending.DueDate).Hours() / 24)

		// Initialize user entry if doesn't exist
		if userBooks[lending.UserID] == nil {
			userBooks[lending.UserID] = &struct {
				Email      string
				Books      []OverdueBook
				LendingIDs []int
			}{
				Email: lending.UserEmail,
			}
		}

		// Add book to user's list
		userBooks[lending.UserID].Books = append(userBooks[lending.UserID].Books, OverdueBook{
			BookTitle:   lending.BookTitle,
			BookAuthor:  lending.BookAuthor,
			LentTo:      lending.LentTo,
			DueDate:     lending.DueDate,
			DaysOverdue: daysOverdue,
		})
		userBooks[lending.UserID].LendingIDs = append(userBooks[lending.UserID].LendingIDs, lending.LendingID)
	}

	// Send one digest email per user
	emailCount := 0
	bookCount := 0
	for userID, userData := range userBooks {
		// Send digest email
		digestData := OverdueDigestData{
			UserEmail:    userData.Email,
			OverdueBooks: userData.Books,
			TotalOverdue: len(userData.Books),
		}

		if err := r.EmailService.SendOverdueDigest(digestData); err != nil {
			log.Printf("Failed to send overdue digest to user %d: %v", userID, err)
			continue
		}

		// Update last_reminder_sent for all books in this digest
		for _, lendingID := range userData.LendingIDs {
			if err := r.updateLastReminderSent(lendingID); err != nil {
				log.Printf("Failed to update last_reminder_sent for lending %d: %v", lendingID, err)
			}
		}

		emailCount++
		bookCount += len(userData.Books)
	}

	log.Printf("Sent %d overdue digest email(s) covering %d book(s)", emailCount, bookCount)
	return nil
}

func (r *ReminderService) updateLastReminderSent(lendingID int) error {
	query := `UPDATE lending SET last_reminder_sent = datetime('now') WHERE id = ?`
	_, err := r.DB.Exec(query, lendingID)
	return err
}
