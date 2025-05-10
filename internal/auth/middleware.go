package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"test-ordent/internal/model"
)

// JWTMiddleware handles JWT authentication
type JWTMiddleware struct {
	jwtSecret string
}

func NewJWTMiddleware(jwtSecret string) *JWTMiddleware {
    if jwtSecret == "" {
        log.Println("WARNING: Empty JWT secret provided to middleware")
    }
    return &JWTMiddleware{
        jwtSecret: jwtSecret,
    }
}

func (m *JWTMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Check if JWT secret is empty
        if m.jwtSecret == "" {
            log.Println("ERROR: JWT secret is empty!")
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Server configuration error"})
        }

        // Get token from header
        authHeader := c.Request().Header.Get("Authorization")
        fmt.Printf("Authorization header: %s\n", authHeader)
        
        if authHeader == "" {
            return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Authorization header is required"})
        }

        // Check if token is in correct format
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid authorization format, use 'Bearer {token}'"})
        }

        // Validate token
        token := parts[1]
        fmt.Printf("Token: %s\n", token[:10] + "...")
        fmt.Printf("JWT Secret length: %d\n", len(m.jwtSecret))  // Safer than printing part of secret
        
        claims, err := ValidateToken(token, m.jwtSecret)
        if err != nil {
            fmt.Printf("Token validation error: %v\n", err)
            return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid token"})
        }

        // Set user info to context
        fmt.Printf("Validated user ID: %d, role: %s\n", claims.UserID, claims.Role)
        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)

        return next(c)
    }
}
// RequireAdmin middleware requires admin role
func (m *JWTMiddleware) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// First require auth
		err := m.RequireAuth(func(c echo.Context) error {
			// Check user role
			role, ok := c.Get("user_role").(string)
			if !ok || role != "admin" {
				return c.JSON(http.StatusForbidden, model.ErrorResponse{Error: "Admin role required"})
			}
			return next(c)
		})(c)

		if err != nil {
			return err
		}

		return nil
	}
}