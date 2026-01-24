package api

import (
	"context"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoggingClient(t *testing.T) {
	mockClient := new(MockClient)
	logger := logging.NewNopLogger()
	client := NewLoggingClient(mockClient, logger)

	ctx := context.Background()
	fileKey := "test-key"

	// Test GetFile
	expectedFile := &File{Key: fileKey}
	mockClient.On("GetFile", ctx, fileKey, mock.Anything).Return(expectedFile, nil)

	file, err := client.GetFile(ctx, fileKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedFile, file)
	mockClient.AssertExpectations(t)

	// Test GetFile error
	mockClient.On("GetFile", ctx, "error-key", mock.Anything).Return(nil, ErrNotFound)
	_, err = client.GetFile(ctx, "error-key")
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)
	mockClient.AssertExpectations(t)
}

func TestLoggingClientWithRequestID(t *testing.T) {
	mockClient := new(MockClient)
	cfg := logging.DefaultConfig()
	cfg.Level = "debug"
	cfg.Format = "json"
	logger, err := logging.NewLogger(cfg)
	assert.NoError(t, err)
	defer func() { _ = logger.Sync() }()

	client := NewLoggingClient(mockClient, logger)

	ctx := logging.NewContextWithRequestID(context.Background(), "req-123")
	fileKey := "test-key"
	expectedFile := &File{Key: fileKey}
	mockClient.On("GetFile", ctx, fileKey, mock.Anything).Return(expectedFile, nil)

	file, err := client.GetFile(ctx, fileKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedFile, file)
	mockClient.AssertExpectations(t)
}
