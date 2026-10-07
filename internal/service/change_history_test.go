package service_test

import (
	"context"

	"github.com/google/uuid"
)

type noopChangeHistoryLogger struct{}

func (n *noopChangeHistoryLogger) Log(ctx context.Context, userID uuid.UUID, action string, fileID, projectID *uuid.UUID, metadata map[string]interface{}) {
	// no-op for testing
}
