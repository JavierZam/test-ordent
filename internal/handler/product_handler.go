package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"test-ordent/internal/model"
	"test-ordent/internal/repository"
)

// ProductHandler handles product related requests
type ProductHandler struct {
	productRepo repository.ProductRepository
}

// NewProductHandler creates a new product handler
func NewProductHandler(productRepo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
	}
}

// GetProducts godoc
// @Summary Get products list
// @Description Get list of all products with optional filtering
// @Tags products
// @Accept json
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} model.ProductsResponse
// @Router /products [get]
func (h *ProductHandler) GetProducts(c echo.Context) error {
	products, err := h.productRepo.FindAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
	}

	return c.JSON(http.StatusOK, model.ProductsResponse{
		Products: products,
		Total:    len(products),
	})
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Get a single product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} model.ProductResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid product ID"})
	}

	product, err := h.productRepo.FindByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Product not found"})
	}

	return c.JSON(http.StatusOK, product)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param product body model.ProductRequest true "Product data"
// @Success 201 {object} model.ProductResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c echo.Context) error {
    var req model.ProductRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request"})
    }
    
    // Validate request
    if req.Name == "" {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Product name is required"})
    }
    
    if req.Price <= 0 {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Price must be greater than 0"})
    }
    
    if req.Stock < 0 {
        return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Stock cannot be negative"})
    }

    product, err := h.productRepo.Create(&req)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to create product"})
    }

    return c.JSON(http.StatusCreated, product)
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body model.ProductRequest true "Product data"
// @Success 200 {object} model.ProductResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid product ID"})
	}

	var req model.ProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request"})
	}

	// Check if product exists
	exists, err := h.productRepo.ExistsByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Database error"})
	}

	if !exists {
		return c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Product not found"})
	}

	// Update product
	product, err := h.productRepo.Update(id, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to update product"})
	}

	return c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Security BearerAuth
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid product ID"})
	}

	// Delete product
	err = h.productRepo.Delete(id)
	if err != nil {
		if err.Error() == "product not found" {
			return c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Product not found"})
		}
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to delete product"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}