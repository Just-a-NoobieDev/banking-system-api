package models

import "time"

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Accounts []AccountMinimal `json:"accounts"`
}

type AccountMinimal struct {
	ID       int     `json:"id"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

type UserDTO struct {
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type UpdateUserPasswordRequest struct {
	NewPassword string `json:"new_password"`
	OldPassword string `json:"old_password"`
}

type ViewBalanceResponse struct {
	Accounts           []AccountBalance    `json:"accounts"`
	BalancesByCurrency map[string]float64 `json:"balances_by_currency"`
}

type AccountBalance struct {
	ID       int     `json:"id"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}

