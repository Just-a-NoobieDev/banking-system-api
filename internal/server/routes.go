package server

import (
	"banking-system/internal/database/models"
	"banking-system/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "banking-system/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

type wrappedResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *wrappedResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

// @title           Banking System API
// @version         1.0
// @description     This is a Banking System API.

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @tags auth
// @tag user
// @tag account
func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.Handle("/docs/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/docs/doc.json"),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DomID("swagger-ui"),
	))

	// Auth Routes
	mux.Handle("/api/auth/register", s.MethodGuard(http.HandlerFunc(s.authService.Register), http.MethodPost))
	mux.Handle("/api/auth/login", s.MethodGuard(http.HandlerFunc(s.authService.Login), http.MethodPost))

	// Account Routes all routes are protected
	mux.Handle("/api/account/create", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.accountService.CreateAccount(w, r, userID)
	})), http.MethodPost))

	mux.Handle("/api/account/get", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		accountID := r.URL.Query().Get("id")
		accountIDInt, err := strconv.Atoi(accountID)
		if err != nil {
			response := models.Response{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid account ID",
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		s.accountService.GetAccount(w, r, accountIDInt, userID)
	})), http.MethodGet))

	mux.Handle("/api/account/get-accounts", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.accountService.GetAccounts(w, r, userID)
	})), http.MethodGet))

	mux.Handle("/api/account/delete", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.accountService.DeleteAccount(w, r, userID)
	})), http.MethodDelete))

	// Transaction Routes all routes are protected
	mux.Handle("/api/transaction/deposit", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.transactionService.Deposit(w, r, userID)
	})), http.MethodPost))

	mux.Handle("/api/transaction/withdraw", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.transactionService.Withdraw(w, r, userID)
	})), http.MethodPost))

	mux.Handle("/api/transaction/", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.transactionService.GetTransactions(w, r, userID)
	})), http.MethodGet))

	mux.Handle("/api/transaction/get", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transactionID := r.URL.Query().Get("id")
		transactionIDInt, err := strconv.Atoi(transactionID)
		if err != nil {
			response := models.Response{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid transaction ID",
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		s.transactionService.GetTransaction(w, r, transactionIDInt)
	})), http.MethodGet))

	// User Routes all routes are protected
	mux.Handle("/api/user/view-balance", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.userService.ViewBalance(w, r, userID)
	})), http.MethodGet))

	mux.Handle("/api/user/update-profile", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.userService.UpdateUser(w, r, userID)
	})), http.MethodPut))

	mux.Handle("/api/user/update-password", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.userService.UpdateUserPassword(w, r, userID)
	})), http.MethodPut))

	mux.Handle("/api/user/me", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.userService.GetUser(w, r, userID)
	})), http.MethodGet))

	// SOA Routes all routes are protected
	mux.Handle("/api/soa/generate", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.soaService.GetSOA(w, r, userID)
	})), http.MethodPost))

	mux.Handle("/api/soa/generated", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		s.soaService.GetGeneratedSOA(w, r, userID)
	})), http.MethodGet))

	mux.Handle("/api/soa/download", s.MethodGuard(s.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		soaID := r.URL.Query().Get("id")
		soaIDInt, err := strconv.Atoi(soaID)
		if err != nil {
			response := models.Response{
				StatusCode: http.StatusBadRequest,
				Success:    false,
				Message:    "Invalid SOA ID",
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		s.soaService.DownloadSOA(w, r, soaIDInt)
	})), http.MethodGet))

	mux.Handle("/health", s.MethodGuard(http.HandlerFunc(s.healthHandler), http.MethodGet))

	// Move the root route to the end
	mux.Handle("/", s.MethodGuard(http.HandlerFunc(s.HelloWorldHandler), http.MethodGet))

	// Wrap the mux with CORS middleware
	return s.LoggerMiddleware(s.corsMiddleware(mux))
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

func (s *Server) AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("AuthGuard")
		token := r.Header.Get("Authorization")
		if token == "" {
			response := models.Response{
				StatusCode: http.StatusUnauthorized,
				Success: false,
				Message: "Unauthorized",
				Data: map[string]interface{}{
					"error": "Unauthorized",
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		log.Println(token)

		tokenString := strings.TrimPrefix(token, "Bearer ")
		
		userID, err := utils.ValidateToken(tokenString)
		if err != nil {
			response := models.Response{
				StatusCode: http.StatusUnauthorized,
				Success: false,
				Message: "Unauthorized",
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) LoggerMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrappedResponseWriter{ResponseWriter: w}
		next.ServeHTTP(wrapped, r)
		log.Printf("Request: %s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// MethodGuard middleware to restrict HTTP methods
func (s *Server) MethodGuard(next http.Handler, allowedMethods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, method := range allowedMethods {
			if r.Method == method {
				next.ServeHTTP(w, r)
				return
			}
		}
		
		response := models.Response{
			StatusCode: http.StatusMethodNotAllowed,
			Success:    false,
			Message:    "Method not allowed",
			Data: map[string]interface{}{
				"allowed_methods": allowedMethods,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
	})
}


