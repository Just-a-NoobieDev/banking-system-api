package repositories

import (
	"banking-system/internal/database"
	"banking-system/internal/database/models"
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TransactionRepository interface {
	Deposit(transaction models.CreateTransactionRequest, userID int) (map[string]interface{}, error)
	Withdraw(transaction models.CreateTransactionRequest, userID int) (map[string]interface{}, error)
	GetTransactions(
		userID int,
		filter *models.TransactionFilter,
		sort *models.SortRequest,
		pagination *models.PaginationRequest,
	) (*models.PaginatedResponse[models.Transaction], error)
	GetTransaction(transactionID int) (models.Transaction, error)
	GetTransactionsForSOA(userID int, request models.GenerateSOACustomRequest) ([]models.Transaction, error)
}

type transactionRepository struct {
	db database.Service
}

func NewTransactionRepository(db database.Service) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Deposit(transaction models.CreateTransactionRequest, userID int) (map[string]interface{}, error) {
	// Round amount to 2 decimal places
	transaction.Amount = math.Round(transaction.Amount*100) / 100

	var transactionID int
	var generatedReferenceID string
	err := r.db.ExecTx(context.Background(), func(tx *sql.Tx) error {
		generatedReferenceID = uuid.New().String()


		query := `
		INSERT INTO transactions (account_id, amount, transaction_type, status, created_at, updated_at, reference_id, user_id)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $5, $6)
		RETURNING id, account_id, amount, transaction_type, status, created_at, updated_at, reference_id
		`

		var (
			accountID       int
			amount         float64
			transactionType string
			status         models.TransactionStatus
			createdAt      time.Time
			updatedAt      time.Time
			referenceID    string
		)

		err := tx.QueryRow(query, 
			transaction.AccountID, 
			transaction.Amount, 
			models.Deposit,
			models.Completed,
			generatedReferenceID,
			userID,
			).Scan(
				&transactionID,
				&accountID,
				&amount,
			&transactionType,
			&status,
			&createdAt,
			&updatedAt,
			&referenceID,
		)

		if err != nil {
			return fmt.Errorf("failed to create deposit transaction: %w", err)
		}

		// Update account balance
		updateQuery := `
		UPDATE accounts 
		SET balance = balance + $1, 
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		`

		result, err := tx.Exec(updateQuery, transaction.Amount, transaction.AccountID)
		if err != nil {
			return fmt.Errorf("failed to update account balance: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("account not found")
		}

		return nil
	}, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})

	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"transaction_id": transactionID,
		"reference_id": generatedReferenceID,
	}, nil
}

func (r *transactionRepository) Withdraw(transaction models.CreateTransactionRequest, userID int) (map[string]interface{}, error) {
	// Round amount to 2 decimal places
	transaction.Amount = math.Round(transaction.Amount*100) / 100

	var transactionID int
	var generatedReferenceID string
	err := r.db.ExecTx(context.Background(), func(tx *sql.Tx) error {
		// First check if account has sufficient balance
		var currentBalance float64
		balanceQuery := `SELECT balance FROM accounts WHERE id = $1 FOR UPDATE`
		err := tx.QueryRow(balanceQuery, transaction.AccountID).Scan(&currentBalance)
		if err != nil {
			return fmt.Errorf("failed to get account balance: %w", err)
		}

		// Round current balance to 2 decimal places for comparison
		currentBalance = math.Round(currentBalance*100) / 100

		if currentBalance < transaction.Amount {
			return fmt.Errorf("insufficient funds")
		}

		generatedReferenceID = uuid.New().String()

		query := `
		INSERT INTO transactions (account_id, amount, transaction_type, status, created_at, updated_at, reference_id, user_id)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $5, $6)
		RETURNING id, account_id, amount, transaction_type, status, created_at, updated_at, reference_id
		`

		var (
			accountID       int
			amount         float64
			transactionType string
			status         models.TransactionStatus
			createdAt      time.Time
			updatedAt      time.Time
			referenceID    string
		)

		err = tx.QueryRow(query, 
			transaction.AccountID, 
			transaction.Amount, 
			models.Withdrawal,
			models.Completed,
			generatedReferenceID,
			userID,
		).Scan(
			&transactionID,
			&accountID,
			&amount,
			&transactionType,
			&status,
			&createdAt,
			&updatedAt,
			&referenceID,
		)

		if err != nil {
			return fmt.Errorf("failed to create withdrawal transaction: %w", err)
		}

		// Update account balance
		updateQuery := `
		UPDATE accounts 
		SET balance = balance - $1, 
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		`

		result, err := tx.Exec(updateQuery, transaction.Amount, transaction.AccountID)
		if err != nil {
			return fmt.Errorf("failed to update account balance: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("account not found")
		}

		return nil
	}, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})

	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"transaction_id": transactionID,
		"reference_id": generatedReferenceID,
	}, nil
}

func buildTransactionFilterClause(filter *models.TransactionFilter, startParam int) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	var conditions []string
	var args []interface{}
	paramCount := startParam

	if filter.MinAmount != nil {
		paramCount++
		conditions = append(conditions, fmt.Sprintf("amount >= $%d", paramCount))
		args = append(args, *filter.MinAmount)
	}

	if filter.MaxAmount != nil {
		paramCount++
		conditions = append(conditions, fmt.Sprintf("amount <= $%d", paramCount))
		args = append(args, *filter.MaxAmount)
	}

	if filter.Type != nil {
		paramCount++
		conditions = append(conditions, fmt.Sprintf("transaction_type = $%d", paramCount))
		args = append(args, *filter.Type)
	}

	if filter.Status != nil {
		paramCount++
		conditions = append(conditions, fmt.Sprintf("status = $%d", paramCount))
		args = append(args, *filter.Status)
	}

	if filter.DateFrom != nil {
		paramCount++
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramCount))
		args = append(args, *filter.DateFrom)
	}

	if filter.DateTo != nil {
		paramCount++
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramCount))
		args = append(args, *filter.DateTo)
	}

	if len(conditions) == 0 {
		return "", nil, nil
	}

	return " AND " + strings.Join(conditions, " AND "), args, nil
}

func buildTransactionSortClause(sort *models.SortRequest) string {
	if sort == nil || sort.Field == "" {
		return " ORDER BY created_at DESC"
	}

	// Define allowed fields for sorting
	allowedFields := map[string]bool{
		"amount":     true,
		"transaction_type":       true,
		"status":     true,
		"created_at": true,
	}

	if !allowedFields[sort.Field] {
		return " ORDER BY created_at DESC"
	}

	direction := "ASC"
	if strings.ToUpper(sort.Direction) == "DESC" {
		direction = "DESC"
	}

	return fmt.Sprintf(" ORDER BY %s %s", sort.Field, direction)
}

func (r *transactionRepository) GetTransactions(
	userID int,
	filter *models.TransactionFilter,
	sort *models.SortRequest,
	pagination *models.PaginationRequest,
) (*models.PaginatedResponse[models.Transaction], error) {
	var transactions []models.Transaction
	var totalRecords int64
	var response *models.PaginatedResponse[models.Transaction]

	err := r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		// Build base query
		baseQuery := `FROM transactions WHERE user_id = $1`
		args := []interface{}{userID}
		paramCount := 1

		// Apply filters
		whereClause, filterArgs, err := buildTransactionFilterClause(filter, paramCount)
		if err != nil {
			return err
		}
		baseQuery += whereClause
		args = append(args, filterArgs...)
		paramCount += len(filterArgs)

		// Get total count for pagination
		countQuery := fmt.Sprintf("SELECT COUNT(*) %s", baseQuery)
		err = tx.QueryRow(countQuery, args...).Scan(&totalRecords)
		if err != nil {
			return fmt.Errorf("failed to get total count: %w", err)
		}

		// Apply sorting
		baseQuery += buildTransactionSortClause(sort)

		// Apply pagination
		if pagination != nil {
			offset := (pagination.Page - 1) * pagination.PageSize
			baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount+1, paramCount+2)
			args = append(args, pagination.PageSize, offset)
		}

		// Execute final query
		query := "SELECT id, account_id, amount, transaction_type, status, created_at, updated_at, reference_id " + baseQuery
		rows, err := tx.Query(query, args...)
		if err != nil {
			return fmt.Errorf("failed to query transactions: %w", err)
		}
		defer rows.Close()

		// Scan results
		transactions = make([]models.Transaction, 0)
		for rows.Next() {
			var t models.Transaction
			if err := rows.Scan(
				&t.ID,
				&t.AccountID,
				&t.Amount,
				&t.Type,
				&t.Status,
				&t.CreatedAt,
				&t.UpdatedAt,
				&t.ReferenceID,
			); err != nil {
				return fmt.Errorf("failed to scan transaction: %w", err)
			}
			transactions = append(transactions, t)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating transactions: %w", err)
		}

		// Calculate pagination metadata
		totalPages := int(math.Ceil(float64(totalRecords) / float64(pagination.PageSize)))

		response = &models.PaginatedResponse[models.Transaction]{
			Data: transactions,
			Pagination: models.PaginationResponse{
				CurrentPage:  pagination.Page,
				PageSize:     pagination.PageSize,
				TotalPages:   totalPages,
				TotalRecords: totalRecords,
			},
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	return response, nil
}

func (r *transactionRepository) GetTransaction(transactionID int) (models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		query := `
		SELECT id, account_id, amount, transaction_type, status, created_at, updated_at
		FROM transactions
		WHERE id = $1
		`

		err := tx.QueryRow(query, transactionID).Scan(
			&transaction.ID,
			&transaction.AccountID,
			&transaction.Amount,
			&transaction.Type,
			&transaction.Status,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)


		if err != nil {
			return fmt.Errorf("failed to get transaction: %w", err)
		}

		return nil
	})

	if err != nil {
		return models.Transaction{}, err
	}

	if transaction.ID == 0 {
		return models.Transaction{}, fmt.Errorf("transaction not found")
	}

	return transaction, nil
}

func (r *transactionRepository) GetTransactionsForSOA(userID int, request models.GenerateSOACustomRequest) ([]models.Transaction, error) {
	// Get all transactions for the account and filter by the request
	var transactions []models.Transaction

	err := r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		query := `
			SELECT t.id, t.account_id, t.amount, t.transaction_type, t.status, t.created_at, t.updated_at, t.reference_id
			FROM transactions t
			JOIN accounts a ON t.account_id = a.id
			WHERE t.account_id IN (SELECT id FROM accounts WHERE user_id = $1)
			AND t.created_at BETWEEN $2 AND $3
		`
		args := []interface{}{userID, request.StartDate, request.EndDate}
		paramCount := 3

		if request.AccountID != 0 {
			query += fmt.Sprintf(" AND t.account_id = $%d", paramCount+1)
			args = append(args, request.AccountID)
			paramCount++
		}

		if request.ItemCount != 0 {
			query += fmt.Sprintf(" LIMIT $%d", paramCount+1)
			args = append(args, request.ItemCount)
			paramCount++
		}

		if request.Currency != "" {
			query += fmt.Sprintf(" AND a.currency = $%d", paramCount+1)
			args = append(args, request.Currency)
			paramCount++
		}

		query += " ORDER BY t.created_at DESC"

		fmt.Println(query)
		fmt.Println(args)
		fmt.Println(paramCount)

		rows, err := tx.Query(query, args...)
		if err != nil {
			return fmt.Errorf("failed to query transactions: %w", err)
		}
		defer rows.Close()

		// Scan results
		transactions = make([]models.Transaction, 0)
		for rows.Next() {
			var t models.Transaction
			if err := rows.Scan(
				&t.ID,
				&t.AccountID,
				&t.Amount,
				&t.Type,
				&t.Status,
				&t.CreatedAt,
				&t.UpdatedAt,
				&t.ReferenceID,
			); err != nil {
				return fmt.Errorf("failed to scan transaction: %w", err)
			}
			transactions = append(transactions, t)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating transactions: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get transactions for SOA: %w", err)
	}

	return transactions, nil
}