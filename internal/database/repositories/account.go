package repositories

import (
	"banking-system/internal/database"
	"banking-system/internal/database/models"
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
)

type AccountRepository interface {
	CreateAccount(account models.CreateAccountRequest, userID int) (int, error)
	GetAccount(id int) (models.Account, error)
	GetAccounts(
		userID int, 
		filter *models.AccountFilter, 
		sort *models.SortRequest, 
		pagination *models.PaginationRequest,
	) (*models.PaginatedResponse[models.Account], error)
	DeleteAccount(id int) error
}


type accountRepository struct {
	db database.Service
}

func NewAccountRepository(db database.Service) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) CreateAccount(account models.CreateAccountRequest, userID int) (int, error) {
	var accountID int
	err := r.db.ExecTx(context.Background(), func(tx *sql.Tx) error {
		// First verify the user exists
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check user existence: %w", err)
		}
		if !exists {
			return fmt.Errorf("user with ID %d does not exist", userID)
		}

		query := `
		INSERT INTO accounts (user_id, balance, currency, account_name, account_description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
		`

		// Simplified RETURNING clause to just get the ID
		err = tx.QueryRow(
			query,
			userID,           // Make sure we use the parameter userID, not the local variable
			0.0,             // Initial balance as decimal
			account.Currency,
			account.AccountName,
			account.AccountDescription,
		).Scan(&accountID)

		if err != nil {
			return fmt.Errorf("failed to create account (userID=%d): %w", userID, err)
		}
		return nil
	}, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // Highest isolation level to ensure consistency
	})

	if err != nil {
		return 0, err
	}

	return accountID, nil
}

const accountColumns = `id, user_id, balance, currency, account_name, account_description, created_at, updated_at`

func (r *accountRepository) GetAccount(id int) (models.Account, error) {
	var account models.Account
	err := r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		query := `
		SELECT ` + accountColumns + `
		FROM accounts
		WHERE id = $1`

		return tx.QueryRow(query, id).Scan(
			&account.ID,
			&account.UserID,
			&account.Balance,
			&account.Currency,
			&account.AccountName,
			&account.AccountDescription,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
	})

	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (r *accountRepository) GetAccounts(
	userID int, 
	filter *models.AccountFilter, 
	sort *models.SortRequest, 
	pagination *models.PaginationRequest,
) (*models.PaginatedResponse[models.Account], error) {
	// Build the base query
	baseQuery := `FROM accounts WHERE user_id = $1`
	args := []interface{}{userID}
	paramCount := 1

	// Apply filters
	whereClause, filterArgs, err := buildAccountFilterClause(filter, paramCount)
	if err != nil {
		return nil, err
	}
	query := baseQuery + whereClause
	args = append(args, filterArgs...)
	paramCount += len(filterArgs)

	var totalRecords int64
	err = r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		// Get total count for pagination
		countQuery := "SELECT COUNT(*) " + query
		return tx.QueryRow(countQuery, args...).Scan(&totalRecords)
	})
	if err != nil {
		return nil, err
	}

	// Apply sorting
	query += buildSortClause(sort, []string{"balance", "currency", "created_at"})

	// Apply pagination
	offset := (pagination.Page - 1) * pagination.PageSize
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount+1, paramCount+2)
	args = append(args, pagination.PageSize, offset)

	// Execute final query
	rows, err := r.db.QueryContext(context.Background(), "SELECT "+accountColumns+" "+query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]models.Account, 0, pagination.PageSize)
	for rows.Next() {
		var account models.Account
		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Balance,
			&account.Currency,
			&account.AccountName,
			&account.AccountDescription,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pagination.PageSize)))
	
	return &models.PaginatedResponse[models.Account]{
		Data: accounts,
		Pagination: models.PaginationResponse{
			CurrentPage:  pagination.Page,
			PageSize:     pagination.PageSize,
			TotalPages:   totalPages,
			TotalRecords: totalRecords,
		},
	}, nil
}

// Helper functions for building query clauses
func buildAccountFilterClause(filter *models.AccountFilter, startParam int) (string, []interface{}, error) {
	if filter == nil {
		return "", nil, nil
	}

	var conditions []string
	var args []interface{}
	paramCount := startParam

	if filter.MinBalance != nil {
		paramCount++
		conditions = append(conditions, fmt.Sprintf("balance >= $%d", paramCount))
		args = append(args, *filter.MinBalance)
	}

	if filter.MaxBalance != nil {
		paramCount++
		conditions = append(conditions, fmt.Sprintf("balance <= $%d", paramCount))
		args = append(args, *filter.MaxBalance)
	}

	if filter.Currency != nil {
		switch *filter.Currency {
		case models.USD, models.EUR, models.GBP:
			paramCount++
			conditions = append(conditions, fmt.Sprintf("currency = $%d", paramCount))
			args = append(args, string(*filter.Currency))
		default:
			return "", nil, fmt.Errorf("invalid currency: %s", *filter.Currency)
		}
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

func buildSortClause(sort *models.SortRequest, allowedFields []string) string {
	if sort == nil || sort.Field == "" {
		return " ORDER BY created_at DESC"
	}

	// Check if the sort field is allowed
	isAllowed := false
	for _, field := range allowedFields {
		if sort.Field == field {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return " ORDER BY created_at DESC"
	}

	direction := "ASC"
	if strings.ToUpper(sort.Direction) == "DESC" {
		direction = "DESC"
	}

	return fmt.Sprintf(" ORDER BY %s %s", sort.Field, direction)
}

func (r *accountRepository) DeleteAccount(id int) error {
	return r.db.ExecTx(context.Background(), func(tx *sql.Tx) error {
		query := `
		DELETE FROM accounts
		WHERE id = $1
		`

		result, err := tx.Exec(query, id)
		if err != nil {
			return fmt.Errorf("failed to delete account: %w", err)
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
}

