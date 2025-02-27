package models

import "time"

type User struct {
	ID                int        `json:"id"`
	Email             string     `json:"email"`
	Password          string     `json:"-"`       // Never send password in JSON
	RoleID            *int       `json:"role_id"` // Added RoleID field
	ResetToken        *string    `json:"-"`
	ResetTokenExpires *time.Time `json:"-"`
	CreatedAt         time.Time  `json:"created_at"`
}

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ResetPasswordRequest struct {
	Email string `json:"email"`
}

type UpdatePasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
