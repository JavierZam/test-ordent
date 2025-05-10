package model

import "time"

type Order struct {
	ID              uint       `json:"id"`
	UserID          uint       `json:"user_id"`
	TotalAmount     float64    `json:"total_amount"`
	Status          string     `json:"status"`
	ShippingAddress string     `json:"shipping_address"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Items           []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
    ID        uint      `json:"id"`
    OrderID   uint      `json:"order_id"`
    ProductID uint      `json:"product_id"`
    Quantity  int       `json:"quantity"`
    Price     float64   `json:"price"`
    Subtotal  float64   `json:"subtotal"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	ShippingAddress string `json:"shipping_address" validate:"required"`
}

type OrderResponse struct {
	ID              uint              `json:"id"`
	TotalAmount     float64           `json:"total_amount"`
	Status          string            `json:"status"`
	ShippingAddress string            `json:"shipping_address"`
	CreatedAt       time.Time         `json:"created_at"`
	Items           []OrderItemDetail `json:"items"`
}

type OrderItemDetail struct {
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

type OrdersResponse struct {
	Orders []OrderResponse `json:"orders"`
}