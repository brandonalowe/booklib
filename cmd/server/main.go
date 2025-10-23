package main

import (
	"booklib/internal/api"
	"booklib/internal/db"
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

//go:embed web/*
var webFiles embed.FS

func main() {
	db.Init("books.db")
	//	db.Init("isbn_cache.db")

	r := chi.NewRouter()
	// Add common middleware including a 15s request timeout
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))
	api.RegisterRoutes(r)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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
		if strings.HasPrefix(r.URL.Path, "/books") || strings.HasPrefix(r.URL.Path, "/api") ||
			strings.HasPrefix(r.URL.Path, "/static/") {
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

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	go func() {
		log.Println("Listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped")
}
