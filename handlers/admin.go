package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"betting-app-backend-go/models"
	"betting-app-backend-go/services"
)

type AdminHandler struct {
	walletService *services.WalletService
}

func NewAdminHandler(walletService *services.WalletService) *AdminHandler {
	return &AdminHandler{walletService: walletService}
}

// GetAllPaymentRequests handles GET /api/admin/payment-requests
func (h *AdminHandler) GetAllPaymentRequests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	
	log.Printf("[Admin] Getting all payment requests (status: %s)\n", status)
	
	// Pass empty userID and isAdmin=true to get all requests
	requests, err := h.walletService.GetPaymentRequests(context.Background(), "", true, status)
	if err != nil {
		log.Printf("[Admin] ❌ Failed to get payment requests: %v\n", err)
		http.Error(w, "failed to get payment requests", http.StatusInternalServerError)
		return
	}
	
	log.Printf("[Admin] ✅ Retrieved %d payment requests\n", len(requests))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// ProcessPaymentRequest handles POST /api/admin/payment-request/:id/:action
func (h *AdminHandler) ProcessPaymentRequest(w http.ResponseWriter, r *http.Request) {
	// Extract request ID and action from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/payment-request/")
	parts := strings.Split(path, "/")
	
	if len(parts) != 2 {
		http.Error(w, "invalid URL format", http.StatusBadRequest)
		return
	}
	
	requestID := parts[0]
	action := parts[1]
	
	if action != "approve" && action != "decline" {
		http.Error(w, "invalid action: must be approve or decline", http.StatusBadRequest)
		return
	}
	
	status := "accepted"
	if action == "decline" {
		status = "declined"
	}
	
	var body struct {
		AdminNotes string `json:"adminNotes"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		// Admin notes are optional
		body.AdminNotes = ""
	}
	
	log.Printf("[Admin] Processing payment request %s: %s\n", requestID, action)
	
	err := h.walletService.ProcessPaymentRequest(context.Background(), requestID, status, body.AdminNotes)
	if err != nil {
		log.Printf("[Admin] ❌ Failed to process payment request: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	log.Printf("[Admin] ✅ Payment request %s %s\n", requestID, status)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Payment request %s", status),
	})
}

// UpdatePaymentDetails handles PUT /api/admin/payment-details
func (h *AdminHandler) UpdatePaymentDetails(w http.ResponseWriter, r *http.Request) {
	var details models.PaymentDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		log.Printf("[Admin] ❌ Invalid request body: %v\n", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	log.Println("[Admin] Updating payment details")
	
	err := h.walletService.UpdatePaymentDetails(context.Background(), &details)
	if err != nil {
		log.Printf("[Admin] ❌ Failed to update payment details: %v\n", err)
		http.Error(w, "failed to update payment details", http.StatusInternalServerError)
		return
	}
	
	log.Println("[Admin] ✅ Payment details updated")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}
