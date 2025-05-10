package repository

import (
	"database/sql"
	"errors"

	"test-ordent/internal/model"
	"test-ordent/pkg/util"
)

type CartRepository interface {
	FindByUserID(userID uint) (*model.Cart, error)
	Create(userID uint) (uint, error)
	GetCartItems(cartID uint) ([]model.CartItemDetail, error)
	AddItem(cartID uint, productID uint, quantity int) error
	UpdateItemQuantity(itemID uint, quantity int) error
	RemoveItem(itemID uint) error
	ClearItems(cartID uint) error
	UpdateLastModified(cartID uint) error
	FindCartItemByID(itemID uint) (*model.CartItem, error)
	FindCartItemByProductID(cartID uint, productID uint) (*model.CartItem, error)
	ClearCart(cartID uint) error
}

type PostgresCartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &PostgresCartRepository{db: db}
}

func (r *PostgresCartRepository) FindByUserID(userID uint) (*model.Cart, error) {
    var cart model.Cart
    var updatedAt sql.NullTime
    
    err := r.db.QueryRow("SELECT id, user_id, created_at, updated_at FROM cart WHERE user_id = $1", userID).
        Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &updatedAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    
    cart.UpdatedAt = util.NullTimeToPointer(updatedAt)
    return &cart, nil
}

func (r *PostgresCartRepository) Create(userID uint) (uint, error) {
    var id uint
    err := r.db.QueryRow(`
        INSERT INTO cart (user_id, created_at, updated_at) 
        VALUES ($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
        RETURNING id
    `, userID).Scan(&id)
    if err != nil {
        return 0, err
    }
    return id, nil
}

func (r *PostgresCartRepository) GetCartItems(cartID uint) ([]model.CartItemDetail, error) {
	rows, err := r.db.Query(`
		SELECT ci.id, ci.product_id, p.name, p.price, ci.quantity
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1
	`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.CartItemDetail
	for rows.Next() {
		var item model.CartItemDetail
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Price, &item.Quantity); err != nil {
			return nil, err
		}
		item.Subtotal = item.Price * float64(item.Quantity)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PostgresCartRepository) AddItem(cartID uint, productID uint, quantity int) error {
	_, err := r.db.Exec("INSERT INTO cart_items (cart_id, product_id, quantity) VALUES ($1, $2, $3)",
		cartID, productID, quantity)
	return err
}

func (r *PostgresCartRepository) UpdateItemQuantity(itemID uint, quantity int) error {
	result, err := r.db.Exec("UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE id = $2",
		quantity, itemID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("cart item not found")
	}

	return nil
}

func (r *PostgresCartRepository) RemoveItem(itemID uint) error {
	result, err := r.db.Exec("DELETE FROM cart_items WHERE id = $1", itemID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("cart item not found")
	}

	return nil
}

func (r *PostgresCartRepository) ClearItems(cartID uint) error {
	_, err := r.db.Exec("DELETE FROM cart_items WHERE cart_id = $1", cartID)
	return err
}

func (r *PostgresCartRepository) UpdateLastModified(cartID uint) error {
	_, err := r.db.Exec("UPDATE cart SET updated_at = NOW() WHERE id = $1", cartID)
	return err
}

func (r *PostgresCartRepository) FindCartItemByID(itemID uint) (*model.CartItem, error) {
    var item model.CartItem
    var updatedAt sql.NullTime
    
    err := r.db.QueryRow(`
        SELECT id, cart_id, product_id, quantity, created_at, updated_at 
        FROM cart_items WHERE id = $1
    `, itemID).Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.CreatedAt, &updatedAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("cart item not found")
        }
        return nil, err
    }
    
    item.UpdatedAt = util.NullTimeToPointer(updatedAt)
    return &item, nil
}

func (r *PostgresCartRepository) FindCartItemByProductID(cartID uint, productID uint) (*model.CartItem, error) {
    var item model.CartItem
    var updatedAt sql.NullTime
    
    err := r.db.QueryRow(`
        SELECT id, cart_id, product_id, quantity, created_at, updated_at 
        FROM cart_items WHERE cart_id = $1 AND product_id = $2
    `, cartID, productID).Scan(&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.CreatedAt, &updatedAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    
    item.UpdatedAt = util.NullTimeToPointer(updatedAt)
    return &item, nil
}

func (r *PostgresCartRepository) ClearCart(cartID uint) error {
    _, err := r.db.Exec("DELETE FROM cart_items WHERE cart_id = $1", cartID)
    return err
}