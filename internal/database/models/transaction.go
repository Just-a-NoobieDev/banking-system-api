package models

import "time"

type Transaction struct {
	ID          int             `json:"id"`
	AccountID   int             `json:"account_id"`
	Amount      float64         `json:"amount"`
	Type        TransactionType `json:"type"`
	Status      TransactionStatus `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	ReferenceID string          `json:"reference_id"`
}


type TransactionDTO struct {
	ID          int             `json:"id"`
	Amount      float64         `json:"amount"`
	Type        TransactionType `json:"type"`
	Status      TransactionStatus `json:"status"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CreateTransactionRequest struct {
	Amount      float64         `json:"amount"`
	AccountID   int             `json:"account_id"`
}

type CreateTransferRequest struct {
	Amount      float64         `json:"amount"`
	Description string          `json:"description"`
	AccountID   int             `json:"account_id"`
	DestinationAccountID int             `json:"destination_account_id"`
}
