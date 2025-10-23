package api

import (
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router) {
	r.Get("/books", listBooks)
	r.Get("/books/{id}", getBook)
	r.Post("/books", createBook)
	r.Put("/books/{id}", updateBook)
	r.Delete("/books/{id}", deleteBook)

	r.Get("/search/{isbn}", searchBook)
}
