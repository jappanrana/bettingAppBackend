package services

import (
	"context"
	"fmt"
	"time"

	"betting-app-backend-go/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WalletService struct {
	db *mongo.Database
}

func NewWalletService(db *mongo.Database) *WalletService {
	return &WalletService{db: db}
}

// GetOrCreateWallet retrieves user's wallet or creates new one
func (s *WalletService) GetOrCreateWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	collection := s.db.Collection("wallets")
	
	var wallet models.Wallet
	err := collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&wallet)
	
	if err == mongo.ErrNoDocuments {
		// Create new wallet
		wallet = models.Wallet{
			UserID:        userID,
			Balance:       0,
			Currency:      "INR",
			LockedBalance: 0,
			LastUpdated:   time.Now(),
			CreatedAt:     time.Now(),
		}
		
		_, err = collection.InsertOne(ctx, wallet)
		if err != nil {
			return nil, fmt.Errorf("failed to create wallet: %w", err)
		}
		return &wallet, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	
	return &wallet, nil
}

// GetBalance returns current wallet balance
func (s *WalletService) GetBalance(ctx context.Context, userID string) (*models.Wallet, error) {
	return s.GetOrCreateWallet(ctx, userID)
}

// CreatePaymentRequest creates a new deposit request
func (s *WalletService) CreatePaymentRequest(ctx context.Context, req *models.PaymentRequest) error {
	collection := s.db.Collection("payment_requests")
	
	req.ID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	req.Status = "pending"
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	
	_, err := collection.InsertOne(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create payment request: %w", err)
	}
	
	return nil
}

// GetPaymentRequests retrieves payment requests (filtered by user or all for admin)
func (s *WalletService) GetPaymentRequests(ctx context.Context, userID string, isAdmin bool, status string) ([]models.PaymentRequest, error) {
	collection := s.db.Collection("payment_requests")
	
	filter := bson.M{}
	if !isAdmin {
		filter["user_id"] = userID
	}
	if status != "" && status != "all" {
		filter["status"] = status
	}
	
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment requests: %w", err)
	}
	defer cursor.Close(ctx)
	
	var requests []models.PaymentRequest
	if err = cursor.All(ctx, &requests); err != nil {
		return nil, fmt.Errorf("failed to decode payment requests: %w", err)
	}
	
	return requests, nil
}

// ProcessPaymentRequest approves or declines a payment request
func (s *WalletService) ProcessPaymentRequest(ctx context.Context, requestID string, status string, adminNotes string) error {
	if status != "accepted" && status != "declined" {
		return fmt.Errorf("invalid status: %s", status)
	}
	
	collection := s.db.Collection("payment_requests")
	
	// Get the payment request
	var req models.PaymentRequest
	err := collection.FindOne(ctx, bson.M{"_id": requestID}).Decode(&req)
	if err != nil {
		return fmt.Errorf("payment request not found: %w", err)
	}
	
	if req.Status != "pending" {
		return fmt.Errorf("payment request already processed")
	}
	
	// Start MongoDB transaction for atomic operation
	session, err := s.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)
	
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Update payment request status
		update := bson.M{
			"$set": bson.M{
				"status":      status,
				"admin_notes": adminNotes,
				"updated_at":  time.Now(),
			},
		}
		
		_, err := collection.UpdateOne(sessCtx, bson.M{"_id": requestID}, update)
		if err != nil {
			return nil, fmt.Errorf("failed to update payment request: %w", err)
		}
		
		// If accepted, credit the wallet
		if status == "accepted" {
			err = s.creditBalanceInTransaction(sessCtx, req.UserID, req.Amount, fmt.Sprintf("Payment request %s accepted", requestID), "deposit")
			if err != nil {
				return nil, fmt.Errorf("failed to credit wallet: %w", err)
			}
		}
		
		return nil, nil
	})
	
	return err
}

// DeductBalance deducts amount from wallet (for game bets)
func (s *WalletService) DeductBalance(ctx context.Context, userID string, amount float64, description string, category string) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	
	walletsCol := s.db.Collection("wallets")
	transactionsCol := s.db.Collection("transactions")
	
	// Start transaction
	session, err := s.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)
	
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Get current wallet
		var wallet models.Wallet
		err := walletsCol.FindOne(sessCtx, bson.M{"user_id": userID}).Decode(&wallet)
		if err != nil {
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		
		if wallet.Balance < amount {
			return nil, fmt.Errorf("insufficient balance")
		}
		
		balanceBefore := wallet.Balance
		balanceAfter := balanceBefore - amount
		
		// Update wallet
		_, err = walletsCol.UpdateOne(
			sessCtx,
			bson.M{"user_id": userID},
			bson.M{
				"$set": bson.M{
					"balance":      balanceAfter,
					"last_updated": time.Now(),
				},
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update wallet: %w", err)
		}
		
		// Create transaction record
		transaction := models.Transaction{
			ID:            fmt.Sprintf("txn_%d", time.Now().UnixNano()),
			UserID:        userID,
			Type:          "debit",
			Amount:        amount,
			Description:   description,
			Category:      category,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			Status:        "completed",
			CreatedAt:     time.Now(),
		}
		
		_, err = transactionsCol.InsertOne(sessCtx, transaction)
		if err != nil {
			return nil, fmt.Errorf("failed to create transaction: %w", err)
		}
		
		return nil, nil
	})
	
	return err
}

// CreditBalance adds amount to wallet (for game wins)
func (s *WalletService) CreditBalance(ctx context.Context, userID string, amount float64, description string, category string) error {
	return s.creditBalanceInTransaction(ctx, userID, amount, description, category)
}

// creditBalanceInTransaction is internal method that can be used within transactions
func (s *WalletService) creditBalanceInTransaction(ctx context.Context, userID string, amount float64, description string, category string) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	
	walletsCol := s.db.Collection("wallets")
	transactionsCol := s.db.Collection("transactions")
	
	// Get or create wallet
	wallet, err := s.GetOrCreateWallet(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}
	
	balanceBefore := wallet.Balance
	balanceAfter := balanceBefore + amount
	
	// Update wallet
	_, err = walletsCol.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{
			"$set": bson.M{
				"balance":      balanceAfter,
				"last_updated": time.Now(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}
	
	// Create transaction record
	transaction := models.Transaction{
		ID:            fmt.Sprintf("txn_%d", time.Now().UnixNano()),
		UserID:        userID,
		Type:          "credit",
		Amount:        amount,
		Description:   description,
		Category:      category,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Status:        "completed",
		CreatedAt:     time.Now(),
	}
	
	_, err = transactionsCol.InsertOne(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	
	return nil
}

// GetTransactions retrieves transaction history for a user
func (s *WalletService) GetTransactions(ctx context.Context, userID string, limit int64) ([]models.Transaction, error) {
	collection := s.db.Collection("transactions")
	
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)
	
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer cursor.Close(ctx)
	
	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %w", err)
	}
	
	return transactions, nil
}

// GetPaymentDetails retrieves admin payment details for deposits
func (s *WalletService) GetPaymentDetails(ctx context.Context) (*models.PaymentDetails, error) {
	collection := s.db.Collection("payment_details")
	
	var details models.PaymentDetails
	err := collection.FindOne(ctx, bson.M{}).Decode(&details)
	if err == mongo.ErrNoDocuments {
		// Return default details
		return &models.PaymentDetails{
			BankName:          "HDFC Bank",
			AccountNumber:     "1234567890",
			IFSCCode:          "HDFC0001234",
			AccountHolderName: "NeonPlay Gaming Pvt Ltd",
			UPIID:             "neonplay@hdfc",
			QRCodeURL:         "/payment-qr.png",
			UpdatedAt:         time.Now(),
		}, nil
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get payment details: %w", err)
	}
	
	return &details, nil
}

// UpdatePaymentDetails updates admin payment details
func (s *WalletService) UpdatePaymentDetails(ctx context.Context, details *models.PaymentDetails) error {
	collection := s.db.Collection("payment_details")
	
	details.UpdatedAt = time.Now()
	
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{},
		bson.M{"$set": details},
		opts,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update payment details: %w", err)
	}
	
	return nil
}
