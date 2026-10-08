package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTestDatabase connects to the test database, applies migrations and returns a GORM instance.
func SetupTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://codestream:codestream@localhost:5432/codestream_test?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := runMigrations(dsn); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

// CleanupTestDatabase truncates all tables and resets sequences.
func CleanupTestDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get raw db: %v", err)
	}

	tables := []string{
		"change_history",
		"document_versions",
		"files",
		"project_members",
		"projects",
		"refresh_tokens",
		"users",
	}

	for _, table := range tables {
		if _, err := sqlDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}

// NewTestRedis returns a Redis client for testing.
func NewTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379/1"
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("failed to parse redis url: %v", err)
	}

	return redis.NewClient(opts)
}

func runMigrations(dsn string) error {
	_, file, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(file), "..", "..", "migrations")
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}

	m, err := migrate.New("file://"+absPath, dsn)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
