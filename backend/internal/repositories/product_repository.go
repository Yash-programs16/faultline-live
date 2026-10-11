package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/telemetry"

	"github.com/jackc/pgx/v5"
)

type ProductRepository struct {
	db  *database.DB
	tel *telemetry.Telemetry
}

func NewProductRepository(db *database.DB, tel *telemetry.Telemetry) *ProductRepository {
	return &ProductRepository{db: db, tel: tel}
}

func (r *ProductRepository) recordDBOperation(start time.Time) {
	if r.tel != nil {
		r.tel.RecordDBLatency(time.Since(start))
	}
}

type ProductFilters struct {
	Query     string
	Category  string
	Brands    []string
	MinPrice  *int64
	MaxPrice  *int64
	MinRating *float64
	Sort      string
	Page      int
	PageSize  int
}

type PaginatedProducts struct {
	Products   []models.Product
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

func (r *ProductRepository) GetAll(ctx context.Context, filters ProductFilters) (*PaginatedProducts, error) {
	started := time.Now()
	defer r.recordDBOperation(started)
	where := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1

	if filters.Query != "" {
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d OR p.brand ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+filters.Query+"%")
		argIdx++
	}
	if filters.Category != "" {
		where = append(where, fmt.Sprintf("c.slug = $%d", argIdx))
		args = append(args, filters.Category)
		argIdx++
	}
	if len(filters.Brands) > 0 {
		placeholders := make([]string, len(filters.Brands))
		for i := range filters.Brands {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, filters.Brands[i])
			argIdx++
		}
		where = append(where, fmt.Sprintf("p.brand IN (%s)", strings.Join(placeholders, ",")))
	}
	if filters.MinPrice != nil {
		where = append(where, fmt.Sprintf("p.price >= $%d", argIdx))
		args = append(args, *filters.MinPrice)
		argIdx++
	}
	if filters.MaxPrice != nil {
		where = append(where, fmt.Sprintf("p.price <= $%d", argIdx))
		args = append(args, *filters.MaxPrice)
		argIdx++
	}
	if filters.MinRating != nil {
		where = append(where, fmt.Sprintf("p.rating >= $%d", argIdx))
		args = append(args, *filters.MinRating)
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE %s
	`, whereClause)
	var total int64
	if err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count products: %w", err)
	}

	// Build order by
	orderBy := "p.created_at DESC"
	switch filters.Sort {
	case "price-asc":
		orderBy = "p.price ASC"
	case "price-desc":
		orderBy = "p.price DESC"
	case "newest":
		orderBy = "p.created_at DESC"
	case "rating":
		orderBy = "p.rating DESC"
	case "relevance":
		orderBy = "(p.bestseller::int + p.featured::int) DESC, p.rating DESC"
	}

	// Pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 20
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}
	offset := (filters.Page - 1) * filters.PageSize

	args = append(args, filters.PageSize, offset)
	limitArgIdx := argIdx
	offsetArgIdx := argIdx + 1

	dataQuery := fmt.Sprintf(`
		SELECT p.id, p.slug, p.name, p.description, p.price, p.price_old, p.category_id, c.slug, p.brand,
		       p.rating, p.reviews_count, p.stock, p.featured, p.bestseller, p.is_new, p.created_at, p.updated_at
		FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, limitArgIdx, offsetArgIdx)

	rows, err := r.db.Pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	products := make([]models.Product, 0, filters.PageSize)
	for rows.Next() {
		var p models.Product
		var priceOld *int64
		if err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.Description, &p.Price, &priceOld, &p.CategoryID, &p.CategorySlug,
			&p.Brand, &p.Rating, &p.ReviewsCount, &p.Stock, &p.Featured, &p.Bestseller, &p.IsNew,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		p.PriceOld = priceOld
		products = append(products, p)
	}

	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))
	return &PaginatedProducts{
		Products:   products,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *ProductRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	started := time.Now()
	defer r.recordDBOperation(started)
	var p models.Product
	var priceOld *int64
	err := r.db.Pool.QueryRow(ctx, `
		SELECT p.id, p.slug, p.name, p.description, p.price, p.price_old, p.category_id, c.slug, p.brand,
		       p.rating, p.reviews_count, p.stock, p.featured, p.bestseller, p.is_new, p.created_at, p.updated_at
		FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE p.slug = $1
	`, slug).Scan(
		&p.ID, &p.Slug, &p.Name, &p.Description, &p.Price, &priceOld, &p.CategoryID, &p.CategorySlug,
		&p.Brand, &p.Rating, &p.ReviewsCount, &p.Stock, &p.Featured, &p.Bestseller, &p.IsNew,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("get product: %w", err)
	}
	p.PriceOld = priceOld
	return &p, nil
}

func (r *ProductRepository) GetCategories(ctx context.Context) ([]models.Category, error) {
	started := time.Now()
	defer r.recordDBOperation(started)
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, slug, name, icon, description, created_at FROM categories ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Icon, &c.Description, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *ProductRepository) GetBrands(ctx context.Context) ([]string, error) {
	started := time.Now()
	defer r.recordDBOperation(started)
	rows, err := r.db.Pool.Query(ctx, `SELECT DISTINCT brand FROM products ORDER BY brand`)
	if err != nil {
		return nil, fmt.Errorf("query brands: %w", err)
	}
	defer rows.Close()

	brands := make([]string, 0)
	for rows.Next() {
		var brand string
		if err := rows.Scan(&brand); err != nil {
			return nil, fmt.Errorf("scan brand: %w", err)
		}
		brands = append(brands, brand)
	}
	return brands, nil
}

func (r *ProductRepository) GetVariants(ctx context.Context, productID int64) (*models.ProductVariant, error) {
	var variant models.ProductVariant
	err := r.db.Pool.QueryRow(ctx, `
		SELECT colors, sizes FROM product_variants WHERE product_id = $1
	`, productID).Scan(&variant.Colors, &variant.Sizes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &models.ProductVariant{Colors: []string{}, Sizes: []string{}}, nil
		}
		return nil, fmt.Errorf("get variants: %w", err)
	}
	variant.ProductID = productID
	return &variant, nil
}

func (r *ProductRepository) GetSpecs(ctx context.Context, productID int64) ([]models.ProductSpec, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT label, value FROM product_specs WHERE product_id = $1 ORDER BY id
	`, productID)
	if err != nil {
		return nil, fmt.Errorf("query specs: %w", err)
	}
	defer rows.Close()

	specs := make([]models.ProductSpec, 0)
	for rows.Next() {
		var s models.ProductSpec
		if err := rows.Scan(&s.Label, &s.Value); err != nil {
			return nil, fmt.Errorf("scan spec: %w", err)
		}
		s.ProductID = productID
		specs = append(specs, s)
	}
	return specs, nil
}

func (r *ProductRepository) GetReviews(ctx context.Context, productID int64) ([]models.ProductReview, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT author, rating, date, comment FROM product_reviews WHERE product_id = $1 ORDER BY date DESC LIMIT 10
	`, productID)
	if err != nil {
		return nil, fmt.Errorf("query reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]models.ProductReview, 0)
	for rows.Next() {
		var r models.ProductReview
		if err := rows.Scan(&r.Author, &r.Rating, &r.Date, &r.Comment); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		r.ProductID = productID
		reviews = append(reviews, r)
	}
	return reviews, nil
}

func (r *ProductRepository) DecrementStock(ctx context.Context, tx pgx.Tx, productID int64, qty int) error {
	result, err := tx.Exec(ctx, `
		UPDATE products SET stock = stock - $1, updated_at = NOW()
		WHERE id = $2 AND stock >= $1
	`, qty, productID)
	if err != nil {
		return fmt.Errorf("decrement stock: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("insufficient stock")
	}
	return nil
}
