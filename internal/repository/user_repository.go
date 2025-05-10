package repository

import (
	"database/sql"
	"errors"

	"test-ordent/internal/model"
)

// UserRepository defines user related database operations
type UserRepository interface {
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) (uint, error)
	ExistsByUsernameOrEmail(username, email string) (bool, error)
}

// PostgresUserRepository implements UserRepository with PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

// FindByID finds a user by ID
func (r *PostgresUserRepository) FindByID(id uint) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow("SELECT id, username, email, password_hash, full_name, role FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// FindByUsername finds a user by username
func (r *PostgresUserRepository) FindByUsername(username string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow("SELECT id, username, email, password_hash, full_name, role FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// FindByEmail finds a user by email
func (r *PostgresUserRepository) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow("SELECT id, username, email, password_hash, full_name, role FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// Create creates a new user
func (r *PostgresUserRepository) Create(user *model.User) (uint, error) {
	var id uint
	err := r.db.QueryRow("INSERT INTO users (username, email, password_hash, full_name, role) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Username, user.Email, user.PasswordHash, user.FullName, user.Role).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ExistsByUsernameOrEmail checks if a user exists by username or email
func (r *PostgresUserRepository) ExistsByUsernameOrEmail(username, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 OR email = $2)", username, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}