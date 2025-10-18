package api

import (
	"booklib/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
)

func writeJson(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func scanBook(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Book, error) {
	var b models.Book
	var readInt int
	var lentAt sql.NullTime
	err := scanner.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Genre, &readInt, &b.LentTo, &lentAt)
	if err != nil {
		return nil, err
	}
	b.Read = readInt == 1
	if lentAt.Valid {
		b.LentAt = &lentAt.Time
	}
	return &b, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
