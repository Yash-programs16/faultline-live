package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/middleware"
	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/repositories"
	"faultline-live/backend/internal/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type AuthHandler struct {
	userRepo *repositories.UserRepository
	cfg      *config.AuthConfig
}

func NewAuthHandler(userRepo *repositories.UserRepository, cfg *config.AuthConfig) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, cfg: cfg}
}

func (h *AuthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/auth/register", h.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", h.Login).Methods("POST")
	r.HandleFunc("/api/auth/logout", h.Logout).Methods("POST")
	r.HandleFunc("/api/auth/me", h.Me).Methods("GET")
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		response.BadRequest(w, "Name, email, and password are required")
		return
	}

	if len(req.Password) < 8 {
		response.BadRequest(w, "Password must be at least 8 characters")
		return
	}

	user, err := h.userRepo.Create(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		if err == repositories.ErrEmailExists {
			response.BadRequest(w, "Email already registered")
			return
		}
		response.InternalError(w, "Failed to create user")
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		response.InternalError(w, "Failed to generate token")
		return
	}

	response.Created(w, AuthResponse{
		User:  user,
		Token: token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.BadRequest(w, "Email and password are required")
		return
	}

	user, err := h.userRepo.VerifyPassword(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == repositories.ErrUserNotFound || err == repositories.ErrInvalidPassword {
			response.Unauthorized(w, "Invalid email or password")
			return
		}
		response.InternalError(w, "Login failed")
		return
	}

	token, err := h.generateToken(user.ID)
	if err != nil {
		response.InternalError(w, "Failed to generate token")
		return
	}

	response.Success(w, AuthResponse{
		User:  user,
		Token: token,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Token invalidation would require a blocklist or short expiry
	// For now, client-side removal is sufficient
	response.NoContent(w)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.Success(w, user)
}

func (h *AuthHandler) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(h.cfg.TokenExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}