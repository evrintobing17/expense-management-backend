package database

import (
	"database/sql"
	"fmt"
	"log"
	"sort"

	_ "github.com/lib/pq"
)

type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

func RunMigrations(db *sql.DB) error {
	err := createMigrationsTable(db)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}

	availableMigrations, err := getAvailableMigrations()
	if err != nil {
		return fmt.Errorf("failed to get available migrations: %v", err)
	}

	for _, migration := range availableMigrations {
		if _, applied := appliedMigrations[migration.Version]; !applied {
			log.Printf("Applying migration: %s", migration.Name)

			_, err := db.Exec(migration.UpSQL)
			if err != nil {
				return fmt.Errorf("failed to apply migration %s: %v", migration.Name, err)
			}

			// Record migration
			err = recordMigration(db, migration.Version, migration.Name)
			if err != nil {
				return fmt.Errorf("failed to record migration %s: %v", migration.Name, err)
			}

			log.Printf("Applied migration: %s", migration.Name)
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[int]bool, error) {
	query := "SELECT version FROM schema_migrations ORDER BY version"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		err := rows.Scan(&version)
		if err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, nil
}

func getAvailableMigrations() ([]Migration, error) {
	// For simplicity, we'll define migrations in code
	// In a real application, you might read from files
	migrations := []Migration{
		{
			Version: 1,
			Name:    "initial_schema",
			UpSQL: `
				CREATE TABLE IF NOT EXISTS users (
					id SERIAL PRIMARY KEY,
					email VARCHAR(255) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					role VARCHAR(20) NOT NULL CHECK (role IN ('employee', 'manager')),
					password_hash VARCHAR(255) NOT NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);

				CREATE TABLE IF NOT EXISTS expenses (
					id SERIAL PRIMARY KEY,
					user_id INTEGER REFERENCES users(id),
					amount_idr INTEGER NOT NULL CHECK (amount_idr >= 10000 AND amount_idr <= 50000000),
					description TEXT NOT NULL,
					receipt_url VARCHAR(500),
					status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'awaiting_approval', 'approved', 'rejected', 'auto_approved', 'processing', 'completed', 'failed')),
					submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					processed_at TIMESTAMP,
					requires_approval BOOLEAN DEFAULT FALSE,
					auto_approved BOOLEAN DEFAULT FALSE
				);

				CREATE TABLE IF NOT EXISTS approvals (
					id SERIAL PRIMARY KEY,
					expense_id INTEGER REFERENCES expenses(id),
					approver_id INTEGER REFERENCES users(id),
					status VARCHAR(20) NOT NULL CHECK (status IN ('approved', 'rejected')),
					notes TEXT,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				);

				-- Insert sample data with properly hashed passwords (password is "password" for both users)
				INSERT INTO users (email, name, role, password_hash) VALUES
				('manager@example.com', 'Manager User', 'manager', '$2a$10$6uvHhDNhqrAqHiTWXSsx/emnFYDJySUHLtya7yRKVuFJfWzEViLaK'),
				('employee@example.com', 'Employee User', 'employee', '$2a$10$6uvHhDNhqrAqHiTWXSsx/emnFYDJySUHLtya7yRKVuFJfWzEViLaK')
				ON CONFLICT (email) DO NOTHING;
			`,
			DownSQL: `
				DROP TABLE IF EXISTS approvals;
				DROP TABLE IF EXISTS expenses;
				DROP TABLE IF EXISTS users;
			`,
		},
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func recordMigration(db *sql.DB, version int, name string) error {
	query := "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)"
	_, err := db.Exec(query, version, name)
	return err
}

func RollbackMigrations(db *sql.DB) error {
	query := "SELECT version, name FROM schema_migrations ORDER BY version DESC LIMIT 1"
	row := db.QueryRow(query)

	var version int
	var name string
	err := row.Scan(&version, &name)
	if err != nil {
		return fmt.Errorf("no migrations to rollback: %v", err)
	}

	migrations, err := getAvailableMigrations()
	if err != nil {
		return err
	}

	var migration *Migration
	for _, m := range migrations {
		if m.Version == version {
			migration = &m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration version %d not found", version)
	}

	log.Printf("Rolling back migration: %s", migration.Name)
	_, err = db.Exec(migration.DownSQL)
	if err != nil {
		return fmt.Errorf("failed to rollback migration %s: %v", migration.Name, err)
	}

	query = "DELETE FROM schema_migrations WHERE version = $1"
	_, err = db.Exec(query, version)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %v", err)
	}

	log.Printf("Rolled back migration: %s", migration.Name)
	return nil
}
