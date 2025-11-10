package models

import (
	"database/sql"
	"time"
)

// UserRole represents user role enum
type UserRole string

const (
	RoleClient        UserRole = "CLIENT"
	RoleCleaner       UserRole = "CLEANER"
	RoleCompanyAdmin  UserRole = "COMPANY_ADMIN"
	RolePlatformAdmin UserRole = "PLATFORM_ADMIN"
)

// User represents a user in the system
type User struct {
	ID             string
	Phone          sql.NullString
	Email          sql.NullString
	FirstName      sql.NullString
	LastName       sql.NullString
	Role           UserRole
	IsActive       bool
	EmailVerified  bool
	PhoneVerified  bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByPhone finds a user by phone number
func (r *UserRepository) GetByPhone(phone string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(`
		SELECT id, phone, email, first_name, last_name, role, is_active,
		       email_verified, phone_verified, created_at, updated_at
		FROM users
		WHERE phone = $1
	`, phone).Scan(
		&user.ID, &user.Phone, &user.Email, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.EmailVerified, &user.PhoneVerified,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail finds a user by email address
func (r *UserRepository) GetByEmail(email string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(`
		SELECT id, phone, email, first_name, last_name, role, is_active,
		       email_verified, phone_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Phone, &user.Email, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.EmailVerified, &user.PhoneVerified,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID finds a user by ID
func (r *UserRepository) GetByID(id string) (*User, error) {
	user := &User{}
	err := r.db.QueryRow(`
		SELECT id, phone, email, first_name, last_name, role, is_active,
		       email_verified, phone_verified, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID, &user.Phone, &user.Email, &user.FirstName, &user.LastName,
		&user.Role, &user.IsActive, &user.EmailVerified, &user.PhoneVerified,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *User) error {
	return r.db.QueryRow(`
		INSERT INTO users (phone, email, first_name, last_name, role, email_verified, phone_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, user.Phone, user.Email, user.FirstName, user.LastName, user.Role, user.EmailVerified, user.PhoneVerified).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// Update updates a user
func (r *UserRepository) Update(user *User) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET phone = $2, email = $3, first_name = $4, last_name = $5, role = $6,
		    is_active = $7, email_verified = $8, phone_verified = $9
		WHERE id = $1
	`, user.ID, user.Phone, user.Email, user.FirstName, user.LastName, user.Role,
		user.IsActive, user.EmailVerified, user.PhoneVerified)
	return err
}
