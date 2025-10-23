package api

import (
	"booklib/internal/db"
	"booklib/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var SQL_COLUMNS string = "title, author, isbn, genre, read, lent_to, lent_at"

func listBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := db.DB.QueryContext(ctx, "SELECT id, "+SQL_COLUMNS+" FROM books;")
	if err != nil {
		InternalServerErrorResponse(w, err)
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			InternalServerErrorResponse(w, err)
			return
		}
		books = append(books, *b)
	}
	OkResponse(w, books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	ctx := r.Context()
	row := db.DB.QueryRowContext(ctx, "SELECT id, "+SQL_COLUMNS+" FROM books WHERE id = ?", id)
	b, err := scanBook(row)
	if err == sql.ErrNoRows {
		NotFoundResponse(w, "book not found")
		return
	}
	OkResponse(w, b)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var in models.Book
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		BadRequestResponse(w, "invalid JSON")
		return
	}
	ctx := r.Context()
	res, err := db.DB.ExecContext(
		ctx,
		`INSERT INTO books (`+SQL_COLUMNS+`) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		in.Title,
		in.Author,
		in.ISBN,
		in.Genre,
		boolToInt(in.Read),
		in.LentTo,
		in.LentAt,
	)
	if err != nil {
		InternalServerErrorResponse(w, err)
		return
	}
	id, _ := res.LastInsertId()
	in.ID = int(id)
	CreatedResponse(w, in)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var in models.Book
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		BadRequestResponse(w, "invalid JSON")
		return
	}
	ctx := r.Context()
	_, err := db.DB.ExecContext(
		ctx,
		`UPDATE books SET title=?, author=?, isbn=?, genre=?, read=?, lent_to=?, lent_at=? WHERE id=?`,
		in.Title,
		in.Author,
		in.ISBN,
		in.Genre,
		boolToInt(in.Read),
		in.LentTo,
		in.LentAt,
		id,
	)
	if err != nil {
		InternalServerErrorResponse(w, err)
		return
	}
	in.ID = id
	OkResponse(w, in)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	ctx := r.Context()
	_, err := db.DB.ExecContext(ctx, "DELETE FROM books WHERE id = ?", id)
	if err != nil {
		InternalServerErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func searchBook(w http.ResponseWriter, r *http.Request) {
	isbn, err := strconv.Atoi(chi.URLParam(r, "isbn"))
	if err != nil {
		BadRequestResponse(w, "Invalid ISBN")
		return
	}
	ctx := r.Context()
	// Check db cache first
	row := db.DB.QueryRowContext(
		ctx,
		"SELECT isbn, title, author, isbn, genre, cover_url, cached_at FROM isbn_cache WHERE isbn = ?",
		isbn,
	)
	b, err := scanCache(row)
	if err == sql.ErrNoRows {
		// No book found in Cache -> search via Google API (propagate context)
		gb, gerr := searchGoogleApi(ctx, isbn)
		if gerr != nil {
			if gerr == ctx.Err() {
				InternalServerErrorResponse(w, gerr)
				return
			}
			InternalServerErrorResponse(w, gerr)
			return
		}
		if gb == nil {
			NotFoundResponse(w, "book not found")
			return
		}
		// If book found via Google API -> Cache in DB and return book
		_, ierr := db.DB.ExecContext(
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
			InternalServerErrorResponse(w, ierr)
			return
		}
		OkResponse(w, gb)
		return
	} else if err != nil {
		InternalServerErrorResponse(w, err)
		return
	}
	// Book found in Cache -> Return Ok Response with book
	OkResponse(w, b)
}
