package models

import (
	"time"
)

type Wallet struct {
	UserID        string    `bson:"user_id" json:"userId"`
	Balance       float64   `bson:"balance" json:"balance"`
	Currency      string    `bson:"currency" json:"currency"`
	LockedBalance float64   `bson:"locked_balance" json:"lockedBalance"`
	LastUpdated   time.Time `bson:"last_updated" json:"lastUpdated"`
	CreatedAt     time.Time `bson:"created_at" json:"createdAt"`
}

type Transaction struct {
	ID            string    `bson:"_id" json:"id"`
	UserID        string    `bson:"user_id" json:"userId"`
	Type          string    `bson:"type" json:"type"` // credit, debit
	Amount        float64   `bson:"amount" json:"amount"`
	Description   string    `bson:"description" json:"description"`
	Category      string    `bson:"category" json:"category"` // deposit, withdrawal, game_win, game_loss, admin_adjustment
	BalanceBefore float64   `bson:"balance_before" json:"balanceBefore"`
	BalanceAfter  float64   `bson:"balance_after" json:"balanceAfter"`
	Status        string    `bson:"status" json:"status"` // pending, completed, failed
	CreatedAt     time.Time `bson:"created_at" json:"createdAt"`
}

type PaymentRequest struct {
	ID            string    `bson:"_id" json:"id"`
	UserID        string    `bson:"user_id" json:"userId"`
	Amount        float64   `bson:"amount" json:"amount"`
	PaymentMethod string    `bson:"payment_method" json:"paymentMethod"` // bank, upi
	TransactionID string    `bson:"transaction_id,omitempty" json:"transactionId,omitempty"`
	ProofURL      string    `bson:"proof_url,omitempty" json:"proofUrl,omitempty"`
	Status        string    `bson:"status" json:"status"` // pending, accepted, declined
	Notes         string    `bson:"notes,omitempty" json:"notes,omitempty"`
	AdminNotes    string    `bson:"admin_notes,omitempty" json:"adminNotes,omitempty"`
	CreatedAt     time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updatedAt"`
}

type PaymentDetails struct {
	BankName          string `bson:"bank_name" json:"bankName"`
	AccountNumber     string `bson:"account_number" json:"accountNumber"`
	IFSCCode          string `bson:"ifsc_code" json:"ifscCode"`
	AccountHolderName string `bson:"account_holder_name" json:"accountHolderName"`
	UPIID             string `bson:"upi_id" json:"upiId"`
	QRCodeURL         string `bson:"qr_code_url" json:"qrCodeUrl"`
	UpdatedAt         time.Time `bson:"updated_at" json:"updatedAt"`
}
