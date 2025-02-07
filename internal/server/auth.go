package server

import (
	"banking-system/internal/database"
	"banking-system/internal/database/models"
	"banking-system/internal/database/repositories"
	"banking-system/utils"
	"encoding/json"
	"net/http"
	"time"

	"banking-system/internal/lib"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db database.Service
}

func NewAuthService(db database.Service) *AuthService {
	return &AuthService{db: db}
}

// Register a new user
// @Summary Register a new user
// @Description Register a new user with email and password
// @Accept json
// @Produce json
// @Param user body models.CreateUserRequest true "User details"
// @Success 201 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /auth/register [post]
// @Tags auth
func (s *AuthService) Register(w http.ResponseWriter, r *http.Request) {
	var user models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userRepository := repositories.NewUserRepository(s.db)
	createdUser, err := userRepository.CreateUser(user)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "User creation failed", err)
		return
	}

	token, err := utils.GenerateToken(createdUser.ID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	// After successful registration, update active users count
	userCount, err := userRepository.GetUserCount()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get user count", err)
		return
	}
	lib.UpdateActiveUsers(userCount)

	utils.WriteJSONResponse(w, http.StatusCreated, "User created successfully", map[string]interface{}{
		"access_token": token,
	})
}

// Login a user
// @Summary Login a user
// @Description Login a user with email and password
// @Accept json
// @Produce json
// @Param user body models.LoginRequest true "User details"
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response
// @Failure 401 {object} models.Response
// @Failure 500 {object} models.Response
// @Router /auth/login [post]
// @Tags auth
func (s *AuthService) Login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var loginRequest models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userRepository := repositories.NewUserRepository(s.db)
	user, err := userRepository.GetUserByEmail(loginRequest.Email)
	if err != nil {
		lib.RecordLoginAttempt(false)
		response := models.Response{
			StatusCode: http.StatusUnauthorized,
			Success:    false,
			Message:    "Invalid credentials",
			Data:       map[string]interface{}{
				"error": "Invalid credentials",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		lib.RecordLoginAttempt(false)
		response := models.Response{
			StatusCode: http.StatusUnauthorized,
			Success:    false,
			Message:    "Invalid credentials",
			Data:       map[string]interface{}{
				"error": "Invalid credentials",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}
	lib.RecordLoginAttempt(true)

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	// Increment active users on successful login
	lib.IncrementActiveUsers()
	
	// Record API latency
	duration := time.Since(start).Seconds()
	lib.RecordRequest(r.URL.Path, r.Method, http.StatusOK, duration)

	utils.WriteJSONResponse(w, http.StatusOK, "User logged in successfully", map[string]interface{}{
		"access_token": token,
	})
}

// @Router /auth/logout [post]
// @Tags auth
// @Summary Logout a user
// @Description Logout a user
// @Accept json
// @Produce json
// @Success 200 {object} models.Response
// @Failure 500 {object} models.Response
func (s *AuthService) Logout(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// Decrement active users on logout
	lib.DecrementActiveUsers()
	
	// Record API latency
	duration := time.Since(start).Seconds()
	lib.RecordRequest(r.URL.Path, r.Method, http.StatusOK, duration)

	utils.WriteJSONResponse(w, http.StatusOK, "User logged out successfully", nil)
}
