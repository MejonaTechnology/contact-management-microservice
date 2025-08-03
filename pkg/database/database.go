package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	User            string
	Password        string
	DBName          string
	Port            string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	SSLMode         bool
	Charset         string
	ParseTime       bool
	Location        string
}

// InitDB initializes the database connection
func InitDB() error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load database config: %v", err)
	}

	// Build DSN (Data Source Name)
	dsn := buildDSN(config)

	// Configure GORM logger
	gormLogger := logger.Default
	if os.Getenv("APP_DEBUG") == "true" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// Open database connection
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // Use plural table names
		},
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL database: %v", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	DB = db
	log.Println("Database connection established successfully")
	return nil
}

// loadConfig loads database configuration from environment variables
func loadConfig() (*DatabaseConfig, error) {
	config := &DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		User:            getEnv("DB_USER", "root"),
		Password:        getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", "mejona_contacts"),
		Port:            getEnv("DB_PORT", "3306"),
		Charset:         getEnv("DB_CHARSET", "utf8mb4"),
		ParseTime:       getEnvBool("DB_PARSE_TIME", true),
		Location:        getEnv("DB_LOCATION", "UTC"),
		SSLMode:         getEnvBool("DB_SSL_MODE", false),
	}

	// Parse integer configurations
	var err error
	config.MaxIdleConns, err = strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %v", err)
	}

	config.MaxOpenConns, err = strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "100"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %v", err)
	}

	connMaxLifetimeSeconds, err := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME", "3600"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME: %v", err)
	}
	config.ConnMaxLifetime = time.Duration(connMaxLifetimeSeconds) * time.Second

	return config, nil
}

// buildDSN builds the MySQL Data Source Name
func buildDSN(config *DatabaseConfig) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.Charset,
		config.ParseTime,
		config.Location,
	)

	// Add SSL configuration
	if config.SSLMode {
		dsn += "&tls=true"
	} else {
		dsn += "&tls=false"
	}

	// Add additional MySQL-specific parameters
	dsn += "&sql_mode='STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO'"
	dsn += "&timeout=30s"
	dsn += "&readTimeout=30s"
	dsn += "&writeTimeout=30s"

	return dsn
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL database: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %v", err)
	}

	log.Println("Database connection closed successfully")
	return nil
}

// IsConnected checks if the database connection is active
func IsConnected() bool {
	if DB == nil {
		return false
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return false
	}

	if err := sqlDB.Ping(); err != nil {
		return false
	}

	return true
}

// GetConnectionStats returns database connection statistics
func GetConnectionStats() map[string]interface{} {
	if DB == nil {
		return map[string]interface{}{
			"connected": false,
		}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return map[string]interface{}{
			"connected": false,
			"error":     err.Error(),
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"connected":        true,
		"open_connections": stats.OpenConnections,
		"in_use":          stats.InUse,
		"idle":            stats.Idle,
		"wait_count":      stats.WaitCount,
		"wait_duration":   stats.WaitDuration.String(),
		"max_idle_closed": stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// Transaction wraps a function in a database transaction
func Transaction(fn func(*gorm.DB) error) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	tx := DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback().Error; rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// Paginate adds pagination to a GORM query
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}

		if pageSize <= 0 {
			pageSize = 10
		}

		if pageSize > 100 {
			pageSize = 100
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// AutoMigrate runs database migrations for all models
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Import models here to avoid circular imports
	// This would typically be done in the main application
	log.Println("Auto-migration should be called from main application with all models")
	return nil
}

// Utility functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// HealthCheck performs a basic database health check
func HealthCheck() map[string]interface{} {
	result := map[string]interface{}{
		"database": "unknown",
		"status":   "unknown",
	}

	if DB == nil {
		result["database"] = "mysql"
		result["status"] = "disconnected"
		result["error"] = "database not initialized"
		return result
	}

	sqlDB, err := DB.DB()
	if err != nil {
		result["database"] = "mysql"
		result["status"] = "error"
		result["error"] = err.Error()
		return result
	}

	// Ping database
	start := time.Now()
	if err := sqlDB.Ping(); err != nil {
		result["database"] = "mysql"
		result["status"] = "unhealthy"
		result["error"] = err.Error()
		return result
	}
	pingDuration := time.Since(start)

	// Get connection stats
	stats := sqlDB.Stats()

	result["database"] = "mysql"
	result["status"] = "healthy"
	result["ping_duration"] = pingDuration.String()
	result["connections"] = map[string]interface{}{
		"open":   stats.OpenConnections,
		"in_use": stats.InUse,
		"idle":   stats.Idle,
	}

	return result
}

// TestQuery performs a simple test query to verify database functionality
func TestQuery() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Perform a simple query to test database functionality
	var result int
	if err := DB.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("test query failed: %v", err)
	}

	if result != 1 {
		return fmt.Errorf("test query returned unexpected result: %d", result)
	}

	return nil
}