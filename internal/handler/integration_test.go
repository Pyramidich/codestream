package handler_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilya/codestream/internal/testutil"
)

func TestAuthFlow(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db := testutil.SetupTestDatabase(t)
	testutil.CleanupTestDatabase(t, db)

	server := testutil.NewTestServer(t, db, nil)
	defer server.Close()

	accessToken, _ := testutil.RegisterAndLogin(t, server)
	assert.NotEmpty(t, accessToken)
}

func TestProjectAndFileFlow(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db := testutil.SetupTestDatabase(t)
	testutil.CleanupTestDatabase(t, db)

	server := testutil.NewTestServer(t, db, nil)
	defer server.Close()

	accessToken, _ := testutil.RegisterAndLogin(t, server)
	require.NotEmpty(t, accessToken)

	projectID := testutil.CreateProject(t, server, accessToken, "My Project")
	assert.NotEmpty(t, projectID)

	fileID := testutil.CreateFile(t, server, accessToken, projectID, "main.go", "/main.go", "go")
	assert.NotEmpty(t, fileID)
}
