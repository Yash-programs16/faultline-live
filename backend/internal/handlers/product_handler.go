package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"faultline-live/backend/internal/models"
	"faultline-live/backend/internal/repositories"
	"faultline-live/backend/internal/response"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	repo *repositories.ProductRepository
}

func NewProductHandler(repo *repositories.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/products", h.List).Methods("GET")
	r.HandleFunc("/api/products/{slug}", h.Get).Methods("GET")
	r.HandleFunc("/api/categories", h.Categories).Methods("GET")
	r.HandleFunc("/api/brands", h.Brands).Methods("GET")
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	filters := repositories.ProductFilters{
		Query:    r.URL.Query().Get("q"),
		Category: r.URL.Query().Get("category"),
		Sort:     r.URL.Query().Get("sort"),
	}

	if brands := r.URL.Query().Get("brands"); brands != "" {
		filters.Brands = strings.Split(brands, ",")
	}
	if minPrice := r.URL.Query().Get("minPrice"); minPrice != "" {
		if v, err := strconv.ParseInt(minPrice, 10, 64); err == nil {
			filters.MinPrice = &v
		}
	}
	if maxPrice := r.URL.Query().Get("maxPrice"); maxPrice != "" {
		if v, err := strconv.ParseInt(maxPrice, 10, 64); err == nil {
			filters.MaxPrice = &v
		}
	}
	if minRating := r.URL.Query().Get("minRating"); minRating != "" {
		if v, err := strconv.ParseFloat(minRating, 64); err == nil {
			filters.MinRating = &v
		}
	}
	if page := r.URL.Query().Get("page"); page != "" {
		if v, err := strconv.Atoi(page); err == nil {
			filters.Page = v
		}
	}
	if pageSize := r.URL.Query().Get("pageSize"); pageSize != "" {
		if v, err := strconv.Atoi(pageSize); err == nil {
			filters.PageSize = v
		}
	}

	result, err := h.repo.GetAll(r.Context(), filters)
	if err != nil {
		response.InternalError(w, "Failed to fetch products")
		return
	}

	response.Success(w, result)
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	product, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		response.NotFound(w, "Product not found")
		return
	}

	// Fetch related data
	variants, _ := h.repo.GetVariants(r.Context(), product.ID)
	specs, _ := h.repo.GetSpecs(r.Context(), product.ID)
	reviews, _ := h.repo.GetReviews(r.Context(), product.ID)

	type ProductDetail struct {
		*models.Product
		Variants *models.ProductVariant `json:"variants,omitempty"`
		Specs    []models.ProductSpec   `json:"specs,omitempty"`
		Reviews  []models.ProductReview `json:"reviews,omitempty"`
	}

	detail := ProductDetail{
		Product:  product,
		Variants: variants,
		Specs:    specs,
		Reviews:  reviews,
	}

	response.Success(w, detail)
}

func (h *ProductHandler) Categories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetCategories(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to fetch categories")
		return
	}
	response.Success(w, categories)
}

func (h *ProductHandler) Brands(w http.ResponseWriter, r *http.Request) {
	brands, err := h.repo.GetBrands(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to fetch brands")
		return
	}
	response.Success(w, brands)
}
