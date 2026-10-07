package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupTestDatabase(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db := SetupTestDatabase(t)
	assert.NotNil(t, db)

	CleanupTestDatabase(t, db)
}
