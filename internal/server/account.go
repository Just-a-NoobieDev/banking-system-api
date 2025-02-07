package server

import (
	"banking-system/internal/database"
	"banking-system/internal/database/models"
	"banking-system/internal/database/repositories"
	"banking-system/internal/lib"
	"banking-system/internal/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type AccountService struct {
	db database.Service
}

func NewAccountService(db database.Service) *AccountService {
	return &AccountService{db: db}
}


// CreateAccount creates a new account for a user
// @Summary Create a new account
// @Description Create a new account for a user with specified currency
// @Accept json
// @Produce json
// @Param createAccountRequest body models.CreateAccountRequest true "Create account request"
// @Success 201 {object} models.Response{data=map[string]int} "Account created successfully"
// @Failure 400 {object} models.Response{data=map[string]string} "Invalid request"
// @Failure 500 {object} models.Response{data=map[string]string} "Internal server error"
// @Router /account/create [post]
// @Tags account
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *AccountService) CreateAccount(w http.ResponseWriter, r *http.Request, userID int) {
	start := time.Now()
	var createAccountRequest models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&createAccountRequest); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if createAccountRequest.Currency == "" {
		createAccountRequest.Currency = models.USD
	}

	accountRepository := repositories.NewAccountRepository(s.db)
	accountID, err := accountRepository.CreateAccount(createAccountRequest, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to create account", err)
		return
	}

	account, err := accountRepository.GetAccount(accountID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get account", err)
		return
	}

	// Record account balance after creation
	lib.RecordAccountBalance(account.Balance, string(account.Currency))
	lib.RecordNewAccount(string(account.Currency))
	
	// Record API latency
	duration := time.Since(start).Seconds()
	lib.RecordRequest(r.URL.Path, r.Method, http.StatusOK, duration)

	utils.WriteJSONResponse(w, http.StatusCreated, "Account created successfully", map[string]interface{}{
		"account_id": accountID,
	})
}

// GetAccounts gets all accounts for a user
// @Summary Get all accounts for a user
// @Description Get all accounts for a user with filtering, sorting, and pagination
// @Accept json
// @Produce json
// @Param minBalance query float64 false "Minimum balance filter"
// @Param maxBalance query float64 false "Maximum balance filter"
// @Param currency query string false "Currency filter (USD, EUR, GBP)"
// @Param dateFrom query string false "Date from filter (RFC3339)"
// @Param dateTo query string false "Date to filter (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(10)
// @Param sortField query string false "Sort field (balance, currency, created_at)"
// @Param sortDirection query string false "Sort direction (ASC, DESC)"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /account/get-accounts [get]
// @Tags account
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *AccountService) GetAccounts(w http.ResponseWriter, r *http.Request, userID int) {
	// Parse query parameters for filtering
	filter := &models.AccountFilter{}
	
	if minBalance := r.URL.Query().Get("minBalance"); minBalance != "" {
		val, err := strconv.ParseFloat(minBalance, 64)
		if err != nil {
			http.Error(w, "Invalid minBalance parameter", http.StatusBadRequest)
			return
		}
		filter.MinBalance = &val
	}

	if maxBalance := r.URL.Query().Get("maxBalance"); maxBalance != "" {
		val, err := strconv.ParseFloat(maxBalance, 64)
		if err != nil {
			http.Error(w, "Invalid maxBalance parameter", http.StatusBadRequest)
			return
		}
		filter.MaxBalance = &val
	}

	if currency := r.URL.Query().Get("currency"); currency != "" {
		curr := models.Currency(currency)
		filter.Currency = &curr
	}

	if dateFrom := r.URL.Query().Get("dateFrom"); dateFrom != "" {
		date, err := time.Parse(time.RFC3339, dateFrom)
		if err != nil {
			http.Error(w, "Invalid dateFrom parameter", http.StatusBadRequest)
			return
		}
		filter.DateFrom = &date
	}

	if dateTo := r.URL.Query().Get("dateTo"); dateTo != "" {
		date, err := time.Parse(time.RFC3339, dateTo)
		if err != nil {
			http.Error(w, "Invalid dateTo parameter", http.StatusBadRequest)
			return
		}
		filter.DateTo = &date
	}

	// Parse pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
			page = val
		}
	}

	pageSize := 10
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 {
			pageSize = val
		}
	}

	// Parse sorting parameters
	sort := &models.SortRequest{
		Field:     r.URL.Query().Get("sortField"),
		Direction: r.URL.Query().Get("sortDirection"),
	}

	pagination := &models.PaginationRequest{
		Page:     page,
		PageSize: pageSize,
	}

	accountRepository := repositories.NewAccountRepository(s.db)
	result, err := accountRepository.GetAccounts(userID, filter, sort, pagination)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get accounts", err)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "Accounts fetched successfully", result)
}

// GetAccount gets an account for a user
// @Summary Get an account by ID
// @Description Get detailed information about a specific account
// @Accept json
// @Produce json
// @Param id query int true "Account ID"
// @Success 200 {object} models.Response "Account details retrieved successfully"
// @Failure 403 {object} models.Response "Unauthorized access to account"
// @Failure 404 {object} models.Response "Account not found"
// @Failure 500 {object} models.Response "Internal server error"
// @Router /account/get [get]
// @Tags account
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *AccountService) GetAccount(w http.ResponseWriter, r *http.Request, accountID int, userID int) {
	start := time.Now()
	accountRepository := repositories.NewAccountRepository(s.db)
	account, err := accountRepository.GetAccount(accountID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get account", err)
		return
	}

	// Record account balance on retrieval
	lib.RecordAccountBalance(account.Balance, string(account.Currency))
	
	// Record API latency
	duration := time.Since(start).Seconds()
	lib.RecordRequest(r.URL.Path, r.Method, http.StatusOK, duration)

	utils.WriteJSONResponse(w, http.StatusOK, "Account fetched successfully", account)
}

// DeleteAccount deletes an account
// @Summary Delete an account
// @Description Permanently delete an account and all associated data
// @Accept json
// @Produce json
// @Param id query int true "Account ID"
// @Success 200 {object} models.Response "Account deleted successfully"
// @Failure 404 {object} models.Response{data=map[string]string} "Account not found"
// @Failure 500 {object} models.Response{data=map[string]string} "Internal server error"
// @Router /account/delete [delete]
// @Tags account
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *AccountService) DeleteAccount(w http.ResponseWriter, r *http.Request, userID int) {
	accountID := r.URL.Query().Get("id")
	accountIDInt, err := strconv.Atoi(accountID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid account ID", err)
		return
	}

	// Validate account ownership
	accountRepository := repositories.NewAccountRepository(s.db)
	account, err := accountRepository.GetAccount(accountIDInt)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get account", err)
		return
	}

	if account.UserID != userID {
		utils.WriteJSONError(w, http.StatusForbidden, "Unauthorized access to account", nil)
		return
	}

	err = accountRepository.DeleteAccount(accountIDInt)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to delete account", err)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "Account deleted successfully", nil)
}


