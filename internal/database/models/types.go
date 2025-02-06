package models

import (
	"time"
)

// Currency related
type Currency string

const (
    USD Currency = "USD"
    EUR Currency = "EUR"
    GBP Currency = "GBP"
)

// Transaction related
type TransactionType string

const (
    Deposit    TransactionType = "DEPOSIT"
    Withdrawal TransactionType = "WITHDRAWAL"
)

// Common response structure
type Response struct {
    StatusCode int    `json:"status_code"`
    Success    bool   `json:"success"`
    Message    string `json:"message"`
    Data       any    `json:"data"`
}

type TransactionStatus string

const (
	Pending TransactionStatus = "pending"
	Completed TransactionStatus = "completed"
	Failed TransactionStatus = "failed"
)

// Common pagination, filtering and sorting types
type PaginationRequest struct {
    Page     int `json:"page"`
    PageSize int `json:"page_size"`
}

type PaginationResponse struct {
    CurrentPage  int   `json:"current_page"`
    PageSize     int   `json:"page_size"`
    TotalPages   int   `json:"total_pages"`
    TotalRecords int64 `json:"total_records"`
}

type SortRequest struct {
    Field     string `json:"field"`
    Direction string `json:"direction"` // "ASC" or "DESC"
}

// Filter types for different entities
type AccountFilter struct {
    MinBalance *float64        `json:"min_balance,omitempty"`
    MaxBalance *float64        `json:"max_balance,omitempty"`
    Currency   *Currency       `json:"currency,omitempty"`
    DateFrom   *time.Time      `json:"date_from,omitempty"`
    DateTo     *time.Time      `json:"date_to,omitempty"`
}

type TransactionFilter struct {
    Type          *TransactionType   `json:"type,omitempty"`
    Status        *TransactionStatus `json:"status,omitempty"`
    MinAmount     *float64          `json:"min_amount,omitempty"`
    MaxAmount     *float64          `json:"max_amount,omitempty"`
    DateFrom      *time.Time        `json:"date_from,omitempty"`
    DateTo        *time.Time        `json:"date_to,omitempty"`
    AccountID     *int              `json:"account_id,omitempty"`
}

type UserFilter struct {
    Email         *string    `json:"email,omitempty"`
    Status        *string    `json:"status,omitempty"`
    RegisteredFrom *time.Time `json:"registered_from,omitempty"`
    RegisteredTo   *time.Time `json:"registered_to,omitempty"`
}

// Common paginated response wrapper
type PaginatedResponse[T any] struct {
    Data       []T               `json:"data"`
    Pagination PaginationResponse `json:"pagination"`
}

