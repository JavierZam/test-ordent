package repository

import (
	"database/sql"
	"errors"
	"time"

	"test-ordent/internal/model"
)

// OrderRepository defines order related database operations
type OrderRepository interface {
	Create(userID uint, totalAmount float64, shippingAddress string) (uint, error)
	AddOrderItem(orderID uint, productID uint, quantity int, price float64) error
	FindByID(id uint) (*model.Order, error)
	FindByUserID(userID uint) ([]model.OrderResponse, error)
	GetOrderItems(orderID uint) ([]model.OrderItemDetail, error)
    AddItem(orderID uint, productID uint, quantity int, price float64, subtotal float64) error
	CreateOrder(userID uint, total float64, shippingAddress string, items []model.OrderItem, cartID uint) (uint, error)
}

// PostgresOrderRepository implements OrderRepository with PostgreSQL
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *sql.DB) OrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Create creates a new order
func (r *PostgresOrderRepository) Create(userID uint, totalAmount float64, shippingAddress string) (uint, error) {
	var id uint
	err := r.db.QueryRow(`
		INSERT INTO orders (user_id, total_amount, status, shipping_address)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, totalAmount, "pending", shippingAddress).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// AddOrderItem adds an item to order
func (r *PostgresOrderRepository) AddOrderItem(orderID uint, productID uint, quantity int, price float64) error {
	_, err := r.db.Exec(`
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
	`, orderID, productID, quantity, price)
	return err
}

func (r *PostgresOrderRepository) FindByID(id uint) (*model.Order, error) {
    var order model.Order
    err := r.db.QueryRow(`
        SELECT id, user_id, total_amount, status, shipping_address, created_at, updated_at
        FROM orders WHERE id = $1
    `, id).Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.ShippingAddress, &order.CreatedAt, &order.UpdatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("order not found")
        }
        return nil, err
    }
    return &order, nil
}

// FindByUserID finds orders by user ID
func (r *PostgresOrderRepository) FindByUserID(userID uint) ([]model.OrderResponse, error) {
	rows, err := r.db.Query(`
		SELECT id, total_amount, status, shipping_address, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.OrderResponse
	for rows.Next() {
		var order model.OrderResponse
		var createdAt time.Time
		if err := rows.Scan(&order.ID, &order.TotalAmount, &order.Status, &order.ShippingAddress, &createdAt); err != nil {
			return nil, err
		}
		order.CreatedAt = createdAt
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// GetOrderItems gets order items with product details
func (r *PostgresOrderRepository) GetOrderItems(orderID uint) ([]model.OrderItemDetail, error) {
	rows, err := r.db.Query(`
		SELECT oi.product_id, p.name, oi.price, oi.quantity
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderItems []model.OrderItemDetail
	for rows.Next() {
		var item model.OrderItemDetail
		if err := rows.Scan(&item.ProductID, &item.Name, &item.Price, &item.Quantity); err != nil {
			return nil, err
		}
		item.Subtotal = item.Price * float64(item.Quantity)
		orderItems = append(orderItems, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orderItems, nil
}

func (r *PostgresOrderRepository) AddItem(orderID uint, productID uint, quantity int, price float64, subtotal float64) error {
    _, err := r.db.Exec(`
        INSERT INTO order_items (order_id, product_id, quantity, price, subtotal, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    `, orderID, productID, quantity, price, subtotal)
    return err
}

func (r *PostgresOrderRepository) CreateOrder(userID uint, total float64, shippingAddress string, items []model.OrderItem, cartID uint) (uint, error) {
    // Start transaction
    tx, err := r.db.Begin()
    if err != nil {
        return 0, err
    }
    defer tx.Rollback()
    
    // Create order
    var orderID uint
    err = tx.QueryRow(`
        INSERT INTO orders (user_id, total, shipping_address, status, created_at, updated_at)
        VALUES ($1, $2, $3, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING id
    `, userID, total, shippingAddress).Scan(&orderID)
    
    if err != nil {
        return 0, err
    }
    
    // Create order items
    for _, item := range items {
        _, err = tx.Exec(`
            INSERT INTO order_items (order_id, product_id, quantity, price, subtotal, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        `, orderID, item.ProductID, item.Quantity, item.Price, item.Subtotal)
        
        if err != nil {
            return 0, err
        }
        
        // Decrease product stock
        _, err = tx.Exec(`
            UPDATE products 
            SET stock = stock - $1, updated_at = CURRENT_TIMESTAMP 
            WHERE id = $2 AND stock >= $1
        `, item.Quantity, item.ProductID)
        
        if err != nil {
            return 0, err
        }
    }
    
    // Clear cart
    _, err = tx.Exec("DELETE FROM cart_items WHERE cart_id = $1", cartID)
    if err != nil {
        return 0, err
    }
    
    // Commit transaction
    if err = tx.Commit(); err != nil {
        return 0, err
    }
    
    return orderID, nil
}