package models

import "time"

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Role      UserRole  `json:"role"`
	Email     string    `json:"email,omitempty"`
	FullName  string    `json:"full_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserUpdatePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func NewUser(username, hashedPassword string, role UserRole, email string) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Password:  hashedPassword,
		Role:      role,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
