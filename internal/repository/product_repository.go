package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"test-ordent/internal/model"
)

// ProductRepository defines product related database operations
type ProductRepository interface {
	FindAll() ([]model.ProductResponse, error)
	FindByID(id int) (*model.ProductResponse, error)
	Create(product *model.ProductRequest) (*model.ProductResponse, error)
	Update(id int, product *model.ProductRequest) (*model.ProductResponse, error)
	Delete(id int) error
	ExistsByID(id int) (bool, error)
	DecreaseStock(id int, quantity int) error
	GetStock(id int) (int, error)
}

// PostgresProductRepository implements ProductRepository with PostgreSQL
type PostgresProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sql.DB) ProductRepository {
	return &PostgresProductRepository{db: db}
}

// FindAll finds all products
func (r *PostgresProductRepository) FindAll() ([]model.ProductResponse, error) {
	rows, err := r.db.Query("SELECT id, name, description, price, stock, category_id, image_url, created_at, updated_at FROM products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.ProductResponse
	for rows.Next() {
		var p model.ProductResponse
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CategoryID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// FindByID finds a product by ID
func (r *PostgresProductRepository) FindByID(id int) (*model.ProductResponse, error) {
	var p model.ProductResponse
	err := r.db.QueryRow(
		"SELECT id, name, description, price, stock, category_id, image_url, created_at, updated_at FROM products WHERE id = $1",
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CategoryID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return &p, nil
}

func (r *PostgresProductRepository) Create(product *model.ProductRequest) (*model.ProductResponse, error) {
    fmt.Println("Trying to create product:", product.Name)
    
    var p model.ProductResponse
    var updatedAt sql.NullTime  // Use sql.NullTime for nullable time columns
    
    err := r.db.QueryRow(
        `INSERT INTO products (name, description, price, stock, category_id, image_url, updated_at) 
        VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP) 
        RETURNING id, name, description, price, stock, category_id, image_url, created_at, updated_at`,
        product.Name, product.Description, product.Price, product.Stock, product.CategoryID, product.ImageURL,
    ).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CategoryID, &p.ImageURL, &p.CreatedAt, &updatedAt)

    if err != nil {
        fmt.Printf("Error creating product: %v\n", err)
        return nil, err
    }
    
    // Convert sql.NullTime to *time.Time
    if updatedAt.Valid {
        t := updatedAt.Time  // Tipe: time.Time
        p.UpdatedAt = &t     // Tipe: *time.Time 
    } else {
        t := time.Now()      // Tipe: time.Time
        p.UpdatedAt = &t     // Tipe: *time.Time
    }

    fmt.Println("Product created successfully with ID:", p.ID)
    return &p, nil
}

// Update updates a product
func (r *PostgresProductRepository) Update(id int, product *model.ProductRequest) (*model.ProductResponse, error) {
	var p model.ProductResponse
	err := r.db.QueryRow(
		`UPDATE products SET name = $1, description = $2, price = $3, stock = $4, category_id = $5, image_url = $6, updated_at = NOW() 
		WHERE id = $7 
		RETURNING id, name, description, price, stock, category_id, image_url, created_at, updated_at`,
		product.Name, product.Description, product.Price, product.Stock, product.CategoryID, product.ImageURL, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CategoryID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	return &p, nil
}

// Delete deletes a product
func (r *PostgresProductRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}

// ExistsByID checks if a product exists by ID
func (r *PostgresProductRepository) ExistsByID(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresProductRepository) DecreaseStock(id int, amount int) error {
    fmt.Printf("Executing SQL to decrease stock for product %d by %d\n", id, amount)
    
    result, err := r.db.Exec(
        "UPDATE products SET stock = stock - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND stock >= $1", 
        amount, id)
    
    if err != nil {
        fmt.Printf("SQL error in DecreaseStock: %v\n", err)
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        fmt.Printf("Error getting rows affected: %v\n", err)
        return err
    }
    
    fmt.Printf("DecreaseStock affected %d rows\n", rowsAffected)
    
    if rowsAffected == 0 {
        return errors.New("not enough stock or product not found")
    }
    
    return nil
}

func (r *PostgresProductRepository) GetStock(id int) (int, error) {
    var stock int
    err := r.db.QueryRow("SELECT stock FROM products WHERE id = $1", id).Scan(&stock)
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, errors.New("product not found")
        }
        return 0, err
    }
    return stock, nil
}

func (r *PostgresProductRepository) UpdateStock(id int, newStock int) error {
    _, err := r.db.Exec("UPDATE products SET stock = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", newStock, id)
    return err
}