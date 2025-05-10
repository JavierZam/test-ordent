package handler

import (
	"database/sql"
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

    var req model.AddToCartRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request"})
    }

    // Check if product exists and has enough stock
    stock, err := h.productRepo.GetStock(int(req.ProductID))
    if err != nil {
        if err.Error() == "product not found" {
            return c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Product not found"})
        }
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
    }

    if stock < req.Quantity {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Not enough stock"})
    }

    // Start transaction
    tx, err := h.db.Begin()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to begin transaction"})
    }
    defer tx.Rollback()

    // Get or create cart
    cart, err := h.cartRepo.FindByUserID(userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
    }

    var cartID uint
    if cart == nil {
        // Create new cart
        cartID, err = h.cartRepo.Create(userID)
        if err != nil {
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to create cart"})
        }
    } else {
        cartID = cart.ID
    }

    // Check if product already in cart
    cartItem, err := h.cartRepo.FindCartItemByProductID(cartID, req.ProductID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
    }

    var totalQuantity int
    if cartItem == nil {
        // Add new cart item
        err = h.cartRepo.AddItem(cartID, req.ProductID, req.Quantity)
        if err != nil {
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to add item to cart"})
        }
        totalQuantity = req.Quantity
    } else {
        // Update existing cart item
        newQuantity := cartItem.Quantity + req.Quantity
        if newQuantity > stock {
            return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Not enough stock"})
        }

        err = h.cartRepo.UpdateItemQuantity(cartItem.ID, newQuantity)
        if err != nil {
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to update cart item"})
        }
        totalQuantity = req.Quantity // Only decrease stock for the newly added quantity
    }

    // Decrease product stock
    err = h.productRepo.DecreaseStock(int(req.ProductID), totalQuantity)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to update product stock: " + err.Error()})
    }

    // Update cart last modified
    err = h.cartRepo.UpdateLastModified(cartID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to update cart"})
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to commit transaction"})
    }

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