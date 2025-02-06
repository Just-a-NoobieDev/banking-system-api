package server

import (
	"banking-system/internal/database"
	"banking-system/internal/database/models"
	"banking-system/internal/database/repositories"
	"banking-system/internal/utils"
	"encoding/json"
	"net/http"
)

type UserService struct {
	db database.Service
}

func NewUserService(db database.Service) *UserService {
	return &UserService{db: db}
}

// GetUser gets user details
// @Summary Get user details
// @Description Get detailed information about a specific user
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=models.User} "User details retrieved successfully"
// @Failure 500 {object} models.Response{data=map[string]string} "Internal server error"
// @Router /user/me [get]
// @Tags user
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *UserService) GetUser(w http.ResponseWriter, r *http.Request, userID int) {
	userRepository := repositories.NewUserRepository(s.db)
	user, err := userRepository.GetUser(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "User retrieved successfully", user)
}

// ViewBalance gets user balance
// @Summary View user balance
// @Description Get the current balance for a specific user
// @Accept json
// @Produce json
// @Success 200 {object} models.Response{data=models.ViewBalanceResponse} "Balance retrieved successfully"
// @Failure 500 {object} models.Response{data=map[string]string} "Internal server error"
// @Router /user/view-balance [get]
// @Tags user
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *UserService) ViewBalance(w http.ResponseWriter, r *http.Request, userID int) {
	userRepository := repositories.NewUserRepository(s.db)
	balance, err := userRepository.ViewBalance(userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get balance", err)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "Balance retrieved successfully", balance)
}

// UpdateUser updates user details
// @Summary Update user information
// @Description Update details of a specific user
// @Accept json
// @Produce json
// @Param user body models.UpdateUserRequest true "User details to update"
// @Success 200 {object} models.Response{data=models.User} "User updated successfully"
// @Failure 400 {object} models.Response{data=map[string]string} "Invalid request"
// @Failure 500 {object} models.Response{data=map[string]string} "Internal server error"
// @Router /user/update-profile [put]
// @Tags user
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *UserService) UpdateUser(w http.ResponseWriter, r *http.Request, userID int) {
	var updateUserRequest models.UpdateUserRequest	
	if err := json.NewDecoder(r.Body).Decode(&updateUserRequest); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userRepository := repositories.NewUserRepository(s.db)
	user, err := userRepository.UpdateUser(updateUserRequest, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to update user", err)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "User updated successfully", user)
}

// UpdateUserPassword updates user password
// @Summary Update user password
// @Description Update the password for a specific user
// @Accept json
// @Produce json
// @Param password body models.UpdateUserPasswordRequest true "Password update details"
// @Success 200 {object} models.Response{data=models.User} "Password updated successfully"
// @Failure 400 {object} models.Response{data=map[string]string} "Invalid request"
// @Failure 500 {object} models.Response{data=map[string]string} "Internal server error"
// @Router /user/update-password [put]
// @Tags user
// @Security ApiKeyAuth
// @param Authorization header string true "Authorization" default(Bearer <Add access token here>)
func (s *UserService) UpdateUserPassword(w http.ResponseWriter, r *http.Request, userID int) {
	var updateUserPasswordRequest models.UpdateUserPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&updateUserPasswordRequest); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userRepository := repositories.NewUserRepository(s.db)
	user, err := userRepository.UpdateUserPassword(updateUserPasswordRequest, userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to update user password", err)
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, "User password updated successfully", user)
}