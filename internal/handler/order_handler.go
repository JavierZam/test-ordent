package handler

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"test-ordent/internal/model"
	"test-ordent/internal/repository"
)

type OrderHandler struct {
    orderRepo   repository.OrderRepository
    cartRepo    repository.CartRepository
    productRepo repository.ProductRepository
    db          *sql.DB  
}

func NewOrderHandler(orderRepo repository.OrderRepository, cartRepo repository.CartRepository, productRepo repository.ProductRepository, db *sql.DB) *OrderHandler {
    return &OrderHandler{
        orderRepo:   orderRepo,
        cartRepo:    cartRepo,
        productRepo: productRepo,
        db:          db,
    }
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order from cart items
// @Tags orders
// @Accept json
// @Produce json
// @Param order body model.CreateOrderRequest true "Order data"
// @Success 201 {object} model.OrderResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(c echo.Context) error {
    userID := c.Get("user_id").(uint)
    
    var req model.CreateOrderRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request"})
    }
    
    // Validate shipping address
    if req.ShippingAddress == "" {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Shipping address is required"})
    }
    
    // Get cart
    cart, err := h.cartRepo.FindByUserID(userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to get cart"})
    }
    
    if cart == nil {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Cart is empty"})
    }
    
    // Get cart items
    items, err := h.cartRepo.GetCartItems(cart.ID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to get cart items"})
    }
    
    if len(items) == 0 {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Cart is empty"})
    }
    
    // Calculate total and prepare order items
    var total float64
    orderItems := make([]model.OrderItem, 0, len(items))
    
    for _, item := range items {
        // Get product details
        product, err := h.productRepo.FindByID(int(item.ProductID))
        if err != nil {
            return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to get product"})
        }
        
        // Check stock
        if product.Stock < item.Quantity {
            return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: fmt.Sprintf("Not enough stock for product: %s", product.Name)})
        }
        
        // Calculate item total
        itemTotal := product.Price * float64(item.Quantity)
        total += itemTotal
        
        // Add to order items
        orderItems = append(orderItems, model.OrderItem{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     product.Price,
            Subtotal:  itemTotal,
        })
    }
    
    // Create order with transaction
    orderID, err := h.orderRepo.CreateOrder(userID, total, req.ShippingAddress, orderItems, cart.ID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to create order: " + err.Error()})
    }
    
    // Get order details
    order, err := h.orderRepo.FindByID(orderID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to get order"})
    }
    
    return c.JSON(http.StatusCreated, order)
}

// GetOrders godoc
// @Summary Get user orders
// @Description Get a list of user's orders
// @Tags orders
// @Accept json
// @Produce json
// @Success 200 {object} model.OrdersResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /orders [get]
func (h *OrderHandler) GetOrders(c echo.Context) error {
	userID := c.Get("user_id").(uint)

	// Get orders
	orders, err := h.orderRepo.FindByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
	}

	// Get order items for each order
	for i := range orders {
		orderItems, err := h.orderRepo.GetOrderItems(orders[i].ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
		}
		orders[i].Items = orderItems
	}

	return c.JSON(http.StatusOK, model.OrdersResponse{Orders: orders})
}