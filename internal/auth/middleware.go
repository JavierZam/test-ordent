package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"test-ordent/internal/model"
)

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
        if m.jwtSecret == "" {
            log.Println("ERROR: JWT secret is empty!")
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Server configuration error"})
        }

        authHeader := c.Request().Header.Get("Authorization")
        
        if authHeader == "" {
            return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Authorization header is required"})
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid authorization format, use 'Bearer {token}'"})
        }

        token := parts[1]
        
        claims, err := ValidateToken(token, m.jwtSecret)
        if err != nil {
            return c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid token"})
        }

        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)

        return next(c)
    }
}

func (m *JWTMiddleware) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := m.RequireAuth(func(c echo.Context) error {
			role, ok := c.Get("role").(string)
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