// Package repository provides data access layer for authentication operations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/excise-tax-portal/backend/internal/auth/model"
	"github.com/excise-tax-portal/backend/pkg/database"
	"github.com/excise-tax-portal/backend/pkg/utils"

	"github.com/jackc/pgx/v5"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when attempting to create a duplicate user
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository handles user data access operations
type UserRepository struct {
	db *database.PostgresDB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *database.PostgresDB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(ctx context.Context, user *model.User, password string) (*model.User, error) {
	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, phone, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, email, first_name, last_name, phone, role, is_active, last_login_at, created_at, updated_at
	`

	now := time.Now()
	row := r.db.QueryRow(ctx, query,
		user.Email,
		hashedPassword,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.Role,
		user.IsActive,
		now,
		now,
	)

	createdUser := &model.User{}
	err = row.Scan(
		&createdUser.ID,
		&createdUser.Email,
		&createdUser.FirstName,
		&createdUser.LastName,
		&createdUser.Phone,
		&createdUser.Role,
		&createdUser.IsActive,
		&createdUser.LastLoginAt,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "duplicate key value violates unique constraint \"users_email_key\"" {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

// GetUserByEmail retrieves a user by email address
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, role, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &model.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Role,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, role, is_active, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &model.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Phone,
		&user.Role,
		&user.IsActive,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// UpdateUser updates user information
func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, phone = $3, role = $4, is_active = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.Exec(ctx, query,
		user.FirstName,
		user.LastName,
		user.Phone,
		user.Role,
		user.IsActive,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `
		UPDATE users
		SET last_login_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, newPassword string) error {
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(ctx, query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ValidateUserCredentials validates user credentials and returns the user if valid
func (r *UserRepository) ValidateUserCredentials(ctx context.Context, email, password string) (*model.User, error) {
	user, err := r.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Compare passwords
	if err := utils.ComparePasswords(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// CreateManufacturer creates a new manufacturer profile
func (r *UserRepository) CreateManufacturer(ctx context.Context, manufacturer *model.Manufacturer) (*model.Manufacturer, error) {
	query := `
		INSERT INTO manufacturers (
			user_id, company_name, tax_id, license_number, business_type,
			address_line1, address_line2, city, state, zip_code,
			phone, website, is_approved, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, user_id, company_name, tax_id, license_number, business_type,
		          address_line1, address_line2, city, state, zip_code,
		          phone, website, is_approved, created_at, updated_at
	`

	now := time.Now()
	row := r.db.QueryRow(ctx, query,
		manufacturer.UserID,
		manufacturer.CompanyName,
		manufacturer.TaxID,
		manufacturer.LicenseNumber,
		manufacturer.BusinessType,
		manufacturer.AddressLine1,
		manufacturer.AddressLine2,
		manufacturer.City,
		manufacturer.State,
		manufacturer.ZipCode,
		manufacturer.Phone,
		manufacturer.Website,
		manufacturer.IsApproved,
		now,
		now,
	)

	created := &model.Manufacturer{}
	err := row.Scan(
		&created.ID,
		&created.UserID,
		&created.CompanyName,
		&created.TaxID,
		&created.LicenseNumber,
		&created.BusinessType,
		&created.AddressLine1,
		&created.AddressLine2,
		&created.City,
		&created.State,
		&created.ZipCode,
		&created.Phone,
		&created.Website,
		&created.IsApproved,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create manufacturer: %w", err)
	}

	return created, nil
}

// GetManufacturerByUserID retrieves manufacturer information by user ID
func (r *UserRepository) GetManufacturerByUserID(ctx context.Context, userID int64) (*model.Manufacturer, error) {
	query := `
		SELECT id, user_id, company_name, tax_id, license_number, business_type,
		       address_line1, address_line2, city, state, zip_code,
		       phone, website, is_approved, created_at, updated_at
		FROM manufacturers
		WHERE user_id = $1
	`

	manufacturer := &model.Manufacturer{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&manufacturer.ID,
		&manufacturer.UserID,
		&manufacturer.CompanyName,
		&manufacturer.TaxID,
		&manufacturer.LicenseNumber,
		&manufacturer.BusinessType,
		&manufacturer.AddressLine1,
		&manufacturer.AddressLine2,
		&manufacturer.City,
		&manufacturer.State,
		&manufacturer.ZipCode,
		&manufacturer.Phone,
		&manufacturer.Website,
		&manufacturer.IsApproved,
		&manufacturer.CreatedAt,
		&manufacturer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Manufacturer profile is optional
		}
		return nil, fmt.Errorf("failed to get manufacturer: %w", err)
	}

	return manufacturer, nil
}

// GetUserWithManufacturer retrieves a user with their manufacturer profile
func (r *UserRepository) GetUserWithManufacturer(ctx context.Context, userID int64) (*model.UserWithManufacturer, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	manufacturer, err := r.GetManufacturerByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &model.UserWithManufacturer{
		User:         user,
		Manufacturer: manufacturer,
	}, nil
}
