package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"test-ordent/internal/auth"
	"test-ordent/internal/model"
	"test-ordent/internal/repository"
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	tokenExpiry time.Duration
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo repository.UserRepository, jwtSecret string, tokenExpiry time.Duration) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

// Login godoc
// @Summary Login user
// @Description Login with username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param login body model.LoginRequest true "Login credentials"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request"})
	}

	// Find user by username
	user, err := h.userRepo.FindByUsername(req.Username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid credentials"})
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid credentials"})
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Role, h.jwtSecret, h.tokenExpiry)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to generate token"})
	}
	log.Printf("Generated token for user %d with role %s", user.ID, user.Role)

	return c.JSON(http.StatusOK, model.LoginResponse{
		Token: token,
		User: model.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
		},
	})
}

// Register godoc
// @Summary Register a new user
// @Description Register with username, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param register body model.RegisterRequest true "Registration data"
// @Success 201 {object} model.RegisterResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request"})
	}

	// Check if username or email already exists
	exists, err := h.userRepo.ExistsByUsernameOrEmail(req.Username, req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
	}

	if exists {
		return c.JSON(http.StatusConflict, model.ErrorResponse{Error: "Username or email already exists"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to hash password"})
	}

	// Create new user
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Role:         "customer",
	}

	userID, err := h.userRepo.Create(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to create user"})
	}

	// Generate token
	token, err := auth.GenerateToken(userID, "customer", h.jwtSecret, h.tokenExpiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to generate token"})
	}

	return c.JSON(http.StatusCreated, model.RegisterResponse{
		Token: token,
		User: model.UserResponse{
			ID:       userID,
			Username: req.Username,
			Role:     "customer",
		},
	})
}

// RegisterAdmin godoc
// @Summary Register a new admin
// @Description Register an admin with secret code
// @Tags auth
// @Accept json
// @Produce json
// @Param register body model.RegisterAdminRequest true "Admin Registration data"
// @Success 201 {object} model.RegisterResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse
// @Router /auth/admin-register [post]
func (h *AuthHandler) RegisterAdmin(c echo.Context) error {
    var req model.RegisterAdminRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request"})
    }

    // Verify admin secret
    if req.AdminSecret != "thisisaverysecretkey" {
        return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid admin secret"})
    }

    // Check if username or email already exists
    exists, err := h.userRepo.ExistsByUsernameOrEmail(req.Username, req.Email)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
    }

    if exists {
        return c.JSON(http.StatusConflict, model.ErrorResponse{Error: "Username or email already exists"})
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to hash password"})
    }

    // Create new admin user
    user := &model.User{
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: string(hashedPassword),
        FullName:     req.FullName,
        Role:         "admin", // Set role to admin
    }

    userID, err := h.userRepo.Create(user)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to create user"})
    }

    // Generate token
    token, err := auth.GenerateToken(userID, "admin", h.jwtSecret, h.tokenExpiry)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to generate token"})
    }

    return c.JSON(http.StatusCreated, model.RegisterResponse{
        Token: token,
        User: model.UserResponse{
            ID:       userID,
            Username: req.Username,
            Role:     "admin",
        },
    })
}