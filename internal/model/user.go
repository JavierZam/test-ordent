package model

// User represents a user in the system
type User struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	FullName     string `json:"full_name"`
	Role         string `json:"role"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Username  string `json:"username" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FullName  string `json:"full_name" validate:"required"`
}

// UserResponse represents user data for response
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// RegisterResponse represents registration response
type RegisterResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error"`
}

type RegisterAdminRequest struct {
    Username    string `json:"username" validate:"required,min=3,max=50"`
    Email       string `json:"email" validate:"required,email"`
    Password    string `json:"password" validate:"required,min=6"`
    FullName    string `json:"full_name" validate:"required"`
    AdminSecret string `json:"admin_secret" validate:"required"`
}