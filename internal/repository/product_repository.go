package repository

import (
	"database/sql"
	"errors"
	"time"

	"test-ordent/internal/model"
)

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

type PostgresProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &PostgresProductRepository{db: db}
}

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
    var p model.ProductResponse
    var updatedAt sql.NullTime 
    
    err := r.db.QueryRow(
        `INSERT INTO products (name, description, price, stock, category_id, image_url, updated_at) 
        VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP) 
        RETURNING id, name, description, price, stock, category_id, image_url, created_at, updated_at`,
        product.Name, product.Description, product.Price, product.Stock, product.CategoryID, product.ImageURL,
    ).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CategoryID, &p.ImageURL, &p.CreatedAt, &updatedAt)

    if err != nil {
        return nil, err
    }
    
    if updatedAt.Valid {
        t := updatedAt.Time
        p.UpdatedAt = &t
    } else {
        t := time.Now()
        p.UpdatedAt = &t
    }

    return &p, nil
}

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

func (r *PostgresProductRepository) ExistsByID(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PostgresProductRepository) DecreaseStock(id int, amount int) error {
    result, err := r.db.Exec(
        "UPDATE products SET stock = stock - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND stock >= $1", 
        amount, id)
    
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
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