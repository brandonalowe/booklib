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

type LendingHandler struct {
	DB *sql.DB
}

// List returns all active (not returned) lending records for the authenticated user
func (h *LendingHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	query := `
		SELECT 
			l.id, l.book_id, l.user_id, l.lent_to, l.lent_at, l.due_date,
			b.id, b.title, b.author, b.isbn, b.genre, b.read
		FROM lending l
		JOIN books b ON l.book_id = b.id
		WHERE l.user_id = ? AND l.returned_at IS NULL
		ORDER BY l.lent_at DESC
	`

	rows, err := h.DB.Query(query, userID)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch lending records"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	lendings := []models.LendingWithBook{}
	for rows.Next() {
		var lending models.LendingWithBook
		var book models.Book
		var readInt int
		var dueDate sql.NullTime

		if err := rows.Scan(
			&lending.ID, &lending.BookID, &lending.UserID, &lending.LentTo, &lending.LentAt, &dueDate,
			&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Genre, &readInt,
		); err != nil {
			continue
		}

		book.Read = readInt == 1
		lending.Book = &book

		if dueDate.Valid {
			lending.DueDate = &dueDate.Time
		}

		lendings = append(lendings, lending)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lendings)
}

// Create lends out a book
func (h *LendingHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var req models.CreateLendingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Validate that the book belongs to the user
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

	// Check if book is already lent out
	var existingID int
	err = h.DB.QueryRow(
		"SELECT id FROM lending WHERE book_id = ? AND returned_at IS NULL",
		req.BookID,
	).Scan(&existingID)
	if err == nil {
		http.Error(w, `{"error":"Book is already lent out"}`, http.StatusConflict)
		return
	}

	// Parse due date if provided
	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			http.Error(w, `{"error":"Invalid due date format"}`, http.StatusBadRequest)
			return
		}
		dueDate = &parsed
	}

	// Create lending record
	result, err := h.DB.Exec(
		"INSERT INTO lending (book_id, user_id, lent_to, due_date) VALUES (?, ?, ?, ?)",
		req.BookID, userID, req.LentTo, dueDate,
	)
	if err != nil {
		http.Error(w, `{"error":"Failed to create lending record"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	lending := models.Lending{
		ID:      int(id),
		BookID:  req.BookID,
		UserID:  userID,
		LentTo:  req.LentTo,
		LentAt:  time.Now(),
		DueDate: dueDate,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lending)
}

// Return marks a book as returned
func (h *LendingHandler) Return(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	lendingID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid lending ID"}`, http.StatusBadRequest)
		return
	}

	// Verify the lending record belongs to the user
	var recordUserID int
	err = h.DB.QueryRow(
		"SELECT user_id FROM lending WHERE id = ? AND returned_at IS NULL",
		lendingID,
	).Scan(&recordUserID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Lending record not found or already returned"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch lending record"}`, http.StatusInternalServerError)
		return
	}
	if recordUserID != userID {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusForbidden)
		return
	}

	// Mark as returned
	_, err = h.DB.Exec(
		"UPDATE lending SET returned_at = CURRENT_TIMESTAMP WHERE id = ?",
		lendingID,
	)
	if err != nil {
		http.Error(w, `{"error":"Failed to mark book as returned"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetHistory returns the complete lending history for a specific book (optional) or all books
func (h *LendingHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var query string
	var args []interface{}

	// Check if a specific book ID is provided
	bookIDParam := chi.URLParam(r, "bookId")
	if bookIDParam != "" {
		bookID, err := strconv.Atoi(bookIDParam)
		if err != nil {
			http.Error(w, `{"error":"Invalid book ID"}`, http.StatusBadRequest)
			return
		}

		query = `
			SELECT 
				l.id, l.book_id, l.user_id, l.lent_to, l.lent_at, l.due_date, l.returned_at,
				b.id, b.title, b.author, b.isbn, b.genre, b.read
			FROM lending l
			JOIN books b ON l.book_id = b.id
			WHERE l.user_id = ? AND l.book_id = ?
			ORDER BY l.lent_at DESC
		`
		args = []interface{}{userID, bookID}
	} else {
		query = `
			SELECT 
				l.id, l.book_id, l.user_id, l.lent_to, l.lent_at, l.due_date, l.returned_at,
				b.id, b.title, b.author, b.isbn, b.genre, b.read
			FROM lending l
			JOIN books b ON l.book_id = b.id
			WHERE l.user_id = ?
			ORDER BY l.lent_at DESC
		`
		args = []interface{}{userID}
	}

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch lending history"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	history := []models.LendingWithBook{}
	for rows.Next() {
		var lending models.LendingWithBook
		var book models.Book
		var readInt int
		var dueDate, returnedAt sql.NullTime

		if err := rows.Scan(
			&lending.ID, &lending.BookID, &lending.UserID, &lending.LentTo,
			&lending.LentAt, &dueDate, &returnedAt,
			&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Genre, &readInt,
		); err != nil {
			continue
		}

		book.Read = readInt == 1
		lending.Book = &book

		if dueDate.Valid {
			lending.DueDate = &dueDate.Time
		}

		history = append(history, lending)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
