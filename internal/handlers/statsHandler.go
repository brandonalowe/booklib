package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"booklib/internal/middleware"
)

type StatsHandler struct {
	DB *sql.DB
}

type StatsResponse struct {
	TotalBooks        int            `json:"total_books"`
	BooksRead         int            `json:"books_read"`
	BooksUnread       int            `json:"books_unread"`
	ReadPercentage    float64        `json:"read_percentage"`
	BooksLentOut      int            `json:"books_lent_out"`
	BooksThisMonth    int            `json:"books_this_month"`
	BooksThisYear     int            `json:"books_this_year"`
	BooksReadThisYear int            `json:"books_read_this_year"`
	TotalLendings     int            `json:"total_lendings"`
	GenreBreakdown    []GenreStat    `json:"genre_breakdown"`
	MonthlyReading    []MonthlyCount `json:"monthly_reading"`
	TopLentBooks      []TopBook      `json:"top_lent_books"`
}

type GenreStat struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

type MonthlyCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type TopBook struct {
	Title     string `json:"title"`
	Author    string `json:"author"`
	LentCount int    `json:"lent_count"`
	CoverURL  string `json:"cover_url,omitempty"`
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	stats := StatsResponse{}

	// Total books
	err := h.DB.QueryRow(
		"SELECT COUNT(*) FROM books WHERE user_id = ?",
		userID,
	).Scan(&stats.TotalBooks)
	if err != nil {
		http.Error(w, `{"error":"Failed to fetch stats"}`, http.StatusInternalServerError)
		return
	}

	// Books read
	err = h.DB.QueryRow(
		"SELECT COUNT(*) FROM books WHERE user_id = ? AND read = 1",
		userID,
	).Scan(&stats.BooksRead)
	if err != nil {
		stats.BooksRead = 0
	}

	// Books unread
	stats.BooksUnread = stats.TotalBooks - stats.BooksRead

	// Read percentage
	if stats.TotalBooks > 0 {
		stats.ReadPercentage = (float64(stats.BooksRead) / float64(stats.TotalBooks)) * 100
	}

	// Books currently lent out (not returned)
	err = h.DB.QueryRow(`
		SELECT COUNT(DISTINCT book_id) 
		FROM lending 
		WHERE user_id = ? AND returned_at IS NULL
	`, userID).Scan(&stats.BooksLentOut)
	if err != nil {
		stats.BooksLentOut = 0
	}

	// Total lendings (all time, including returned)
	err = h.DB.QueryRow(
		"SELECT COUNT(*) FROM lending WHERE user_id = ?",
		userID,
	).Scan(&stats.TotalLendings)
	if err != nil {
		stats.TotalLendings = 0
	}

	// Books added this month
	now := time.Now()
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	err = h.DB.QueryRow(`
		SELECT COUNT(*) FROM books 
		WHERE user_id = ? AND created_at >= ?
	`, userID, firstDayOfMonth).Scan(&stats.BooksThisMonth)
	if err != nil {
		stats.BooksThisMonth = 0
	}

	// Books added this year
	firstDayOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	err = h.DB.QueryRow(`
		SELECT COUNT(*) FROM books 
		WHERE user_id = ? AND created_at >= ?
	`, userID, firstDayOfYear).Scan(&stats.BooksThisYear)
	if err != nil {
		stats.BooksThisYear = 0
	}

	// Books read this year
	err = h.DB.QueryRow(`
		SELECT COUNT(*) FROM reading_history 
		WHERE user_id = ? AND read = 1 AND completed_at >= ?
	`, userID, firstDayOfYear).Scan(&stats.BooksReadThisYear)

	// Genre breakdown (top 5)
	rows, err := h.DB.Query(`
		SELECT genre, COUNT(*) as count 
		FROM books 
		WHERE user_id = ? AND genre != '' 
		GROUP BY genre 
		ORDER BY count DESC 
		LIMIT 5
	`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var genre GenreStat
			if err := rows.Scan(&genre.Genre, &genre.Count); err == nil {
				stats.GenreBreakdown = append(stats.GenreBreakdown, genre)
			}
		}
	}
	if stats.GenreBreakdown == nil {
		stats.GenreBreakdown = []GenreStat{}
	}

	// Monthly reading for last 12 months
	stats.MonthlyReading = []MonthlyCount{}
	for i := 11; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0)

		var count int
		err = h.DB.QueryRow(`
			SELECT COUNT(*) FROM books 
			WHERE user_id = ? AND created_at >= ? AND created_at < ?
		`, userID, monthStart, monthEnd).Scan(&count)

		if err == nil {
			stats.MonthlyReading = append(stats.MonthlyReading, MonthlyCount{
				Month: monthStart.Format("Jan 2006"),
				Count: count,
			})
		}
	}

	// Top lent books
	rows, err = h.DB.Query(`
		SELECT b.title, b.author, COUNT(l.id) as lent_count, b.cover_url
		FROM books b
		JOIN lending l ON b.id = l.book_id
		WHERE b.user_id = ?
		GROUP BY b.id, b.title, b.author, b.cover_url
		ORDER BY lent_count DESC
		LIMIT 5
	`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var book TopBook
			var coverURL sql.NullString
			if err := rows.Scan(&book.Title, &book.Author, &book.LentCount, &coverURL); err == nil {
				if coverURL.Valid {
					book.CoverURL = coverURL.String
				}
				stats.TopLentBooks = append(stats.TopLentBooks, book)
			}
		}
	}
	if stats.TopLentBooks == nil {
		stats.TopLentBooks = []TopBook{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
