// Package models provides shared data models for all microservices.
package models

import "time"

// User represents a user in the system.
type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	FullName  string    `json:"full_name" db:"full_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// IsAdmin returns true if the user is an admin.
func (u *User) IsAdmin() bool {
	return u.Username == "admin"
}

// DisplayName returns the display name for the user.
func (u *User) DisplayName() string {
	if u.FullName != "" {
		return u.FullName
	}
	return u.Username
}
