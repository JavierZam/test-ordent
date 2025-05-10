package model

import "time"

type Product struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    Stock       int       `json:"stock"`
    CategoryID  int       `json:"category_id"`
    ImageURL    string    `json:"image_url"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type ProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"required,gte=0"`
	CategoryID  uint    `json:"category_id"`
	ImageURL    string  `json:"image_url"`
}

type ProductResponse struct {
    ID          int        `json:"id"`
    Name        string     `json:"name"`
    Description string     `json:"description"`
    Price       float64    `json:"price"`
    Stock       int        `json:"stock"`
    CategoryID  int        `json:"category_id"`
    ImageURL    string     `json:"image_url"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type ProductsResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int               `json:"total"`
}