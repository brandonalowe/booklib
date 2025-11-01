package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"booklib/internal/models"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	DB *sql.DB
}

type Stats struct {
	TotalUsers    int            `json:"total_users"`
	TotalBooks    int            `json:"total_books"`
	BooksRead     int            `json:"books_read"`
	BooksUnread   int            `json:"books_unread"`
	BooksLentOut  int            `json:"books_lent_out"`
	OverdueBooks  int            `json:"overdue_books"`
	RecentUsers   []models.User  `json:"recent_users"`
	RecentBooks   []BookWithUser `json:"recent_books"`
	PopularGenres []GenreCount   `json:"popular_genres"`
}

type BookWithUser struct {
	models.Book
	Username string `json:"username"`
}

type GenreCount struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

type UserWithStats struct {
	models.User
	BookCount int `json:"book_count"`
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var stats Stats

	h.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	h.DB.QueryRow("SELECT COUNT(*) FROM books").Scan(&stats.TotalBooks)

	h.DB.QueryRow("SELECT COUNT(*) FROM books WHERE read = 1").Scan(&stats.BooksRead)
	h.DB.QueryRow("SELECT COUNT(*) FROM books WHERE read = 0").Scan(&stats.BooksUnread)

	// Lending stats
	h.DB.QueryRow("SELECT COUNT(*) FROM lending WHERE returned_at IS NULL").Scan(&stats.BooksLentOut)
	h.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM lending 
		WHERE returned_at IS NULL 
		AND due_date IS NOT NULL 
		AND due_date < datetime('now')
	`).Scan(&stats.OverdueBooks)

	rows, err := h.DB.Query(`
		SELECT id, username, email, role, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT 5`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var user models.User
			rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
			stats.RecentUsers = append(stats.RecentUsers, user)
		}
	}

	rows, err = h.DB.Query(`
		SELECT b.id, b.title, b.author, b.isbn, b.genre, b.read, b.created_at, u.username
		FROM books b
		JOIN users u ON b.user_id = u.id
		ORDER BY b.created_at DESC
		LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var book BookWithUser
			var readInt int
			rows.Scan(
				&book.ID,
				&book.Title,
				&book.Author,
				&book.ISBN,
				&book.Genre,
				&readInt,
				&book.CreatedAt,
				&book.Username,
			)
			book.Read = readInt == 1
			stats.RecentBooks = append(stats.RecentBooks, book)
		}
	}

	rows, err = h.DB.Query(`
		SELECT genre, COUNT(*) as count
		FROM books
		WHERE genre != ''
		GROUP BY genre
		ORDER BY count DESC
		LIMIT 10`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var gc GenreCount
			rows.Scan(&gc.Genre, &gc.Count)
			stats.PopularGenres = append(stats.PopularGenres, gc)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT u.id, u.username, u.email, u.role, u.created_at, COUNT(b.id) as book_count
		FROM users u
		LEFT JOIN books b ON u.id = b.user_id
		GROUP BY u.id
		ORDER BY u.created_at DESC`)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch users"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []UserWithStats{}
	for rows.Next() {
		var user UserWithStats
		rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.CreatedAt,
			&user.BookCount,
		)
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	var user models.User
	err = h.DB.QueryRow(
		"SELECT id, username, email, role, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	rows, err := h.DB.Query(
		"SELECT id, title, author, isbn, genre, read, created_at FROM books WHERE user_id = ?",
		userID,
	)
	if err == nil {
		defer rows.Close()
		books := []models.Book{}
		for rows.Next() {
			var book models.Book
			var readInt int
			rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Genre, &readInt, &book.CreatedAt)
			book.Read = readInt == 1
			books = append(books, book)
		}
		response := map[string]any{
			"user":  user,
			"books": books,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusInternalServerError)
		return
	}

	result, err := h.DB.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		http.Error(w, `{"error":"Failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Role != "user" && req.Role != "admin" {
		http.Error(w, `{"error":"Role must be 'user' or 'admin'"}`, http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE users SET role = ? WHERE id = ?", req.Role, userID)
	if err != nil {
		http.Error(w, `{"error":"Failed to update role"}`, http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Role updated successfully"})
}

func (h *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings := make(map[string]string)

	rows, err := h.DB.Query("SELECT key, value FROM settings")
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch settings"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		settings[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

func (h *AdminHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, `{"error":"Key is required"}`, http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(`
		INSERT INTO settings (key, value, updated_at) 
		VALUES (?, ?, datetime('now'))
		ON CONFLICT(key) DO UPDATE SET 
			value = excluded.value,
			updated_at = datetime('now')
	`, req.Key, req.Value)

	if err != nil {
		http.Error(w, `{"error":"Failed to update setting"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Setting updated successfully"})
}
