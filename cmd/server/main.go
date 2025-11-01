package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"booklib/internal/db"
	"booklib/internal/handlers"
	"booklib/internal/middleware"
	"booklib/internal/services"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron/v3"
)

// splitAndTrim splits a string by delimiter and trims whitespace
func splitAndTrim(s, delim string) []string {
	parts := strings.Split(s, delim)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func main() {
	// Load .env file if it exists (optional, won't fail if missing)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	// Initialize database
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./database/booklib.db"
	}
	if err := db.Init(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	authHandler := &handlers.AuthHandler{DB: db.GetDB()}
	bookHandler := &handlers.BookHandler{DB: db.GetDB()}
	adminHandler := &handlers.AdminHandler{DB: db.GetDB()}
	lendingHandler := &handlers.LendingHandler{DB: db.GetDB()}
	statsHandler := &handlers.StatsHandler{DB: db.GetDB()}
	readingHistoryHandler := &handlers.ReadingHistoryHandler{DB: db.GetDB()}
	userSettingsHandler := &handlers.UserSettingsHandler{DB: db.GetDB()}

	// Initialize email and reminder services
	emailService := services.NewEmailService()
	reminderService := services.NewReminderService(db.GetDB(), emailService)

	// Setup cron scheduler for daily reminders
	c := cron.New()

	// Run daily at 9 AM (adjust timezone as needed)
	cronSchedule := os.Getenv("REMINDER_CRON_SCHEDULE")
	if cronSchedule == "" {
		cronSchedule = "0 9 * * *" // Default: 9 AM daily
	}

	_, err := c.AddFunc(cronSchedule, func() {
		log.Println("Running scheduled reminder check...")
		reminderService.CheckAndSendReminders()
	})

	if err != nil {
		log.Printf("Warning: Failed to schedule reminder cron job: %v", err)
	} else {
		c.Start()
		log.Printf("Reminder cron job scheduled: %s", cronSchedule)

		// Run immediately on startup if enabled
		if os.Getenv("RUN_REMINDERS_ON_STARTUP") == "true" {
			log.Println("Running initial reminder check on startup...")
			go reminderService.CheckAndSendReminders()
		}
	}

	defer c.Stop()

	r := chi.NewRouter()
	// Add common middleware including a 15s request timeout
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Configure CORS based on environment
	allowedOrigins := []string{"http://localhost:5173"}
	if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
		// Parse comma-separated origins
		allowedOrigins = append(allowedOrigins, splitAndTrim(corsOrigins, ",")...)
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"booklib-backend"}`))
	})

	// Public auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)

		// Protected auth routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Get("/me", authHandler.Me)
		})
	})

	// Protected book routes
	r.Route("/api/books", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/", bookHandler.List)
		r.Post("/", bookHandler.Create)
		r.Get("/search", bookHandler.Search)
		r.Get("/search/{isbn}", bookHandler.SearchByISBN)
		r.Get("/{id}", bookHandler.Get)
		r.Put("/{id}", bookHandler.Update)
		r.Delete("/{id}", bookHandler.Delete)
	})

	// Protected lending routes
	r.Route("/api/lending", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/", lendingHandler.List)
		r.Post("/", lendingHandler.Create)
		r.Delete("/{id}", lendingHandler.Return)
		r.Get("/history", lendingHandler.GetHistory)
		r.Get("/history/{bookId}", lendingHandler.GetHistory)
	})

	// Protected stats routes
	r.Route("/api/stats", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/", statsHandler.GetStats)
	})

	// Protected reading history routes
	r.Route("/api/reading-history", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Post("/start", readingHistoryHandler.StartReading)
		r.Put("/{id}/finish", readingHistoryHandler.FinishReading)
		r.Get("/book/{bookId}", readingHistoryHandler.GetBookReadingHistory)
		r.Get("/book/{bookId}/active", readingHistoryHandler.GetActiveReadingSession)
	})

	// Protected user settings routes
	r.Route("/api/user-settings", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Get("/", userSettingsHandler.GetUserSettings)
		r.Put("/", userSettingsHandler.UpdateUserSettings)
	})

	// Admin routes
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.AdminMiddleware)

		r.Get("/stats", adminHandler.GetStats)
		r.Get("/users", adminHandler.ListUsers)
		r.Get("/users/{id}", adminHandler.GetUser)
		r.Delete("/users/{id}", adminHandler.DeleteUser)
		r.Put("/users/{id}/role", adminHandler.UpdateUserRole)
		r.Get("/settings", adminHandler.GetSettings)
		r.Put("/settings", adminHandler.UpdateSetting)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Printf("Database path: %s", dbPath)
	log.Printf("Allowed CORS origins: %v", allowedOrigins)

	http.ListenAndServe(":"+port, r)
}
