package models

import "time"

type Role struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoleRequest is used for creating/updating roles
type RoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Permissions []int  `json:"permissions"` // Array of permission IDs
}

// RoleResponse is used for API responses
type RoleResponse struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       int       `json:"role_id"`
	PermissionID int       `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	UserID         int    `json:"user_id"`
	PermissionName string `json:"permission_name"`
}
