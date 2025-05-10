package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"test-ordent/config"
	_ "test-ordent/docs"
	"test-ordent/internal/auth"
	"test-ordent/internal/database"
	"test-ordent/internal/handler"
	"test-ordent/internal/model"
	"test-ordent/internal/repository"
	"test-ordent/pkg/logger"
)

// @title E-Commerce API
// @version 1.0
// @description API for E-Commerce platform
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email ini_email@ordent.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if cfg.Auth.JWTSecret == "" {
		log.Fatal("JWT secret cannot be empty")
	}

	// Initialize logger
	logger := initLogger(cfg.Server.Debug)

	// Initialize database connection
    db, err := database.NewPostgresConnection(cfg.Database)
    if err != nil {
        logger.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

	// Initialize repositories
    userRepo := repository.NewUserRepository(db)
    productRepo := repository.NewProductRepository(db)
    cartRepo := repository.NewCartRepository(db)
    orderRepo := repository.NewOrderRepository(db)

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize JWT middleware
	jwtMiddleware := auth.NewJWTMiddleware(cfg.Auth.JWTSecret)

	// Register routes
	api := e.Group("/api")
	
	// Auth routes
	authHandler := handler.NewAuthHandler(userRepo, cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/admin-register", authHandler.RegisterAdmin)

	// Product routes
	productHandler := handler.NewProductHandler(productRepo)
	api.GET("/products", productHandler.GetProducts)
	api.GET("/products/:id", productHandler.GetProduct)
	api.POST("/products", productHandler.CreateProduct, jwtMiddleware.RequireAdmin)
	api.PUT("/products/:id", productHandler.UpdateProduct, jwtMiddleware.RequireAdmin)
	api.DELETE("/products/:id", productHandler.DeleteProduct, jwtMiddleware.RequireAdmin)

	// Cart routes
	cartHandler := handler.NewCartHandler(cartRepo, productRepo, db)
	api.GET("/cart", cartHandler.GetCart, jwtMiddleware.RequireAuth)
	api.POST("/cart/items", cartHandler.AddItem, jwtMiddleware.RequireAuth)
	api.DELETE("/cart/items/:id", cartHandler.RemoveItem, jwtMiddleware.RequireAuth)

	// Order routes
    orderHandler := handler.NewOrderHandler(orderRepo, cartRepo, productRepo, db)
    api.POST("/orders", orderHandler.CreateOrder, jwtMiddleware.RequireAuth)
    api.GET("/orders", orderHandler.GetOrders, jwtMiddleware.RequireAuth)

	// Categories routes
	api.GET("/categories", func(c echo.Context) error {
		rows, err := db.Query("SELECT id, name, description FROM categories")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to query categories"})
		}
		defer rows.Close()
		
		var categories []map[string]interface{}
		for rows.Next() {
			var id int
			var name, description string
			if err := rows.Scan(&id, &name, &description); err != nil {
				return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to scan category"})
			}
			categories = append(categories, map[string]interface{}{
				"id": id,
				"name": name,
				"description": description,
			})
		}
		
		return c.JSON(http.StatusOK, map[string]interface{}{
			"categories": categories,
		})
	})

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Start server
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info(fmt.Sprintf("Server starting on %s", serverAddr))
	if err := e.Start(serverAddr); err != http.ErrServerClosed {
		logger.Fatal("Failed to start server:", err)
	}
}

// loadConfig loads application configuration
func loadConfig() (*config.Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.yaml"
	}
	
	return config.LoadConfig(configPath)
}

// initLogger initializes the logger
func initLogger(debug bool) *logger.Logger {
	return logger.NewLogger(debug)
}