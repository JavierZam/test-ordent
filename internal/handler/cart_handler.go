// internal/handler/cart_handler.go
package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"test-ordent/internal/model"
	"test-ordent/internal/repository"
)

type CartHandler struct {
    cartRepo    repository.CartRepository
    productRepo repository.ProductRepository
    db          *sql.DB 
}

func NewCartHandler(cartRepo repository.CartRepository, productRepo repository.ProductRepository, db *sql.DB) *CartHandler {
    return &CartHandler{
        cartRepo:    cartRepo,
        productRepo: productRepo,
        db:          db,
    }
}

// GetCart godoc
// @Summary Get user's cart
// @Description Get the current user's shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Success 200 {object} model.CartResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /cart [get]
func (h *CartHandler) GetCart(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	// Get or create cart
	cart, err := h.cartRepo.FindByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
	}

	var cartID uint
	if cart == nil {
		// Create new cart if not exists
		cartID, err = h.cartRepo.Create(userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to create cart"})
		}
	} else {
		cartID = cart.ID
	}

	// Get cart items
	items, err := h.cartRepo.GetCartItems(cartID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
	}

	// Calculate total
	var total float64
	for _, item := range items {
		total += item.Subtotal
	}

	return c.JSON(http.StatusOK, model.CartResponse{
		ID:    cartID,
		Items: items,
		Total: total,
	})
}

// AddItem godoc
// @Summary Add item to cart
// @Description Add a product to the user's shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Param item body model.AddToCartRequest true "Item to add"
// @Success 200 {object} model.CartResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /cart/items [post]
func (h *CartHandler) AddItem(c echo.Context) error {
    userID := c.Get("user_id").(uint)
    fmt.Printf("Adding item to cart for user ID: %d\n", userID)

    var req model.AddToCartRequest
    if err := c.Bind(&req); err != nil {
        fmt.Printf("Error binding request: %v\n", err)
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request"})
    }

    fmt.Printf("Add to cart request: ProductID=%d, Quantity=%d\n", req.ProductID, req.Quantity)

    // Check if product exists and has enough stock
    stock, err := h.productRepo.GetStock(int(req.ProductID))
    if err != nil {
        fmt.Printf("Error checking stock: %v\n", err)
        if err.Error() == "product not found" {
            return c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Product not found"})
        }
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
    }

    fmt.Printf("Current stock for product %d: %d\n", req.ProductID, stock)

    if stock < req.Quantity {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Not enough stock"})
    }

    // Start transaction
    tx, err := h.db.Begin()
    if err != nil {
        fmt.Printf("Error starting transaction: %v\n", err)
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to begin transaction"})
    }
    defer tx.Rollback()

    // Get or create cart
    cart, err := h.cartRepo.FindByUserID(userID)
    if err != nil {
        fmt.Printf("Error finding cart: %v\n", err)
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
    }

    var cartID uint
    if cart == nil {
        // Create new cart
        fmt.Println("Creating new cart")
        cartID, err = h.cartRepo.Create(userID)
        if err != nil {
            fmt.Printf("Error creating cart: %v\n", err)
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to create cart"})
        }
    } else {
        cartID = cart.ID
    }

    fmt.Printf("Using cart ID: %d\n", cartID)

    // Check if product already in cart
    cartItem, err := h.cartRepo.FindCartItemByProductID(cartID, req.ProductID)
    if err != nil {
        fmt.Printf("Error finding cart item: %v\n", err)
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
    }

    var totalQuantity int
    if cartItem == nil {
        // Add new cart item
        fmt.Println("Adding new item to cart")
        err = h.cartRepo.AddItem(cartID, req.ProductID, req.Quantity)
        if err != nil {
            fmt.Printf("Error adding item to cart: %v\n", err)
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to add item to cart"})
        }
        totalQuantity = req.Quantity
    } else {
        // Update existing cart item
        fmt.Printf("Updating existing cart item %d\n", cartItem.ID)
        newQuantity := cartItem.Quantity + req.Quantity
        if newQuantity > stock {
            return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Not enough stock"})
        }

        err = h.cartRepo.UpdateItemQuantity(cartItem.ID, newQuantity)
        if err != nil {
            fmt.Printf("Error updating cart item: %v\n", err)
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to update cart item"})
        }
        totalQuantity = req.Quantity // Only decrease stock for the newly added quantity
    }

    // Decrease product stock
    fmt.Printf("Decreasing stock for product %d by %d\n", req.ProductID, totalQuantity)
    err = h.productRepo.DecreaseStock(int(req.ProductID), totalQuantity)
    if err != nil {
        fmt.Printf("Error decreasing stock: %v\n", err)
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to update product stock: " + err.Error()})
    }

    // Update cart last modified
    err = h.cartRepo.UpdateLastModified(cartID)
    if err != nil {
        fmt.Printf("Error updating last modified: %v\n", err)
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to update cart"})
    }

    // Commit transaction
    fmt.Println("Committing transaction")
    if err = tx.Commit(); err != nil {
        fmt.Printf("Error committing transaction: %v\n", err)
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to commit transaction"})
    }

    fmt.Println("Successfully added item to cart and decreased stock")

    // Get updated cart
    return h.GetCart(c)
}

// RemoveItem godoc
// @Summary Remove item from cart
// @Description Remove an item from the user's shopping cart
// @Tags cart
// @Accept json
// @Produce json
// @Param id path int true "Cart Item ID"
// @Success 200 {object} model.CartResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /cart/items/{id} [delete]
func (h *CartHandler) RemoveItem(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	
	itemID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid item ID"})
	}

	// Get cart
	cart, err := h.cartRepo.FindByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
	}
	if cart == nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Cart not found"})
	}

	// Get cart item
	cartItem, err := h.cartRepo.FindCartItemByID(uint(itemID))
	if err != nil {
		if err.Error() == "cart item not found" {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Item not found in your cart"})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
	}

	// Check if item belongs to user's cart
	if cartItem.CartID != cart.ID {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Item not found in your cart"})
	}

	// Remove cart item
	err = h.cartRepo.RemoveItem(uint(itemID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to remove item from cart"})
	}

	// Update cart last modified
	err = h.cartRepo.UpdateLastModified(cart.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to update cart"})
	}

	// Get updated cart
	return h.GetCart(c)
}