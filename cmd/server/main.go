package main

import (
	"booklib/internal/api"
	"booklib/internal/db"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed web/*
var webFiles embed.FS

func main() {
	db.Init("books.db")

	r := chi.NewRouter()
	api.RegisterRoutes(r)

	// Serve static assets under /static/.
	// Prefer local ./web directory for development so edits are visible immediately.
	var staticHandler http.Handler
	if fi, err := os.Stat("web"); err == nil && fi.IsDir() {
		pwd, _ := os.Getwd()
		log.Printf("serving static assets from local %s", filepath.Join(pwd, "web"))
		staticHandler = http.StripPrefix("/static/", http.FileServer(http.Dir("web")))
	} else {
		sub, err := fs.Sub(webFiles, "web")
		if err != nil {
			log.Fatalf("failed to create sub fs: %v", err)
		}
		staticFS := http.FS(sub)
		staticHandler = http.StripPrefix("/static/", http.FileServer(staticFS))
	}
	r.Handle("/static/*", staticHandler)

	// Serve index.html at root (prefer local file if present)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if fi, err := os.Stat("web/index.html"); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, "web/index.html")
			return
		}
		data, err := webFiles.ReadFile("web/index.html")
		if err != nil {
			http.Error(w, "index not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// Simple NotFound that returns index.html for SPA paths (but not API paths)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/books") || strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/static/") {
			http.NotFound(w, r)
			return
		}
		if fi, err := os.Stat("web/index.html"); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, "web/index.html")
			return
		}
		data, err := webFiles.ReadFile("web/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", r)
}
