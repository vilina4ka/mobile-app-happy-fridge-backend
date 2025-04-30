package repository

import (
	"fmt"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/config"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/domain"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// dbStore holds the database connection and repository implementations.
type dbStore struct {
	db          *sqlx.DB
	userRepo    UserRepository
	fridgeRepo  FridgeRepository
	productRepo ProductRepository
}

// NewStore creates a new database store and initializes repositories.
func NewStore(cfg *config.Config, logger *logrus.Logger) (Store, error) {
	connStr := cfg.DatabaseURL
	if connStr == "" {
		return nil, fmt.Errorf("database URL is not configured")
	}

	logger.Infof("Connecting to database...")
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		logger.Errorf("Failed to connect to database: %v", err)
		return nil, domain.NewAppError(err, "Failed to connect to database", 500)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	if err := db.Ping(); err != nil {
		logger.Errorf("Failed to ping database: %v", err)
		db.Close() // Close the connection if ping fails
		return nil, domain.NewAppError(err, "Failed to ping database", 500)
	}

	logger.Info("Database connection established successfully")

	s := &dbStore{
		db: db,
	}

	s.userRepo = newSQLUserRepository(s.db)
	s.fridgeRepo = newSQLFridgeRepository(s.db)
	s.productRepo = newSQLProductRepository(s.db)

	return s, nil
}

// Close terminates the database connection.
func (s *dbStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Ping checks the database connection.
func (s *dbStore) Ping() error {
	if s.db != nil {
		return s.db.Ping()
	}
	return fmt.Errorf("database connection is not initialized")
}

// User returns the user repository implementation.
func (s *dbStore) User() UserRepository {
	return s.userRepo
}

// Fridge returns the fridge repository implementation.
func (s *dbStore) Fridge() FridgeRepository {
	return s.fridgeRepo
}

// Product returns the product repository implementation.
func (s *dbStore) Product() ProductRepository {
	return s.productRepo
}

// DB returns the underlying sqlx.DB object.
func (s *dbStore) DB() *sqlx.DB {
	return s.db
}
