package repositories

import (
	"context"
	"errors"
	"fmt"

	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/models"

	"github.com/jackc/pgx/v5"
)

type OrderRepository struct {
	db *database.DB
}

func NewOrderRepository(db *database.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

type CreateOrderParams struct {
	UserID           int64
	Items            []OrderItemParams
	AddressID        int64
	PaymentMethod    string
	CouponCode       string
}

type OrderItemParams struct {
	ProductID int64
	Slug      string
	Name      string
	Price     int64
	Image     string
	Color     *string
	Size      *string
	Qty       int
}

func (r *OrderRepository) Create(ctx context.Context, tx pgx.Tx, params CreateOrderParams) (*models.Order, error) {
	var subtotal int64
	for _, item := range params.Items {
		subtotal += item.Price * int64(item.Qty)
	}

	// Calculate discount
	var discount int64
	if params.CouponCode != "" {
		var percent float64
		err := tx.QueryRow(ctx, `
			SELECT percent FROM coupons WHERE code = $1 AND active = true AND valid_from <= NOW() AND (valid_until IS NULL OR valid_until > NOW())
		`, params.CouponCode).Scan(&percent)
		if err == nil {
			discount = int64(float64(subtotal) * percent)
		}
	}

	// Calculate shipping
	var shipping int64 = 2490 // 24.90 in cents
	if subtotal-discount >= 29900 { // 299.00 in cents
		shipping = 0
	}

	total := subtotal - discount + shipping

	orderID := models.NewOrderID()
	var order models.Order
	err := tx.QueryRow(ctx, `
		INSERT INTO orders (id, user_id, subtotal, discount, shipping, total, payment_method, status, shipping_address_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, subtotal, discount, shipping, total, payment_method, status, shipping_address_id, created_at, updated_at
	`, orderID, params.UserID, subtotal, discount, shipping, total, params.PaymentMethod,
		map[string]string{"card": "paid", "pix": "placed", "boleto": "placed"}[params.PaymentMethod], params.AddressID,
	).Scan(
		&order.ID, &order.UserID, &order.Subtotal, &order.Discount, &order.Shipping, &order.Total,
		&order.PaymentMethod, &order.Status, &order.ShippingAddressID, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	for _, item := range params.Items {
		_, err := tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, slug, name, price, image, color, size, qty)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, orderID, item.ProductID, item.Slug, item.Name, item.Price, item.Image, item.Color, item.Size, item.Qty)
		if err != nil {
			return nil, fmt.Errorf("create order item: %w", err)
		}
	}

	return &order, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, userID int64, orderID string) (*models.Order, error) {
	var order models.Order
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, subtotal, discount, shipping, total, payment_method, status, shipping_address_id, created_at, updated_at
		FROM orders WHERE id = $1 AND user_id = $2
	`, orderID, userID).Scan(
		&order.ID, &order.UserID, &order.Subtotal, &order.Discount, &order.Shipping, &order.Total,
		&order.PaymentMethod, &order.Status, &order.ShippingAddressID, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) GetItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, order_id, product_id, slug, name, price, image, color, size, qty
		FROM order_items WHERE order_id = $1 ORDER BY id
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("query order items: %w", err)
	}
	defer rows.Close()

	items := make([]models.OrderItem, 0)
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Slug, &item.Name, &item.Price, &item.Image, &item.Color, &item.Size, &item.Qty); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

type PaginatedOrders struct {
	Orders     []models.Order
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

func (r *OrderRepository) GetByUser(ctx context.Context, userID int64, page, pageSize int) (*PaginatedOrders, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	var total int64
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count orders: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, user_id, subtotal, discount, shipping, total, payment_method, status, shipping_address_id, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]models.Order, 0, pageSize)
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.Subtotal, &order.Discount, &order.Shipping, &order.Total,
			&order.PaymentMethod, &order.Status, &order.ShippingAddressID, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, order)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &PaginatedOrders{
		Orders:     orders,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}