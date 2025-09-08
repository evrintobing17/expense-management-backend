package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgresConnection(host, port, user, password, dbname string) (*sql.DB, error) {
	adminConnStr := fmt.Sprintf("host=%s port=%s user=postgres password=postgres sslmode=disable",
		host, port)
	
	adminDb, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open admin database connection: %v", err)
	}
	defer adminDb.Close()
	
	var userExists bool
	err = adminDb.QueryRow(
		"SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1)",
		user,
	).Scan(&userExists)
	
	if err != nil {
		log.Printf("Warning: Could not check if user exists: %v", err)
	}
	
	if !userExists {
		log.Printf("User %s does not exist, attempting to create it", user)
		_, err = adminDb.Exec(
			fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", user, password),
		)
		if err != nil {
			log.Printf("Warning: Could not create user: %v", err)
		} else {
			log.Printf("User %s created successfully", user)
		}
	}
	
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	var connectionError error
	for i := 0; i < 10; i++ {
		connectionError = db.Ping()
		if connectionError == nil {
			break
		}
		log.Printf("Database connection attempt %d failed: %v", i+1, connectionError)
		time.Sleep(2 * time.Second)
	}
	
	if connectionError != nil {
		return nil, fmt.Errorf("failed to connect to database after retries: %v", connectionError)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}