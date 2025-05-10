package model

import "time"

type Cart struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Items     []CartItem `json:"items,omitempty"`
}

type CartItem struct {
	ID        uint      `json:"id"`
	CartID    uint      `json:"cart_id"`
	ProductID uint      `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Product   Product   `json:"product,omitempty"`
}

type AddToCartRequest struct {
    ProductID uint `json:"product_id" validate:"required"`
    Quantity  int  `json:"quantity" validate:"required,min=1"`
}

type CartResponse struct {
	ID    uint             `json:"id"`
	Items []CartItemDetail `json:"items"`
	Total float64          `json:"total"`
}

type CartItemDetail struct {
	ID        uint    `json:"id"`
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}