package main

import (
	"contact-service/pkg/database"
	"contact-service/pkg/logger"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	logger.InitLogger()

	// Initialize database connection
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate/main.go [up|down|status|create <name>]")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		runMigrationsUp()
	case "down":
		runMigrationsDown()
	case "status":
		showMigrationStatus()
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run cmd/migrate/main.go create <migration_name>")
			os.Exit(1)
		}
		createMigration(os.Args[2])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: up, down, status, create")
		os.Exit(1)
	}
}

// runMigrationsUp executes all pending migrations
func runMigrationsUp() {
	fmt.Println("Running migrations up...")
	
	// Create migrations table if it doesn't exist
	createMigrationsTable()
	
	// Get list of migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		log.Fatal("Failed to get migration files:", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations()
	if err != nil {
		log.Fatal("Failed to get applied migrations:", err)
	}

	executed := 0
	for _, file := range migrationFiles {
		migrationName := strings.TrimSuffix(filepath.Base(file), ".sql")
		
		// Check if already applied
		if _, exists := appliedMigrations[migrationName]; exists {
			fmt.Printf("SKIP: %s (already applied)\n", migrationName)
			continue
		}

		// Execute migration
		if err := executeMigrationFile(file); err != nil {
			log.Fatalf("Failed to execute migration %s: %v", migrationName, err)
		}

		// Record migration as applied
		if err := recordMigration(migrationName); err != nil {
			log.Fatalf("Failed to record migration %s: %v", migrationName, err)
		}

		fmt.Printf("APPLIED: %s\n", migrationName)
		executed++
	}

	if executed == 0 {
		fmt.Println("No pending migrations to apply.")
	} else {
		fmt.Printf("Successfully applied %d migration(s).\n", executed)
	}
}

// runMigrationsDown rolls back the last migration
func runMigrationsDown() {
	fmt.Println("Rolling back last migration...")
	
	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations()
	if err != nil {
		log.Fatal("Failed to get applied migrations:", err)
	}

	if len(appliedMigrations) == 0 {
		fmt.Println("No migrations to roll back.")
		return
	}

	// Find the last applied migration
	var lastMigration string
	for migration := range appliedMigrations {
		if lastMigration == "" || migration > lastMigration {
			lastMigration = migration
		}
	}

	// Remove from migrations table
	if err := removeMigration(lastMigration); err != nil {
		log.Fatalf("Failed to remove migration record %s: %v", lastMigration, err)
	}

	fmt.Printf("ROLLED BACK: %s\n", lastMigration)
	fmt.Println("Note: This only removes the migration record. You may need to manually undo database changes.")
}

// showMigrationStatus shows the status of all migrations
func showMigrationStatus() {
	fmt.Println("Migration Status:")
	fmt.Println("================")

	// Get migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		log.Fatal("Failed to get migration files:", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations()
	if err != nil {
		log.Fatal("Failed to get applied migrations:", err)
	}

	for _, file := range migrationFiles {
		migrationName := strings.TrimSuffix(filepath.Base(file), ".sql")
		status := "PENDING"
		
		if appliedTime, exists := appliedMigrations[migrationName]; exists {
			status = fmt.Sprintf("APPLIED (%s)", appliedTime.Format("2006-01-02 15:04:05"))
		}

		fmt.Printf("%-50s %s\n", migrationName, status)
	}
}

// createMigration creates a new migration file
func createMigration(name string) {
	// Generate timestamp
	timestamp := getCurrentTimestamp()
	fileName := fmt.Sprintf("%s_%s.sql", timestamp, strings.ReplaceAll(name, " ", "_"))
	filePath := filepath.Join("migrations", fileName)

	// Create migration template
	template := fmt.Sprintf(`-- Migration: %s
-- Created: %s
-- Description: %s

-- Add your SQL statements here
-- Example:
-- CREATE TABLE IF NOT EXISTS example_table (
--     id INT PRIMARY KEY AUTO_INCREMENT,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );
`, name, timestamp, name)

	// Write file
	if err := os.WriteFile(filePath, []byte(template), 0644); err != nil {
		log.Fatalf("Failed to create migration file: %v", err)
	}

	fmt.Printf("Created migration file: %s\n", filePath)
}

// Helper functions

func createMigrationsTable() {
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		id INT PRIMARY KEY AUTO_INCREMENT,
		migration_name VARCHAR(255) NOT NULL UNIQUE,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_migration_name (migration_name),
		INDEX idx_applied_at (applied_at)
	)`

	if err := database.DB.Exec(query).Error; err != nil {
		log.Fatal("Failed to create migrations table:", err)
	}
}

func getMigrationFiles() ([]string, error) {
	var files []string
	
	err := filepath.Walk("migrations", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			files = append(files, path)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Sort files to ensure proper order
	sort.Strings(files)
	return files, nil
}

func getAppliedMigrations() (map[string]time.Time, error) {
	var migrations []struct {
		MigrationName string    `gorm:"column:migration_name"`
		AppliedAt     time.Time `gorm:"column:applied_at"`
	}

	if err := database.DB.Table("migrations").Find(&migrations).Error; err != nil {
		return nil, err
	}

	result := make(map[string]time.Time)
	for _, m := range migrations {
		result[m.MigrationName] = m.AppliedAt
	}

	return result, nil
}

func executeMigrationFile(filePath string) error {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	// Split by semicolons and execute each statement
	statements := strings.Split(string(content), ";")
	
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}

		if err := database.DB.Exec(statement).Error; err != nil {
			return fmt.Errorf("failed to execute statement: %v", err)
		}
	}

	return nil
}

func recordMigration(migrationName string) error {
	query := "INSERT INTO migrations (migration_name) VALUES (?)"
	return database.DB.Exec(query, migrationName).Error
}

func removeMigration(migrationName string) error {
	query := "DELETE FROM migrations WHERE migration_name = ?"
	return database.DB.Exec(query, migrationName).Error
}

func getCurrentTimestamp() string {
	return time.Now().Format("20060102150405")
}

