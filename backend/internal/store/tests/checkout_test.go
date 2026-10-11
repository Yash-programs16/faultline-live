package storetests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"faultline-live/backend/internal/store"
)

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "INFO: TEST_DATABASE_URL unset; integration tests will skip.")
		os.Exit(m.Run())
	}

	ctx := context.Background()
	st, err := store.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: connect to test database: %v\n", err)
		os.Exit(1)
	}

	if err := st.Migrate(ctx); err != nil {
		st.Close()
		fmt.Fprintf(os.Stderr, "ERROR: migrate test database: %v\n", err)
		os.Exit(1)
	}

	st.Close()
	os.Exit(m.Run())
}

func getTestStore(t *testing.T) *store.Store {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	ctx := context.Background()

	st, err := store.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	return st
}

func cleanupOrderItems(t *testing.T, st *store.Store, ctx context.Context, orderID string) {
	t.Helper()

	_, _ = st.Pool().Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, orderID)
	_, _ = st.Pool().Exec(ctx, `DELETE FROM orders WHERE id = $1`, orderID)
}

func TestCheckout_ValidProductPositiveQuantity(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	products, err := st.ListProducts(ctx)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("no products available for testing")
	}

	product := products[0]
	initialStock := product.Stock

	items := []store.CartItem{
		{ProductID: product.ID, Quantity: 1},
	}

	orderID, err := st.Checkout(ctx, items)
	if err != nil {
		t.Fatalf("checkout failed: %v", err)
	}
	if orderID == "" {
		t.Fatal("expected order ID, got empty string")
	}
	t.Cleanup(func() { cleanupOrderItems(t, st, ctx, orderID) })

	updatedProduct, err := st.GetProduct(ctx, product.ID)
	if err != nil {
		t.Fatalf("failed to get product after checkout: %v", err)
	}
	if updatedProduct.Stock != initialStock-1 {
		t.Errorf("expected stock %d, got %d", initialStock-1, updatedProduct.Stock)
	}
}

func TestCheckout_UnknownProductID(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	items := []store.CartItem{
		{ProductID: 999999, Quantity: 1},
	}

	_, err := st.Checkout(ctx, items)
	if err == nil {
		t.Fatal("expected error for unknown product ID, got nil")
	}
	if err.Error() == "" {
		t.Fatal("expected error message, got empty string")
	}
}

func TestCheckout_ZeroQuantity(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	products, err := st.ListProducts(ctx)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("no products available for testing")
	}

	items := []store.CartItem{
		{ProductID: products[0].ID, Quantity: 0},
	}

	_, err = st.Checkout(ctx, items)
	if err == nil {
		t.Fatal("expected error for zero quantity, got nil")
	}
}

func TestCheckout_NegativeQuantity(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	products, err := st.ListProducts(ctx)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("no products available for testing")
	}

	items := []store.CartItem{
		{ProductID: products[0].ID, Quantity: -1},
	}

	_, err = st.Checkout(ctx, items)
	if err == nil {
		t.Fatal("expected error for negative quantity, got nil")
	}
}

func TestCheckout_ExceedsStock(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	products, err := st.ListProducts(ctx)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("no products available for testing")
	}

	product := products[0]
	items := []store.CartItem{
		{ProductID: product.ID, Quantity: product.Stock + 1},
	}

	_, err = st.Checkout(ctx, items)
	if err == nil {
		t.Fatal("expected error for exceeding stock, got nil")
	}
}

func TestCheckout_EmptyCart(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	_, err := st.Checkout(ctx, []store.CartItem{})
	if err == nil {
		t.Fatal("expected error for empty cart, got nil")
	}
}

func TestCheckout_SuccessfulOrderCreation(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	products, err := st.ListProducts(ctx)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) < 2 {
		t.Fatal("need at least 2 products for testing")
	}

	product1 := products[0]
	product2 := products[1]
	initialStock1 := product1.Stock
	initialStock2 := product2.Stock

	items := []store.CartItem{
		{ProductID: product1.ID, Quantity: 1},
		{ProductID: product2.ID, Quantity: 2},
	}

	orderID, err := st.Checkout(ctx, items)
	if err != nil {
		t.Fatalf("checkout failed: %v", err)
	}
	if orderID == "" {
		t.Fatal("expected order ID, got empty string")
	}
	t.Cleanup(func() { cleanupOrderItems(t, st, ctx, orderID) })

	updatedProduct1, err := st.GetProduct(ctx, product1.ID)
	if err != nil {
		t.Fatalf("failed to get product1 after checkout: %v", err)
	}
	if updatedProduct1.Stock != initialStock1-1 {
		t.Errorf("expected product1 stock %d, got %d", initialStock1-1, updatedProduct1.Stock)
	}

	updatedProduct2, err := st.GetProduct(ctx, product2.ID)
	if err != nil {
		t.Fatalf("failed to get product2 after checkout: %v", err)
	}
	if updatedProduct2.Stock != initialStock2-2 {
		t.Errorf("expected product2 stock %d, got %d", initialStock2-2, updatedProduct2.Stock)
	}

	rows, err := st.Pool().Query(ctx,
		`SELECT product_id, quantity, unit_price_cents FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		t.Fatalf("failed to query order_items: %v", err)
	}
	defer rows.Close()

	itemCount := 0
	for rows.Next() {
		itemCount++

		var productID int64
		var quantity int64
		var unitPriceCents int64

		if err := rows.Scan(&productID, &quantity, &unitPriceCents); err != nil {
			t.Fatalf("failed to scan order item: %v", err)
		}

		switch productID {
		case product1.ID:
			if quantity != 1 {
				t.Errorf("expected product1 quantity 1, got %d", quantity)
			}
			if unitPriceCents != product1.PriceCents {
				t.Errorf("expected product1 price %d, got %d", product1.PriceCents, unitPriceCents)
			}
		case product2.ID:
			if quantity != 2 {
				t.Errorf("expected product2 quantity 2, got %d", quantity)
			}
			if unitPriceCents != product2.PriceCents {
				t.Errorf("expected product2 price %d, got %d", product2.PriceCents, unitPriceCents)
			}
		default:
			t.Errorf("unexpected product ID %d in order", productID)
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("error iterating order items: %v", err)
	}
	if itemCount != 2 {
		t.Errorf("expected 2 order items, got %d", itemCount)
	}
}

func TestCheckout_FailedCheckoutLeavesStockUnchanged(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	products, err := st.ListProducts(ctx)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) < 2 {
		t.Fatal("need at least 2 products for testing")
	}

	product1 := products[0]
	product2 := products[1]
	initialStock1 := product1.Stock
	initialStock2 := product2.Stock

	items := []store.CartItem{
		{ProductID: product1.ID, Quantity: 1},
		{ProductID: product2.ID, Quantity: product2.Stock + 100},
	}

	_, err = st.Checkout(ctx, items)
	if err == nil {
		t.Fatal("expected error for exceeding stock, got nil")
	}

	updatedProduct1, err := st.GetProduct(ctx, product1.ID)
	if err != nil {
		t.Fatalf("failed to get product1 after failed checkout: %v", err)
	}
	if updatedProduct1.Stock != initialStock1 {
		t.Errorf("expected product1 stock unchanged (%d), got %d", initialStock1, updatedProduct1.Stock)
	}

	updatedProduct2, err := st.GetProduct(ctx, product2.ID)
	if err != nil {
		t.Fatalf("failed to get product2 after failed checkout: %v", err)
	}
	if updatedProduct2.Stock != initialStock2 {
		t.Errorf("expected product2 stock unchanged (%d), got %d", initialStock2, updatedProduct2.Stock)
	}
}

func TestCartValidation_DuplicateProductIDs(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	products, err := st.ListProducts(ctx)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("no products available for testing")
	}

	product := products[0]
	items := []store.CartItem{
		{ProductID: product.ID, Quantity: 1},
		{ProductID: product.ID, Quantity: 2},
	}

	result, err := st.ValidateCart(ctx, items)
	if err != nil {
		t.Fatalf("ValidateCart failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected validation to fail for duplicate product IDs")
	}

	found := false
	for _, e := range result.Errors {
		if contains(e, "duplicate") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about duplicate product IDs, got: %v", result.Errors)
	}
}

func TestCartValidation_EmptyCart(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	result, err := st.ValidateCart(ctx, []store.CartItem{})
	if err != nil {
		t.Fatalf("ValidateCart failed: %v", err)
	}
	if result.Valid {
		t.Fatal("expected validation to fail for empty cart")
	}

	found := false
	for _, e := range result.Errors {
		if contains(e, "empty") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error about empty cart, got: %v", result.Errors)
	}
}

func TestCartValidation_ValidCart(t *testing.T) {
	st := getTestStore(t)
	defer st.Close()
	ctx := context.Background()

	products, err := st.ListProducts(ctx)
	if err != nil {
		t.Fatalf("failed to list products: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("no products available for testing")
	}

	product := products[0]
	items := []store.CartItem{
		{ProductID: product.ID, Quantity: 1},
	}

	result, err := st.ValidateCart(ctx, items)
	if err != nil {
		t.Fatalf("ValidateCart failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected validation to pass, got errors: %v", result.Errors)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 validated item, got %d", len(result.Items))
	}
	if result.Items[0].ProductID != product.ID {
		t.Errorf("expected product ID %d, got %d", product.ID, result.Items[0].ProductID)
	}
	if result.Items[0].Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", result.Items[0].Quantity)
	}
	if result.TotalCents != product.PriceCents {
		t.Errorf("expected total %d, got %d", product.PriceCents, result.TotalCents)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
