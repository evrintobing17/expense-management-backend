package main

import (
	"flag"
	"log"

	"github.com/evrintobing17/expense-management-backend/config"
	"github.com/evrintobing17/expense-management-backend/pkg/database"
)

func main() {
	rollback := flag.Bool("rollback", false, "Rollback the last migration")
	useAdmin := flag.Bool("use-admin", false, "Use admin credentials for migration")
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Use admin credentials if specified
	dbUser := cfg.DBUser
	dbPassword := cfg.DBPassword
	if *useAdmin {
		dbUser = ""
		dbPassword = "" // or read from environment
		log.Println("Using admin credentials for migration")
	}

	// Initialize database
	db, err := database.NewPostgresConnection(cfg.DBHost, cfg.DBPort, dbUser, dbPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if *rollback {
		// Rollback the last migration
		err = database.RollbackMigrations(db)
		if err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
		log.Println("Migration rollback completed successfully")
	} else {
		// Run database migrations
		err = database.RunMigrations(db)
		if err != nil {
			log.Fatalf("Failed to run database migrations: %v", err)
		}
		log.Println("Database migrations completed successfully")
	}
}
