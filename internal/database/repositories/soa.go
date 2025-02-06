package repositories

import (
	"banking-system/internal/database"
	"banking-system/internal/database/models"
	"context"
	"database/sql"
	"fmt"
	"math"
)

type SOARepository interface {
	GetSOA(userID int, request models.GenerateSOACustomRequest) (*models.SOA, error)
	SavePDF(pdfURL string, userID int) error
	GetGeneratedSOA(userID int) (*models.PaginatedResponse[models.SOA], error)
	GetSOAByID(soaID int) (*models.SOA, error)
}

type soaRepository struct {
	db database.Service
}

func NewSOARepository(db database.Service) SOARepository {
	return &soaRepository{db: db}
}

func (r *soaRepository) GetSOA(userID int, request models.GenerateSOACustomRequest) (*models.SOA, error) {
	// Get all transactions for the account and filter by the request

	transactionRepository := NewTransactionRepository(r.db)

	transactions, err := transactionRepository.GetTransactionsForSOA(userID, request)
	if err != nil {
		return nil, err
	}

	userRepository := NewUserRepository(r.db)
	user, err := userRepository.GetUser(userID)
	if err != nil {
		return nil, err
	}

	totalAmount := 0.0

	userBalance, err := userRepository.ViewBalance(userID)
	if err != nil {
		return nil, err
	}
	
	if request.Currency == "" {
		totalAmount = userBalance.BalancesByCurrency[request.Currency]
	} else {
		for _, account := range userBalance.Accounts {
			if account.Currency == request.Currency {
				totalAmount = account.Balance
			}
		}
	}

	if request.AccountID == 0 {
		totalAmount = userBalance.BalancesByCurrency[request.Currency]
	} else {
		for _, account := range userBalance.Accounts {
			if account.ID == request.AccountID {
				totalAmount = account.Balance
			}
		}
	}
	

	pdfURL, err := r.db.GeneratePDF(transactions, totalAmount, userID, user.FirstName + " " + user.LastName)
	if err != nil {
		return nil, err
	}

	return &models.SOA{
		PDFUrl: pdfURL,
	}, nil
}

func (r *soaRepository) SavePDF(pdfURL string, userID int) error {
	query := `
		INSERT INTO statements (pdf_url, user_id, created_at, statement_date, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := r.db.Exec(context.Background(), query, pdfURL, userID)
	return err
}

func (r *soaRepository) GetGeneratedSOA(userID int) (*models.PaginatedResponse[models.SOA], error) {
	var statements []models.SOA
	var totalRecords int64
	var response *models.PaginatedResponse[models.SOA]

	err := r.db.ExecTxReadOnly(context.Background(), func(tx *sql.Tx) error {
		// Get total count for pagination
		countQuery := `SELECT COUNT(*) FROM statements WHERE user_id = $1`
		err := tx.QueryRow(countQuery, userID).Scan(&totalRecords)
		if err != nil {
			return fmt.Errorf("failed to get total count: %w", err)
		}

		// Get paginated results
		query := `
			SELECT id, pdf_url, user_id, created_at, statement_date, updated_at
			FROM statements 
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`

		// Default pagination values if not provided
		pageSize := 10
		page := 1
		offset := (page - 1) * pageSize

		rows, err := tx.Query(query, userID, pageSize, offset)
		if err != nil {
			return fmt.Errorf("failed to query statements: %w", err)
		}
		defer rows.Close()

		statements = make([]models.SOA, 0)
		for rows.Next() {
			var s models.SOA
			if err := rows.Scan(
				&s.ID,
				&s.PDFUrl,
				&s.UserID,
				&s.CreatedAt,
				&s.StatementDate,
				&s.UpdatedAt,
			); err != nil {
				return fmt.Errorf("failed to scan statement: %w", err)
			}
			statements = append(statements, s)
		}

		if err = rows.Err(); err != nil {
			return fmt.Errorf("error iterating statements: %w", err)
		}

		// Calculate pagination metadata
		totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

		response = &models.PaginatedResponse[models.SOA]{
			Data: statements,
			Pagination: models.PaginationResponse{
				CurrentPage:  page,
				PageSize:     pageSize,
				TotalPages:   totalPages,
				TotalRecords: totalRecords,
			},
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get statements: %w", err)
	}

	return response, nil
}

func (r *soaRepository) GetSOAByID(soaID int) (*models.SOA, error) {
	query := `
		SELECT id, pdf_url, user_id, created_at, statement_date, updated_at
		FROM statements
		WHERE id = $1
	`

	var soa models.SOA
	err := r.db.QueryRow(context.Background(), query, soaID).Scan(
		&soa.ID,
		&soa.PDFUrl,
		&soa.UserID,
		&soa.CreatedAt,
		&soa.StatementDate,
		&soa.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &soa, nil
}


