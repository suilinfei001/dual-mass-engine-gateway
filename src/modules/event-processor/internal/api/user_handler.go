package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github-hub/event-processor/internal/models"
	"github-hub/event-processor/internal/storage"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userStorage *storage.MySQLUserStorage
}

func NewUserHandler(userStorage *storage.MySQLUserStorage) *UserHandler {
	return &UserHandler{
		userStorage: userStorage,
	}
}

func (h *UserHandler) RegisterRoutes(router *mux.Router, authMiddleware func(http.Handler) http.Handler, adminMiddleware func(http.Handler) http.Handler) {
	router.HandleFunc("/api/auth/register", h.handleRegister).Methods("POST")
	router.HandleFunc("/api/auth/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/api/auth/logout", h.handleLogout).Methods("POST")
	router.HandleFunc("/api/auth/status", h.handleAuthStatus).Methods("GET")
	router.Handle("/api/auth/me", authMiddleware(http.HandlerFunc(h.handleGetCurrentUser))).Methods("GET")
	router.HandleFunc("/api/auth/forgot-password", h.handleForgotPassword).Methods("POST")
	router.Handle("/api/admin/users", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleSearchUsers)))).Methods("GET")
	router.Handle("/api/admin/users/{id}/password", authMiddleware(adminMiddleware(http.HandlerFunc(h.handleUpdateUserPassword)))).Methods("PUT")
}

func (h *UserHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Email is required",
		})
		return
	}

	if !strings.HasSuffix(strings.ToLower(req.Email), "@aishu.cn") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Email must end with @aishu.cn",
		})
		return
	}

	if req.Password == "" || len(req.Password) < 6 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Password must be at least 6 characters",
		})
		return
	}

	existingUser, err := h.userStorage.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Email already registered",
		})
		return
	}

	username := extractUsernameFromEmail(req.Email)
	existingUsername, err := h.userStorage.GetUserByUsername(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existingUsername != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Username already exists",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := models.NewUser(username, string(hashedPassword), models.RoleUser, req.Email)

	if err := h.userStorage.CreateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Registration successful",
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (h *UserHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Username and password are required",
		})
		return
	}

	user, err := h.userStorage.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid username or password",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid username or password",
		})
		return
	}

	session, err := h.userStorage.CreateSession(user.ID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session["session_id"].(string),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"role":     string(user.Role),
			"email":    user.Email,
		},
	})
}

func (h *UserHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		h.userStorage.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	})
}

func (h *UserHandler) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  false,
			"loggedIn": false,
		})
		return
	}

	session, err := h.userStorage.GetSessionWithUser(cookie.Value)
	if err != nil || session == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  false,
			"loggedIn": false,
		})
		return
	}

	userData := session["user"].(map[string]interface{})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"loggedIn": true,
		"user":     userData,
	})
}

func (h *UserHandler) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userStorage.GetUser(userID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"role":     string(user.Role),
			"email":    user.Email,
		},
	})
}

func (h *UserHandler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Please contact Yabo.sui@aishu.cn to reset your password",
	})
}

func (h *UserHandler) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	users, total, err := h.userStorage.GetUsersWithPaginationAndKeyword(page, pageSize, keyword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(users))
	for i, u := range users {
		result[i] = map[string]interface{}{
			"id":         u.ID,
			"username":   u.Username,
			"email":      u.Email,
			"role":       string(u.Role),
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
		"total":   total,
	})
}

func (h *UserHandler) handleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID := 0
	if _, err := regexp.MatchString(`^\d+$`, userIDStr); err == nil {
		var parseErr error
		userID, parseErr = parseInt(userIDStr)
		if parseErr != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
	}

	var req models.UserUpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" || len(req.NewPassword) < 6 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Password must be at least 6 characters",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := h.userStorage.UpdatePassword(userID, string(hashedPassword)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

func extractUsernameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return email
}

func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrInvalidInt
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}

var ErrInvalidInt = &InvalidIntError{}

type InvalidIntError struct{}

func (e *InvalidIntError) Error() string {
	return "invalid integer"
}
