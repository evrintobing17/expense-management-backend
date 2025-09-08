package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/evrintobing17/expense-management-backend/config"
	approvalRepository "github.com/evrintobing17/expense-management-backend/internal/approval/repository"
	"github.com/evrintobing17/expense-management-backend/internal/expense/handler"

	userRepository "github.com/evrintobing17/expense-management-backend/internal/user/repository"
	"github.com/evrintobing17/expense-management-backend/pkg/database"

	authService "github.com/evrintobing17/expense-management-backend/internal/auth/service"
	authUsecase "github.com/evrintobing17/expense-management-backend/internal/auth/usecase"

	authHandler "github.com/evrintobing17/expense-management-backend/internal/auth/handler"

	expenseRepository "github.com/evrintobing17/expense-management-backend/internal/expense/repository"
	expenseUsecase "github.com/evrintobing17/expense-management-backend/internal/expense/usecase"

	healthHandler "github.com/evrintobing17/expense-management-backend/internal/health/handler"
	"github.com/evrintobing17/expense-management-backend/internal/middleware"
)

func main() {
	cfg := config.Load()

	db, err := database.NewPostgresConnection(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := userRepository.NewUserRepository(db)
	expenseRepo := expenseRepository.NewExpenseRepository(db)
	approvalRepo := approvalRepository.NewApprovalRepository(db)

	// Initialize services
	authService := authService.NewAuthService(userRepo, cfg.JWTSecret)

	// Initialize use cases
	authUseCase := authUsecase.NewAuthUseCase(authService)
	expenseUseCase := expenseUsecase.NewExpenseUseCase(expenseRepo, approvalRepo)

	// Initialize handlers
	authHandler := authHandler.NewAuthHandler(authUseCase)
	expenseHandler := handler.NewExpenseHandler(expenseUseCase)
	healthHandler := healthHandler.NewHealthHandler(db)

	// Initialize router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/health", healthHandler.Check).Methods("GET")

	// Protected routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.AuthMiddleware(authService))

	apiRouter.HandleFunc("/expenses", expenseHandler.CreateExpense).Methods("POST")
	apiRouter.HandleFunc("/expenses", expenseHandler.GetExpenses).Methods("GET")
	apiRouter.HandleFunc("/expenses/{id}", expenseHandler.GetExpense).Methods("GET")

	// Manager-only routes
	managerRouter := apiRouter.PathPrefix("").Subrouter()
	managerRouter.Use(middleware.ManagerOnlyMiddleware)

	managerRouter.HandleFunc("/expenses/{id}/approve", expenseHandler.ApproveExpense).Methods("PUT")
	managerRouter.HandleFunc("/expenses/{id}/reject", expenseHandler.RejectExpense).Methods("PUT")
	managerRouter.HandleFunc("/expenses/pending", expenseHandler.GetPendingApproval).Methods("GET")

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
