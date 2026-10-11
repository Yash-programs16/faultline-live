package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"faultline-live/backend/internal/middleware"
	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/repositories"
	"faultline-live/backend/internal/response"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderHandler struct {
	repo        *repositories.OrderRepository
	productRepo *repositories.ProductRepository
	accountRepo *repositories.AccountRepository
	db          *pgxpool.Pool
}

func NewOrderHandler(orderRepo *repositories.OrderRepository, productRepo *repositories.ProductRepository, accountRepo *repositories.AccountRepository, db *pgxpool.Pool) *OrderHandler {
	return &OrderHandler{
		repo:        orderRepo,
		productRepo: productRepo,
		accountRepo: accountRepo,
		db:          db,
	}
}

func (h *OrderHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/orders", h.Create).Methods("POST")
	r.HandleFunc("/api/orders", h.List).Methods("GET")
	r.HandleFunc("/api/orders/{id}", h.Get).Methods("GET")
}

type CreateOrderRequest struct {
	Items         []OrderItemRequest `json:"items"`
	AddressID     int64              `json:"address_id"`
	PaymentMethod string             `json:"payment_method"` // card, pix, boleto
	CouponCode    string             `json:"coupon_code,omitempty"`
}

type OrderItemRequest struct {
	ProductID int64  `json:"product_id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Price     int64  `json:"price"` // in cents
	Image     string `json:"image"`
	Color     string `json:"color,omitempty"`
	Size      string `json:"size,omitempty"`
	Qty       int    `json:"qty"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if len(req.Items) == 0 {
		response.BadRequest(w, "Order must contain at least one item")
		return
	}

	validPaymentMethods := map[string]bool{"card": true, "pix": true, "boleto": true}
	if !validPaymentMethods[req.PaymentMethod] {
		response.BadRequest(w, "Invalid payment method")
		return
	}

	// Verify address belongs to user
	addresses, err := h.accountRepo.GetAddresses(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "Failed to verify address")
		return
	}
	addressValid := false
	for _, a := range addresses {
		if a.ID == req.AddressID {
			addressValid = true
			break
		}
	}
	if !addressValid {
		response.BadRequest(w, "Invalid address")
		return
	}

	// Validate items and check stock
	tx, err := h.db.Begin(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback(r.Context())

	items := make([]repositories.OrderItemParams, 0, len(req.Items))
	for _, item := range req.Items {
		if item.Qty <= 0 {
			response.BadRequest(w, "Invalid quantity")
			return
		}

		// Verify product exists and has stock
		product, err := h.productRepo.GetBySlug(r.Context(), item.Slug)
		if err != nil {
			response.BadRequest(w, "Product not found: "+item.Slug)
			return
		}

		if product.Stock < item.Qty {
			response.BadRequest(w, "Insufficient stock for "+product.Name)
			return
		}

		// Decrement stock
		if err := h.productRepo.DecrementStock(r.Context(), tx, product.ID, item.Qty); err != nil {
			response.BadRequest(w, "Insufficient stock for "+product.Name)
			return
		}

		color := item.Color
		if color == "" {
			color = ""
		}
		size := item.Size
		if size == "" {
			size = ""
		}

		items = append(items, repositories.OrderItemParams{
			ProductID: product.ID,
			Slug:      product.Slug,
			Name:      product.Name,
			Price:     product.Price, // Use server-side price
			Image:     productImages(product.Slug)[0],
			Color:     &color,
			Size:      &size,
			Qty:       item.Qty,
		})
	}

	params := repositories.CreateOrderParams{
		UserID:        userID,
		Items:         items,
		AddressID:     req.AddressID,
		PaymentMethod: req.PaymentMethod,
		CouponCode:    req.CouponCode,
	}

	order, err := h.repo.Create(r.Context(), tx, params)
	if err != nil {
		response.InternalError(w, "Failed to create order")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		response.InternalError(w, "Failed to commit order")
		return
	}

	// Fetch full order with items
	orderItems, _ := h.repo.GetItems(r.Context(), order.ID)
	type OrderResponse struct {
		*models.Order
		Items []models.OrderItem `json:"items"`
	}
	response.Created(w, OrderResponse{
		Order: order,
		Items: orderItems,
	})
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	page := 1
	pageSize := 10
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			page = v
		}
	}
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil {
			pageSize = v
		}
	}

	result, err := h.repo.GetByUser(r.Context(), userID, page, pageSize)
	if err != nil {
		response.InternalError(w, "Failed to fetch orders")
		return
	}

	response.Success(w, result)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return
	}

	vars := mux.Vars(r)
	orderID := vars["id"]

	order, err := h.repo.GetByID(r.Context(), userID, orderID)
	if err != nil {
		response.NotFound(w, "Order not found")
		return
	}

	items, _ := h.repo.GetItems(r.Context(), orderID)

	type OrderResponse struct {
		*models.Order
		Items []models.OrderItem `json:"items"`
	}

	response.Success(w, OrderResponse{
		Order: order,
		Items: items,
	})
}

// productImages generates deterministic image URLs from slug (matching frontend logic)
func productImages(slug string) []string {
	ids := []string{
		"1523275335684-37898b6baf30", "1491553895911-0055eca6402d", "1542291026-7eec264c27ff",
		"1585386959984-a4155224a1ad", "1560769629-975ec94e6a86", "1608231387042-66d1773070a5",
		"1441986300917-64674bd600d8", "1526170375885-4d8ecf77b99f", "1572635196237-14b3f281503f",
		"1503341504253-dff4815485f1", "1504204267155-aaad8e81290d",
	}
	h := 0
	for i := 0; i < len(slug); i++ {
		h = (h*31 + int(slug[i])) >> 0
	}
	result := make([]string, 4)
	for i := 0; i < 4; i++ {
		idx := (h + i*7) % len(ids)
		result[i] = "https://images.unsplash.com/photo-" + ids[idx] + "?w=600&h=600&fit=crop&q=70&auto=format"
	}
	return result
}
