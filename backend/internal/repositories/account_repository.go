package repositories

import (
	"context"
	"errors"
	"fmt"

	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/models"

	"github.com/jackc/pgx/v5"
)

type AccountRepository struct {
	db *database.DB
}

func NewAccountRepository(db *database.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Addresses

func (r *AccountRepository) GetAddresses(ctx context.Context, userID int64) ([]models.Address, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, user_id, label, recipient, street, number, complement, district, city, state, zip, is_default, created_at
		FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query addresses: %w", err)
	}
	defer rows.Close()

	addresses := make([]models.Address, 0)
	for rows.Next() {
		var a models.Address
		var complement *string
		if err := rows.Scan(&a.ID, &a.UserID, &a.Label, &a.Recipient, &a.Street, &a.Number, &complement, &a.District, &a.City, &a.State, &a.Zip, &a.IsDefault, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan address: %w", err)
		}
		a.Complement = complement
		addresses = append(addresses, a)
	}
	return addresses, nil
}

func (r *AccountRepository) CreateAddress(ctx context.Context, userID int64, addr models.Address) (*models.Address, error) {
	// If this is the first address or isDefault is true, unset other defaults
	if addr.IsDefault {
		_, _ = r.db.Pool.Exec(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1`, userID)
	} else {
		// Check if this is the first address
		var count int
		r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM addresses WHERE user_id = $1`, userID).Scan(&count)
		if count == 0 {
			addr.IsDefault = true
		}
	}

	var created models.Address
	var complement *string
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO addresses (user_id, label, recipient, street, number, complement, district, city, state, zip, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, label, recipient, street, number, complement, district, city, state, zip, is_default, created_at
	`, userID, addr.Label, addr.Recipient, addr.Street, addr.Number, addr.Complement, addr.District, addr.City, addr.State, addr.Zip, addr.IsDefault).
		Scan(&created.ID, &created.UserID, &created.Label, &created.Recipient, &created.Street, &created.Number, &complement, &created.District, &created.City, &created.State, &created.Zip, &created.IsDefault, &created.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create address: %w", err)
	}
	created.Complement = complement
	return &created, nil
}

func (r *AccountRepository) UpdateAddress(ctx context.Context, userID, addressID int64, updates map[string]interface{}) (*models.Address, error) {
	if len(updates) == 0 {
		return nil, errors.New("no updates provided")
	}

	// If setting is_default=true, unset others
	if isDefault, ok := updates["is_default"].(bool); ok && isDefault {
		_, _ = r.db.Pool.Exec(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1 AND id != $2`, userID, addressID)
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+2)
	argIdx := 1
	for key, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", key, argIdx))
		args = append(args, value)
		argIdx++
	}
	args = append(args, userID, addressID)

	query := fmt.Sprintf(`
		UPDATE addresses SET %s WHERE user_id = $%d AND id = $%d
		RETURNING id, user_id, label, recipient, street, number, complement, district, city, state, zip, is_default, created_at
	`, join(setParts, ", "), argIdx, argIdx+1)

	var updated models.Address
	var complement *string
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(
		&updated.ID, &updated.UserID, &updated.Label, &updated.Recipient, &updated.Street, &updated.Number, &complement,
		&updated.District, &updated.City, &updated.State, &updated.Zip, &updated.IsDefault, &updated.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("address not found")
		}
		return nil, fmt.Errorf("update address: %w", err)
	}
	updated.Complement = complement
	return &updated, nil
}

func (r *AccountRepository) DeleteAddress(ctx context.Context, userID, addressID int64) error {
	result, err := r.db.Pool.Exec(ctx, `DELETE FROM addresses WHERE user_id = $1 AND id = $2`, userID, addressID)
	if err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("address not found")
	}
	// If deleted was default, set another as default
	var newDefault int64
	err = r.db.Pool.QueryRow(ctx, `SELECT id FROM addresses WHERE user_id = $1 ORDER BY created_at LIMIT 1`, userID).Scan(&newDefault)
	if err == nil {
		_, _ = r.db.Pool.Exec(ctx, `UPDATE addresses SET is_default = true WHERE id = $1`, newDefault)
	}
	return nil
}

// Wishlist

func (r *AccountRepository) GetWishlist(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := r.db.Pool.Query(ctx, `SELECT product_id FROM wishlist_items WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query wishlist: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan wishlist: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *AccountRepository) AddToWishlist(ctx context.Context, userID, productID int64) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO wishlist_items (user_id, product_id) VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING
	`, userID, productID)
	if err != nil {
		return fmt.Errorf("add to wishlist: %w", err)
	}
	return nil
}

func (r *AccountRepository) RemoveFromWishlist(ctx context.Context, userID, productID int64) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM wishlist_items WHERE user_id = $1 AND product_id = $2`, userID, productID)
	if err != nil {
		return fmt.Errorf("remove from wishlist: %w", err)
	}
	return nil
}

func join(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	result := elems[0]
	for _, e := range elems[1:] {
		result += sep + e
	}
	return result
}