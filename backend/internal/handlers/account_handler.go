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
)

type AccountHandler struct {
	repo *repositories.AccountRepository
}

func NewAccountHandler(repo *repositories.AccountRepository) *AccountHandler {
	return &AccountHandler{repo: repo}
}

func (h *AccountHandler) RegisterRoutes(r *mux.Router) {
	// Addresses
	r.HandleFunc("/api/account/addresses", h.ListAddresses).Methods("GET")
	r.HandleFunc("/api/account/addresses", h.CreateAddress).Methods("POST")
	r.HandleFunc("/api/account/addresses/{id}", h.UpdateAddress).Methods("PATCH")
	r.HandleFunc("/api/account/addresses/{id}", h.DeleteAddress).Methods("DELETE")

	// Wishlist
	r.HandleFunc("/api/account/wishlist", h.GetWishlist).Methods("GET")
	r.HandleFunc("/api/account/wishlist/{productId}", h.AddToWishlist).Methods("POST")
	r.HandleFunc("/api/account/wishlist/{productId}", h.RemoveFromWishlist).Methods("DELETE")
}

func (h *AccountHandler) getUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.Unauthorized(w, "Not authenticated")
		return 0, false
	}
	return userID, true
}

func (h *AccountHandler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	addresses, err := h.repo.GetAddresses(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "Failed to fetch addresses")
		return
	}
	response.Success(w, addresses)
}

type AddressRequest struct {
	Label      string `json:"label"`
	Recipient  string `json:"recipient"`
	Street     string `json:"street"`
	Number     string `json:"number"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        string `json:"zip"`
	IsDefault  bool   `json:"is_default"`
}

func (h *AccountHandler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	var req AddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.Label == "" || req.Recipient == "" || req.Street == "" || req.Number == "" || req.District == "" || req.City == "" || req.State == "" || req.Zip == "" {
		response.BadRequest(w, "All address fields are required")
		return
	}

	addr := models.Address{
		Label:      req.Label,
		Recipient:  req.Recipient,
		Street:     req.Street,
		Number:     req.Number,
		Complement: &req.Complement,
		District:   req.District,
		City:       req.City,
		State:      req.State,
		Zip:        req.Zip,
		IsDefault:  req.IsDefault,
	}

	created, err := h.repo.CreateAddress(r.Context(), userID, addr)
	if err != nil {
		response.InternalError(w, "Failed to create address")
		return
	}

	response.Created(w, created)
}

func (h *AccountHandler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid address ID")
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	updated, err := h.repo.UpdateAddress(r.Context(), userID, id, req)
	if err != nil {
		response.NotFound(w, "Address not found")
		return
	}

	response.Success(w, updated)
}

func (h *AccountHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid address ID")
		return
	}

	if err := h.repo.DeleteAddress(r.Context(), userID, id); err != nil {
		response.NotFound(w, "Address not found")
		return
	}

	response.NoContent(w)
}

func (h *AccountHandler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	ids, err := h.repo.GetWishlist(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "Failed to fetch wishlist")
		return
	}
	response.Success(w, ids)
}

func (h *AccountHandler) AddToWishlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	if err := h.repo.AddToWishlist(r.Context(), userID, productID); err != nil {
		response.InternalError(w, "Failed to add to wishlist")
		return
	}

	response.NoContent(w)
}

func (h *AccountHandler) RemoveFromWishlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid product ID")
		return
	}

	if err := h.repo.RemoveFromWishlist(r.Context(), userID, productID); err != nil {
		response.InternalError(w, "Failed to remove from wishlist")
		return
	}

	response.NoContent(w)
}