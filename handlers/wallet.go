package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"betting-app-backend-go/middleware"
	"betting-app-backend-go/models"
	"betting-app-backend-go/services"
)

type WalletHandler struct {
	service *services.WalletService
}

func NewWalletHandler(service *services.WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

// GetBalance handles GET /api/wallet/balance
func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	log.Printf("[Wallet] Getting balance for user: %s\n", userID)
	
	wallet, err := h.service.GetBalance(context.Background(), userID)
	if err != nil {
		log.Printf("[Wallet] ❌ Failed to get balance: %v\n", err)
		http.Error(w, "failed to get balance", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[Wallet] ✅ Balance retrieved: %.2f %s\n", wallet.Balance, wallet.Currency)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

// CreatePaymentRequest handles POST /api/wallet/payment-request
func (h *WalletHandler) CreatePaymentRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Wallet] ❌ Invalid request body: %v\n", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	req.UserID = userID
	
	log.Printf("[Wallet] Creating payment request: %.2f INR via %s\n", req.Amount, req.PaymentMethod)
	
	if req.Amount <= 0 {
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}
	
	if req.PaymentMethod != "bank" && req.PaymentMethod != "upi" {
		http.Error(w, "invalid payment method", http.StatusBadRequest)
		return
	}
	
	err := h.service.CreatePaymentRequest(context.Background(), &req)
	if err != nil {
		log.Printf("[Wallet] ❌ Failed to create payment request: %v\n", err)
		http.Error(w, "failed to create payment request", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[Wallet] ✅ Payment request created: %s\n", req.ID)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(req)
}

// GetPaymentRequests handles GET /api/wallet/payment-requests
func (h *WalletHandler) GetPaymentRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	role, _ := middleware.GetUserRole(r)
	isAdmin := role == "admin"
	
	status := r.URL.Query().Get("status")
	
	log.Printf("[Wallet] Getting payment requests (user: %s, admin: %v, status: %s)\n", userID, isAdmin, status)
	
	requests, err := h.service.GetPaymentRequests(context.Background(), userID, isAdmin, status)
	if err != nil {
		log.Printf("[Wallet] ❌ Failed to get payment requests: %v\n", err)
		http.Error(w, "failed to get payment requests", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[Wallet] ✅ Retrieved %d payment requests\n", len(requests))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// GetTransactions handles GET /api/wallet/transactions
func (h *WalletHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	
	limitStr := r.URL.Query().Get("limit")
	limit := int64(50) // default
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}
	
	log.Printf("[Wallet] Getting transactions for user: %s (limit: %d)\n", userID, limit)
	
	transactions, err := h.service.GetTransactions(context.Background(), userID, limit)
	if err != nil {
		log.Printf("[Wallet] ❌ Failed to get transactions: %v\n", err)
		http.Error(w, "failed to get transactions", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[Wallet] ✅ Retrieved %d transactions\n", len(transactions))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// GetPaymentDetails handles GET /api/wallet/payment-details
func (h *WalletHandler) GetPaymentDetails(w http.ResponseWriter, r *http.Request) {
	log.Println("[Wallet] Getting payment details")
	
	details, err := h.service.GetPaymentDetails(context.Background())
	if err != nil {
		log.Printf("[Wallet] ❌ Failed to get payment details: %v\n", err)
		http.Error(w, "failed to get payment details", http.StatusInternalServerError)
		return
	}
	
	log.Println("[Wallet] ✅ Payment details retrieved")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}
