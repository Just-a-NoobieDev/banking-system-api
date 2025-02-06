package repositories

import (
	"banking-system/internal/database"
	"banking-system/internal/database/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(user models.CreateUserRequest) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	UpdateUser(user models.UpdateUserRequest, id int) (models.User, error)
	UpdateUserPassword(user models.UpdateUserPasswordRequest, id int) (models.User, error)
	GetUser(id int) (models.User, error)
	ViewBalance(id int) (models.ViewBalanceResponse, error)
}

type userRepository struct {
	db database.Service
}

func NewUserRepository(db database.Service) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user models.CreateUserRequest) (models.User, error) {
	var newUser models.User
	err := r.db.ExecTx(context.Background(), func(tx *sql.Tx) error {
		query := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES ($1, $2, $3, $4)
		RETURNING id, first_name, last_name, email, created_at, updated_at
		`

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		return tx.QueryRow(query, 
			user.FirstName, 
			user.LastName, 
			user.Email, 
			hashedPassword,
		).Scan(
			&newUser.ID,
			&newUser.FirstName,
			&newUser.LastName,
			&newUser.Email,
			&newUser.CreatedAt,
			&newUser.UpdatedAt,
		)
	}, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})

	if err != nil {
		return models.User{}, err
	}

	return newUser, nil
}

func (r *userRepository) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		query := `
		SELECT id, first_name, last_name, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
		`

		return tx.QueryRow(query, email).Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
	})

	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *userRepository) UpdateUser(user models.UpdateUserRequest, id int) (models.User, error) {
	var updatedUser models.User
	var accountsJSON string
	
	err := r.db.ExecTx(context.Background(), func(tx *sql.Tx) error {
		// Check if user exists
		existingUserQuery := `
		SELECT id, first_name, last_name, email
		FROM users
		WHERE id = $1`
		
		err := tx.QueryRow(existingUserQuery, id).Scan(
			&updatedUser.ID,
			&updatedUser.FirstName,
			&updatedUser.LastName,
			&updatedUser.Email,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("user not found")
			}
			return err
		}

		// Build dynamic update query based on provided fields
		query := "UPDATE users SET"
		params := []interface{}{}
		paramCount := 1

		if user.FirstName != "" {
			query += fmt.Sprintf(" first_name = $%d,", paramCount)
			params = append(params, user.FirstName)
			paramCount++
		}
		if user.LastName != "" {
			query += fmt.Sprintf(" last_name = $%d,", paramCount)
			params = append(params, user.LastName)
			paramCount++
		}
		if user.Email != "" {
			query += fmt.Sprintf(" email = $%d,", paramCount)
			params = append(params, user.Email)
			paramCount++
		}

		// Remove trailing comma and add WHERE clause
		query = query[:len(query)-1] + fmt.Sprintf(" WHERE id = $%d", paramCount)
		params = append(params, id)
		paramCount++

		// Add RETURNING clause
		query += " RETURNING id, first_name, last_name, email, created_at, updated_at"

		// If no fields to update, return the existing user
		if len(params) == 1 {
			return nil
		}

		err = tx.QueryRow(query, params...).Scan(
			&updatedUser.ID,
			&updatedUser.FirstName,
			&updatedUser.LastName,
			&updatedUser.Email,
			&updatedUser.CreatedAt,
			&updatedUser.UpdatedAt,
		)
		if err != nil {
			return err
		}

		// Get accounts
		accountsQuery := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'id', id,
					'balance', balance,
					'currency', currency
				)
				ORDER BY currency, id
			)::text,
			'[]'
		) as accounts
		FROM accounts
		WHERE user_id = $1`

		return tx.QueryRow(accountsQuery, id).Scan(&accountsJSON)
	}, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})

	if err != nil {
		return models.User{}, err
	}

	// Unmarshal accounts into user.Accounts
	err = json.Unmarshal([]byte(accountsJSON), &updatedUser.Accounts)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	return updatedUser, nil
}

func (r *userRepository) UpdateUserPassword(user models.UpdateUserPasswordRequest, id int) (models.User, error) {
	var updatedUser models.User
	var currentHashedPassword string
	var accountsJSON string

	err := r.db.ExecTx(context.Background(), func(tx *sql.Tx) error {
		// Get current password hash
		getCurrentPwQuery := `
		SELECT password FROM users WHERE id = $1
		`
		err := tx.QueryRow(getCurrentPwQuery, id).Scan(&currentHashedPassword)
		if err != nil {
			return fmt.Errorf("failed to get current password: %w", err)
		}

		// Verify old password
		err = bcrypt.CompareHashAndPassword([]byte(currentHashedPassword), []byte(user.OldPassword))
		if err != nil {
			return fmt.Errorf("invalid old password")
		}

		// Update with new password
		updateQuery := `
		UPDATE users
		SET password = $1
		WHERE id = $2
		RETURNING id, first_name, last_name, email, created_at, updated_at
		`

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash new password: %w", err)
		}

		err = tx.QueryRow(updateQuery, hashedPassword, id).Scan(
			&updatedUser.ID,
			&updatedUser.FirstName,
			&updatedUser.LastName,
			&updatedUser.Email,
			&updatedUser.CreatedAt,
			&updatedUser.UpdatedAt,
		)
		if err != nil {
			return err
		}

		// Get accounts
		accountsQuery := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'id', id,
					'balance', balance,
					'currency', currency
				)
				ORDER BY currency, id
			)::text,
			'[]'
		) as accounts
		FROM accounts
		WHERE user_id = $1`

		return tx.QueryRow(accountsQuery, id).Scan(&accountsJSON)
	}, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})

	if err != nil {
		return models.User{}, err
	}

	// Unmarshal accounts into user.Accounts
	err = json.Unmarshal([]byte(accountsJSON), &updatedUser.Accounts)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	return updatedUser, nil
}

func (r *userRepository) GetUser(id int) (models.User, error) {
	var user models.User
	var accountsJSON string

	err := r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		// First query to get user info
		userQuery := `
		SELECT id, first_name, last_name, email, created_at, updated_at
		FROM users
		WHERE id = $1`

		err := tx.QueryRow(userQuery, id).Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return err
		}

		// Second query to get all accounts
		accountsQuery := `
		SELECT COALESCE(
			json_agg(
				json_build_object(
					'id', id,
					'balance', balance,
					'currency', currency
				)
				ORDER BY currency, id
			)::text,
			'[]'
		) as accounts
		FROM accounts
		WHERE user_id = $1`

		return tx.QueryRow(accountsQuery, id).Scan(&accountsJSON)
	})

	if err != nil {
		return models.User{}, err
	}

	// Unmarshal accounts into user.Accounts
	err = json.Unmarshal([]byte(accountsJSON), &user.Accounts)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	return user, nil
}

func (r *userRepository) ViewBalance(id int) (models.ViewBalanceResponse, error) {
	var userBalance models.ViewBalanceResponse
	var accountsJSON string
	
	err := r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		query := `
		SELECT 
			COALESCE(
				json_agg(
					json_build_object(
						'id', a.id,
						'balance', a.balance,
						'currency', a.currency
					)
					ORDER BY a.currency, a.id
				)::text,
				'[]'
			) as accounts
		FROM accounts a
		WHERE a.user_id = $1`

		return tx.QueryRow(query, id).Scan(&accountsJSON)
	})

	if err != nil {
		if err == sql.ErrNoRows {
			return models.ViewBalanceResponse{
				Accounts: []models.AccountBalance{},
			}, nil
		}
		return models.ViewBalanceResponse{}, err
	}

	err = json.Unmarshal([]byte(accountsJSON), &userBalance.Accounts)
	if err != nil {
		return models.ViewBalanceResponse{}, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	// Calculate total balance per currency
	balancesByCurrency := make(map[string]float64)
	for _, account := range userBalance.Accounts {
		balancesByCurrency[account.Currency] += account.Balance
	}
	userBalance.BalancesByCurrency = balancesByCurrency

	return userBalance, nil
}