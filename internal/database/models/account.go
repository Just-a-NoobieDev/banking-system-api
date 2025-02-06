package models

import "time"

type Account struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  Currency  `json:"currency"`
	AccountName string `json:"account_name"`
	AccountDescription string `json:"account_description"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AccountDTO struct {
	ID        int       `json:"id"`
	Balance   float64   `json:"balance"`
	Currency  Currency  `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}	

type CreateAccountRequest struct {
	Currency Currency `json:"currency"`
	AccountName string `json:"account_name"`
	AccountDescription string `json:"account_description"`
}

type GetAccountRequest struct {
	ID int `json:"id"`
}

type DeleteAccountRequest struct {
	ID int `json:"id"`
}



