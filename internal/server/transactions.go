package server

import (
	"banking-system/internal/database"
	"banking-system/internal/database/models"
	"banking-system/internal/database/repositories"
	"banking-system/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"banking-system/internal/lib"
)

type TransactionService struct {
	db database.Service
}

func NewTransactionService(db database.Service) *TransactionService {
	return &TransactionService{db: db}
}

// @Summary Deposit money into an account
// @Description Make a deposit transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body models.CreateTransactionRequest true "Deposit details"
// @Success 201 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /transaction/deposit [post]
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *TransactionService) Deposit(w http.ResponseWriter, r *http.Request, userID int) {
	start := time.Now()
	ctx := r.Context()
	
	var depositRequest models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&depositRequest); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate account ownership
	if err := s.validateAccountOwnership(ctx, depositRequest.AccountID, userID); err != nil {
		utils.WriteJSONError(w, http.StatusForbidden, "Unauthorized access to account", err)
		return
	}

	// Enhanced amount validation
	if err := validateTransactionAmount(depositRequest.Amount); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid amount", err)
		return
	}

	transactionRepository := repositories.NewTransactionRepository(s.db)
	transaction, err := transactionRepository.Deposit(depositRequest, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Deposit failed", err)
		return
	}

	// Record transaction metrics
	lib.RecordTransaction("deposit", depositRequest.Amount)
	
	// Update account balance metrics
	accountRepository := repositories.NewAccountRepository(s.db)
	account, err := accountRepository.GetAccount(depositRequest.AccountID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to update account balance", err)
		return
	}
	
	lib.RecordAccountBalance(account.Balance, string(account.Currency))

	// Record API latency
	duration := time.Since(start).Seconds()
	lib.RecordRequest(r.URL.Path, r.Method, http.StatusOK, duration)

	utils.WriteJSONResponse(w, http.StatusCreated, "Deposit successful", map[string]interface{}{
		"transaction_id": transaction["transaction_id"],
		"reference_id": transaction["reference_id"],
	})
}

// @Summary Withdraw money from an account
// @Description Make a withdrawal transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param transaction body models.CreateTransactionRequest true "Withdrawal details" default(models.CreateTransactionRequest{AccountID: 1, Amount: 1000, Type: "WITHDRAWAL"})
// @Success 201 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /transaction/withdraw [post]
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *TransactionService) Withdraw(w http.ResponseWriter, r *http.Request, userID int) {
	start := time.Now()
	ctx := r.Context()
	
	var withdrawRequest models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&withdrawRequest); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate account ownership
	if err := s.validateAccountOwnership(ctx, withdrawRequest.AccountID, userID); err != nil {
		utils.WriteJSONError(w, http.StatusForbidden, "Unauthorized access to account", err)
		return
	}

	// Enhanced amount validation
	if err := validateTransactionAmount(withdrawRequest.Amount); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid amount", err)
		return
	}

	transactionRepository := repositories.NewTransactionRepository(s.db)
	transaction, err := transactionRepository.Withdraw(withdrawRequest, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Withdrawal failed", err)
		return
	}

	// Record transaction metrics
	lib.RecordTransaction("withdraw", withdrawRequest.Amount)
	
	// Update account balance metrics
	accountRepository := repositories.NewAccountRepository(s.db)
	account, err := accountRepository.GetAccount(withdrawRequest.AccountID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to update account balance", err)
		return
	}
	lib.RecordAccountBalance(account.Balance, string(account.Currency))

	// Record API latency
	duration := time.Since(start).Seconds()
	lib.RecordRequest(r.URL.Path, r.Method, http.StatusOK, duration)

	utils.WriteJSONResponse(w, http.StatusCreated, "Withdrawal successful", map[string]interface{}{
		"transaction_id": transaction["transaction_id"],
		"reference_id": transaction["reference_id"],
	})
}

// @Summary Get all transactions of authenticated user
// @Description Get all transactions with optional filtering, sorting, and pagination
// @Tags transactions
// @Accept json
// @Produce json
// @Param filter query models.TransactionFilter false "Filter parameters"
// @Param sort query models.SortRequest false "Sort parameters"
// @Param pagination query models.PaginationRequest false "Pagination parameters"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /transaction [get]
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *TransactionService) GetTransactions(w http.ResponseWriter, r *http.Request, userID int) {
	// Parse query parameters for filter, sort, and pagination
	filter := &models.TransactionFilter{}
	sort := &models.SortRequest{}
	pagination := &models.PaginationRequest{
		Page:     1,
		PageSize: 10,
	}
	
	// Add query parameter parsing here if needed

	transactionRepository := repositories.NewTransactionRepository(s.db)
	paginatedResponse, err := transactionRepository.GetTransactions(userID, filter, sort, pagination)
	if err != nil {
		response := models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to get transactions",
			Data:       map[string]interface{}{
				"error": err.Error(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Transactions retrieved successfully",
		Data:       map[string]interface{}{
			"transactions": paginatedResponse.Data,
			"pagination":   paginatedResponse.Pagination,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// @Summary Get a specific transaction of authenticated user
// @Description Get detailed information about a specific transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param id query int true "Transaction ID"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 404 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /transaction/get [get]
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *TransactionService) GetTransaction(w http.ResponseWriter, r *http.Request, transactionID int) {
	transactionRepository := repositories.NewTransactionRepository(s.db)

	transaction, err := transactionRepository.GetTransaction(transactionID)
	if err != nil {
		response := models.Response{
			StatusCode: http.StatusInternalServerError,
			Success:    false,
			Message:    "Failed to get transaction",
			Data:       map[string]interface{}{
				"error": err.Error(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	if transaction.ID == 0 {
		response := models.Response{
			StatusCode: http.StatusNotFound,
			Success:    false,
			Message:    "Transaction not found",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.Response{
		StatusCode: http.StatusOK,
		Success:    true,
		Message:    "Transaction retrieved successfully",
		Data:       map[string]interface{}{
			"transaction": transaction,
		},
	}	

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *TransactionService) validateAccountOwnership(ctx context.Context, accountID, userID int) error {
	// Implement account ownership validation using the database service
	accountRepository := repositories.NewAccountRepository(s.db)
	account, err := accountRepository.GetAccount(accountID)
	if err != nil {
		return err
	}
	if account.UserID != userID {
		return fmt.Errorf("account does not belong to user")
	}
	return nil
}

func validateTransactionAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}
	if amount > 1000000 { // Add maximum limit
		return fmt.Errorf("amount exceeds maximum allowed")
	}
	return nil
}
