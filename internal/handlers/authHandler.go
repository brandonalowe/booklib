package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"booklib/internal/middleware"
	"booklib/internal/models"
	"booklib/internal/services"
)

type AuthHandler struct {
	DB *sql.DB
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	hash, err := services.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	result, err := h.DB.Exec(
		"INSERT INTO users (username, email, password_hash, role) VALUES (?, ?, ?, 'user')",
		req.Username,
		req.Email,
		hash,
	)

	if err != nil {
		http.Error(w, `{"error":"User already exists"}`, http.StatusConflict)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, `{"error":"Failed to get user ID"}`, http.StatusInternalServerError)
		return
	}
	userID := int(id)

	token, err := services.GenerateJWT(userID, req.Username, "user")
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Determine if we're in production (HTTPS)
	isProduction := r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400, // 24hrs
		HttpOnly: true,
		Secure:   isProduction,          // true in production (HTTPS)
		SameSite: http.SameSiteNoneMode, // Required for cross-origin
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "user created successfully",
		"user": map[string]any{
			"id":       userID,
			"username": req.Username,
			"email":    req.Email,
			"role":     "user",
		},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, username, email, password_hash, role FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role)

	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if !services.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := services.GenerateJWT(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Determine if we're in production (HTTPS)
	isProduction := r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   isProduction,          // true in production (HTTPS)
		SameSite: http.SameSiteNoneMode, // Required for cross-origin
	})

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Login successful",
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Determine if we're in production (HTTPS)
	isProduction := r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,          // Must match the original cookie
		SameSite: http.SameSiteNoneMode, // Must match the original cookie
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())
	username, _ := middleware.GetUsername(r.Context())
	role, _ := middleware.GetRole(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":       userID,
		"username": username,
		"role":     role,
	})
}
