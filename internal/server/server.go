package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"banking-system/internal/database"
)

type Server struct {
	port int

	db database.Service
	authService *AuthService
	accountService *AccountService
	transactionService *TransactionService
	userService *UserService
	soaService *SOAService
}



func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	db := database.New()
	authService := NewAuthService(db)
	accountService := NewAccountService(db)
	transactionService := NewTransactionService(db)
	userService := NewUserService(db)
	soaService := NewSOAService(db)

	NewServer := &Server{
		port: port,
		db: db,
		authService: authService,
		accountService: accountService,
		transactionService: transactionService,
		userService: userService,
		soaService: soaService,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
