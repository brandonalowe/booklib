package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"booklib/internal/api"
	"booklib/internal/middleware"
	"booklib/internal/models"

	"github.com/go-chi/chi/v5"
)

type BookHandler struct {
	DB *sql.DB
}

func (h *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	rows, err := h.DB.Query(
		"SELECT id, title, author, isbn, genre, read, created_at FROM books WHERE user_id = ?",
		userID,
	)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch books"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		var book models.Book
		var readInt int
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Genre, &readInt, &book.CreatedAt); err != nil {
			continue
		}
		book.Read = readInt == 1
		books = append(books, book)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Check if user already owns a book with this ISBN
	if book.ISBN != "" {
		var existingID int
		err := h.DB.QueryRow(
			"SELECT id FROM books WHERE user_id = ? AND isbn = ?",
			userID, book.ISBN,
		).Scan(&existingID)

		if err == nil {
			// Book with this ISBN already exists for this user
			http.Error(w, `{"error":"You already own a book with this ISBN"}`, http.StatusConflict)
			return
		} else if err != sql.ErrNoRows {
			// Some other database error occurred
			http.Error(w, `{"error":"Failed to check for duplicate ISBN"}`, http.StatusInternalServerError)
			return
		}
	}

	readInt := 0
	if book.Read {
		readInt = 1
	}

	result, err := h.DB.Exec(
		"INSERT INTO books (user_id, title, author, isbn, genre, read) VALUES (?, ?, ?, ?, ?, ?)",
		userID, book.Title, book.Author, book.ISBN, book.Genre, readInt,
	)
	if err != nil {
		// Check if it's a unique constraint violation (in case the check above was bypassed)
		http.Error(w, `{"error":"Failed to create book"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	book.ID = int(id)

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	bookID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid book ID"}`, http.StatusBadRequest)
		return
	}

	var book models.Book
	var readInt int
	err = h.DB.QueryRow(
		"SELECT id, title, author, isbn, genre, read, created_at FROM books WHERE id = ? AND user_id = ?",
		bookID,
		userID,
	).Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Genre, &readInt, &book.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Book not found"}`, http.StatusNotFound)
		return
	}

	book.Read = readInt == 1

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	bookID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid book ID"}`, http.StatusBadRequest)
		return
	}

	var book models.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	readInt := 0
	if book.Read {
		readInt = 1
	}

	result, err := h.DB.Exec(
		"UPDATE books SET title=?, author=?, isbn=?, genre=?, read=? WHERE id=? AND user_id=?",
		book.Title,
		book.Author,
		book.ISBN,
		book.Genre,
		readInt,
		bookID,
		userID,
	)
	if err != nil {
		http.Error(w, `{"error":"Failed to update book"}`, http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"Book not found"}`, http.StatusNotFound)
		return
	}

	book.ID = bookID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	bookID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"Invalid book ID"}`, http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("DELETE FROM books WHERE id=? AND user_id=?", bookID, userID)
	if err != nil {
		http.Error(w, `{"error":"Failed to delete book"}`, http.StatusInternalServerError)
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		http.Error(w, `{"error":"Book not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Book deleted successfully"})
}

func (h *BookHandler) SearchByISBN(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	isbnParam := chi.URLParam(r, "isbn")

	isbn, err := strconv.Atoi(isbnParam)
	if err != nil {
		http.Error(w, `{"error":"Invalid ISBN"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Check if user already owns this book
	var existingBookID int
	err = h.DB.QueryRowContext(ctx,
		"SELECT id FROM books WHERE user_id = ? AND isbn = ?",
		userID, isbnParam,
	).Scan(&existingBookID)

	alreadyOwned := err == nil

	// Check db cache first
	row := h.DB.QueryRowContext(
		ctx,
		"SELECT isbn, title, author, genre, cover_url, cached_at FROM isbn_cache WHERE isbn = ?",
		isbnParam,
	)

	var cache models.IsbnCache
	err = row.Scan(&cache.ISBN, &cache.Title, &cache.Author, &cache.Genre, &cache.CoverUrl, &cache.CachedAt)

	if err == sql.ErrNoRows {
		// No book found in Cache -> search via Google API
		gb, gerr := api.SearchGoogleApi(ctx, isbn)
		if gerr != nil {
			http.Error(w, `{"error":"Failed to search Google Books API"}`, http.StatusInternalServerError)
			return
		}
		if gb == nil {
			http.Error(w, `{"error":"Book not found"}`, http.StatusNotFound)
			return
		}

		// Cache in DB
		_, ierr := h.DB.ExecContext(
			ctx,
			`INSERT INTO isbn_cache (isbn, title, author, genre, cover_url, cached_at) VALUES (?, ?, ?, ?, ?, ?)`,
			gb.ISBN,
			gb.Title,
			gb.Author,
			gb.Genre,
			gb.CoverUrl,
			gb.CachedAt,
		)
		if ierr != nil {
			// Log but don't fail the request
			http.Error(w, `{"error":"Failed to cache book data"}`, http.StatusInternalServerError)
			return
		}

		// Return result with ownership status
		result := map[string]interface{}{
			"isbn":          gb.ISBN,
			"title":         gb.Title,
			"author":        gb.Author,
			"genre":         gb.Genre,
			"cover_url":     gb.CoverUrl,
			"already_owned": alreadyOwned,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	} else if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	// Book found in Cache -> Return with ownership status
	result := map[string]interface{}{
		"isbn":          cache.ISBN,
		"title":         cache.Title,
		"author":        cache.Author,
		"genre":         cache.Genre,
		"cover_url":     cache.CoverUrl,
		"already_owned": alreadyOwned,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *BookHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	query := r.URL.Query().Get("q")

	if query == "" {
		http.Error(w, `{"error":"Search query is required"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Check if the query looks like an ISBN (only digits, 10 or 13 characters)
	isISBN := len(query) >= 10 && len(query) <= 13
	for _, c := range query {
		if c < '0' || c > '9' {
			isISBN = false
			break
		}
	}

	var books []*models.IsbnCache

	// If it's an ISBN, check cache first
	if isISBN {
		row := h.DB.QueryRowContext(
			ctx,
			"SELECT isbn, title, author, genre, cover_url, cached_at FROM isbn_cache WHERE isbn = ?",
			query,
		)

		var cache models.IsbnCache
		err := row.Scan(&cache.ISBN, &cache.Title, &cache.Author, &cache.Genre, &cache.CoverUrl, &cache.CachedAt)

		if err == nil {
			// Found in cache
			books = []*models.IsbnCache{&cache}
		}
	}

	// If not found in cache or not an ISBN, search Google Books API
	if len(books) == 0 {
		results, err := api.SearchGoogleBooks(ctx, query)
		if err != nil {
			http.Error(w, `{"error":"Failed to search Google Books API"}`, http.StatusInternalServerError)
			return
		}
		books = results

		// Cache the results if we got any
		for _, book := range books {
			if book.ISBN != "" {
				// Check if already cached
				var exists int
				err := h.DB.QueryRowContext(ctx, "SELECT 1 FROM isbn_cache WHERE isbn = ?", book.ISBN).Scan(&exists)
				if err == sql.ErrNoRows {
					// Not cached, insert it
					_, _ = h.DB.ExecContext(
						ctx,
						`INSERT INTO isbn_cache (isbn, title, author, genre, cover_url, cached_at) VALUES (?, ?, ?, ?, ?, ?)`,
						book.ISBN,
						book.Title,
						book.Author,
						book.Genre,
						book.CoverUrl,
						book.CachedAt,
					)
				}
			}
		}
	}

	// Check ownership for each book
	results := make([]map[string]interface{}, 0, len(books))
	for _, book := range books {
		alreadyOwned := false
		if book.ISBN != "" {
			var existingBookID int
			err := h.DB.QueryRowContext(ctx,
				"SELECT id FROM books WHERE user_id = ? AND isbn = ?",
				userID, book.ISBN,
			).Scan(&existingBookID)
			alreadyOwned = err == nil
		}

		results = append(results, map[string]interface{}{
			"isbn":          book.ISBN,
			"title":         book.Title,
			"author":        book.Author,
			"genre":         book.Genre,
			"cover_url":     book.CoverUrl,
			"already_owned": alreadyOwned,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
