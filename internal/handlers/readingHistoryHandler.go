package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"booklib/internal/middleware"
	"booklib/internal/models"

	"github.com/go-chi/chi/v5"
)

type ReadingHistoryHandler struct {
	DB *sql.DB
}

// StartReading creates a new reading history entry with started_at timestamp
func (h *ReadingHistoryHandler) StartReading(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req models.StartReadingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Verify the book belongs to the user
	var bookUserID int
	err := h.DB.QueryRow("SELECT user_id FROM books WHERE id = ?", req.BookID).Scan(&bookUserID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Book not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"Failed to verify book ownership"}`, http.StatusInternalServerError)
		return
	}
	if bookUserID != userID {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusForbidden)
		return
	}

	// Check if there's already an active reading session (started but not completed)
	var activeSessionID int
	err = h.DB.QueryRow(`
		SELECT id FROM reading_history 
		WHERE book_id = ? AND user_id = ? AND completed_at IS NULL
		ORDER BY started_at DESC LIMIT 1
	`, req.BookID, userID).Scan(&activeSessionID)

	if err == nil {
		// Active session exists, return it
		http.Error(w, `{"error":"Already have an active reading session for this book"}`, http.StatusConflict)
		return
	}

	// Create new reading history entry
	result, err := h.DB.Exec(`
		INSERT INTO reading_history (book_id, user_id, started_at)
		VALUES (?, ?, ?)
	`, req.BookID, userID, time.Now())

	if err != nil {
		http.Error(w, `{"error":"Failed to start reading session"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	// Return the created reading history entry
	var history models.ReadingHistory
	err = h.DB.QueryRow(`
		SELECT id, book_id, user_id, started_at, completed_at
		FROM reading_history WHERE id = ?
	`, id).Scan(&history.ID, &history.BookID, &history.UserID, &history.StartedAt, &history.CompletedAt)

	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve reading session"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(history)
}

// FinishReading marks a reading history entry as completed
func (h *ReadingHistoryHandler) FinishReading(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	historyID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid reading history ID"}`, http.StatusBadRequest)
		return
	}

	// Verify the reading history belongs to the user and is not already completed
	var historyUserID int
	var completedAt sql.NullTime
	err = h.DB.QueryRow(`
		SELECT user_id, completed_at FROM reading_history WHERE id = ?
	`, historyID).Scan(&historyUserID, &completedAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Reading history not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"Failed to verify reading history"}`, http.StatusInternalServerError)
		return
	}
	if historyUserID != userID {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusForbidden)
		return
	}
	if completedAt.Valid {
		http.Error(w, `{"error":"Reading session already completed"}`, http.StatusConflict)
		return
	}

	// Get the book_id from the reading history
	var bookID int
	err = h.DB.QueryRow(`
		SELECT book_id FROM reading_history WHERE id = ?
	`, historyID).Scan(&bookID)

	if err != nil {
		http.Error(w, `{"error":"Failed to get book ID"}`, http.StatusInternalServerError)
		return
	}

	// Mark reading session as completed
	_, err = h.DB.Exec(`
		UPDATE reading_history SET completed_at = ? WHERE id = ?
	`, time.Now(), historyID)

	if err != nil {
		http.Error(w, `{"error":"Failed to complete reading session"}`, http.StatusInternalServerError)
		return
	}

	// Mark the book as read
	_, err = h.DB.Exec(`
		UPDATE books SET read = TRUE WHERE id = ? AND user_id = ?
	`, bookID, userID)

	if err != nil {
		http.Error(w, `{"error":"Failed to mark book as read"}`, http.StatusInternalServerError)
		return
	}

	// Return the updated reading history entry
	var history models.ReadingHistory
	err = h.DB.QueryRow(`
		SELECT id, book_id, user_id, started_at, completed_at
		FROM reading_history WHERE id = ?
	`, historyID).Scan(&history.ID, &history.BookID, &history.UserID, &history.StartedAt, &history.CompletedAt)

	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve reading session"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// GetBookReadingHistory returns all reading history for a specific book
func (h *ReadingHistoryHandler) GetBookReadingHistory(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	bookID, err := strconv.Atoi(chi.URLParam(r, "bookId"))
	if err != nil {
		http.Error(w, `{"error":"Invalid book ID"}`, http.StatusBadRequest)
		return
	}

	// Verify the book belongs to the user
	var bookUserID int
	err = h.DB.QueryRow("SELECT user_id FROM books WHERE id = ?", bookID).Scan(&bookUserID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Book not found"}`, http.StatusNotFound)
		return
	}
	if bookUserID != userID {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusForbidden)
		return
	}

	// Get all reading history for this book
	rows, err := h.DB.Query(`
		SELECT id, book_id, user_id, started_at, completed_at
		FROM reading_history 
		WHERE book_id = ? AND user_id = ?
		ORDER BY started_at DESC
	`, bookID, userID)

	if err != nil {
		http.Error(w, `{"error":"Failed to fetch reading history"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	history := []models.ReadingHistory{}
	for rows.Next() {
		var h models.ReadingHistory
		if err := rows.Scan(&h.ID, &h.BookID, &h.UserID, &h.StartedAt, &h.CompletedAt); err != nil {
			continue
		}
		history = append(history, h)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// GetActiveReadingSession returns the current active reading session for a book (if any)
func (h *ReadingHistoryHandler) GetActiveReadingSession(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	bookID, err := strconv.Atoi(chi.URLParam(r, "bookId"))
	if err != nil {
		http.Error(w, `{"error":"Invalid book ID"}`, http.StatusBadRequest)
		return
	}

	// Get active reading session (started but not completed)
	var history models.ReadingHistory
	err = h.DB.QueryRow(`
		SELECT id, book_id, user_id, started_at, completed_at
		FROM reading_history 
		WHERE book_id = ? AND user_id = ? AND completed_at IS NULL
		ORDER BY started_at DESC LIMIT 1
	`, bookID, userID).Scan(&history.ID, &history.BookID, &history.UserID, &history.StartedAt, &history.CompletedAt)

	if err == sql.ErrNoRows {
		// No active session
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nil)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch active reading session"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
