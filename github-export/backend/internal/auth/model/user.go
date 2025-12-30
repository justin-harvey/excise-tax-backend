// Package model defines domain models for the authentication service.
package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID          int64      `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose in JSON
	FirstName   string     `json:"first_name" db:"first_name"`
	LastName    string     `json:"last_name" db:"last_name"`
	Phone       string     `json:"phone" db:"phone"`
	Role        string     `json:"role" db:"role"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Manufacturer represents manufacturer profile information
type Manufacturer struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	CompanyName  string    `json:"company_name" db:"company_name"`
	TaxID        string    `json:"tax_id" db:"tax_id"`
	LicenseNumber string   `json:"license_number" db:"license_number"`
	BusinessType string    `json:"business_type" db:"business_type"` // brewery, winery, distillery
	AddressLine1 string    `json:"address_line1" db:"address_line1"`
	AddressLine2 string    `json:"address_line2,omitempty" db:"address_line2"`
	City         string    `json:"city" db:"city"`
	State        string    `json:"state" db:"state"`
	ZipCode      string    `json:"zip_code" db:"zip_code"`
	Phone        string    `json:"phone" db:"phone"`
	Website      string    `json:"website,omitempty" db:"website"`
	IsApproved   bool      `json:"is_approved" db:"is_approved"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Session represents a user session stored in Redis
type Session struct {
	ID           string    `json:"id"`
	UserID       int64     `json:"user_id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	RefreshToken string    `json:"refresh_token"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// UserWithManufacturer combines user and manufacturer information
type UserWithManufacturer struct {
	User         *User         `json:"user"`
	Manufacturer *Manufacturer `json:"manufacturer,omitempty"`
}

// Role constants
const (
	RoleManufacturer = "manufacturer"
	RoleAdmin        = "admin"
	RoleSuperAdmin   = "super_admin"
	RoleReviewer     = "reviewer"
	RoleReadOnly     = "read_only"
)

// Permission constants
const (
	PermissionReadReports   = "read:reports"
	PermissionWriteReports  = "write:reports"
	PermissionApproveReports = "approve:reports"
	PermissionManageUsers   = "manage:users"
	PermissionManagePayments = "manage:payments"
	PermissionReadPayments  = "read:payments"
	PermissionWritePayments = "write:payments"
	PermissionSystemAdmin   = "system:admin"
)

// GetRolePermissions returns the permissions for a given role
func GetRolePermissions(role string) []string {
	switch role {
	case RoleSuperAdmin:
		return []string{
			PermissionSystemAdmin,
			PermissionManageUsers,
			PermissionApproveReports,
			PermissionWriteReports,
			PermissionReadReports,
			PermissionManagePayments,
			PermissionWritePayments,
			PermissionReadPayments,
		}
	case RoleAdmin:
		return []string{
			PermissionApproveReports,
			PermissionWriteReports,
			PermissionReadReports,
			PermissionManagePayments,
			PermissionWritePayments,
			PermissionReadPayments,
			PermissionManageUsers,
		}
	case RoleReviewer:
		return []string{
			PermissionApproveReports,
			PermissionReadReports,
			PermissionReadPayments,
		}
	case RoleManufacturer:
		return []string{
			PermissionWriteReports,
			PermissionReadReports,
			PermissionWritePayments,
			PermissionReadPayments,
		}
	case RoleReadOnly:
		return []string{
			PermissionReadReports,
			PermissionReadPayments,
		}
	default:
		return []string{}
	}
}

// IsAdmin checks if a role is an admin role
func IsAdmin(role string) bool {
	return role == RoleAdmin || role == RoleSuperAdmin
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	validRoles := []string{
		RoleManufacturer,
		RoleAdmin,
		RoleSuperAdmin,
		RoleReviewer,
		RoleReadOnly,
	}
	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

// PublicUser returns a user object safe for public display (without sensitive fields)
func (u *User) PublicUser() *User {
	return &User{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Phone:       u.Phone,
		Role:        u.Role,
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// FullName returns the user's full name
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return ""
	}
	return u.FirstName + " " + u.LastName
}
