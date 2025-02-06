package models

import "time"

type SOA struct {
	ID          int       `json:"id"`
	AccountID   int       `json:"account_id"`
	StatementDate time.Time `json:"statement_date"`
	PDFUrl      string    `json:"pdf_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      int       `json:"user_id"`
}

type GenerateSOACustomRequestUnformatted struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	AccountID int       `json:"account_id"`
	ItemCount int `json:"item_count" default:"100"`
	Currency string `json:"currency" default:"USD"`
}

type GenerateSOACustomRequest struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	AccountID int       `json:"account_id"`
	ItemCount int `json:"item_count" default:"100"`
	Currency string `json:"currency" default:"USD"`
}






