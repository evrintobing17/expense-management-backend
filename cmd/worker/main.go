package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evrintobing17/expense-management-backend/config"
	"github.com/evrintobing17/expense-management-backend/internal/expense/repository"
	"github.com/evrintobing17/expense-management-backend/internal/payment/service"
	"github.com/evrintobing17/expense-management-backend/internal/payment/worker"
	"github.com/evrintobing17/expense-management-backend/pkg/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewPostgresConnection(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	expenseRepo := repository.NewExpenseRepository(db)

	// Initialize services
	paymentService := service.NewPaymentService(cfg.PaymentAPIURL)

	// Initialize worker
	paymentWorker := worker.NewPaymentWorker(expenseRepo, paymentService, time.Duration(cfg.WorkerInterval)*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker in a goroutine
	go paymentWorker.Start(ctx)

	log.Printf("Payment worker started with interval %d seconds", cfg.WorkerInterval)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down worker...")

	cancel()

	// Give the worker a moment to finish
	time.Sleep(1 * time.Second)

	log.Println("Worker exited")
}
