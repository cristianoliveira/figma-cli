package api_test

import (
	"context"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example of a table-driven test using mock client
func TestGetFile_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		fileKey       string
		mockFile      *api.File
		mockError     error
		expectedFile  *api.File
		expectedError bool
	}{
		{
			name:    "success",
			fileKey: "abc123",
			mockFile: &api.File{
				Key:  "abc123",
				Name: "Test Design",
			},
			expectedFile: &api.File{
				Key:  "abc123",
				Name: "Test Design",
			},
		},
		{
			name:          "not found",
			fileKey:       "missing",
			mockError:     api.ErrNotFound,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &api.MockClient{}
			mockClient.On("GetFile", context.Background(), tt.fileKey, "").Return(tt.mockFile, tt.mockError)

			file, err := mockClient.GetFile(context.Background(), tt.fileKey, "")
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedFile, file)
			mockClient.AssertExpectations(t)
		})
	}
}

// Example of using fixtures in tests
func TestGetFile_WithFixture(t *testing.T) {
	// In a real test, you could load a fixture using testhelpers.LoadFixtureJSON
	// For demonstration, we'll create a mock response.
	mockClient := &api.MockClient{}
	expectedFile := &api.File{
		Key:          "fixture",
		Name:         "Fixture Design",
		LastModified: "2025-01-23T08:33:00Z",
	}
	mockClient.On("GetFile", context.Background(), "fixture", "").Return(expectedFile, nil)

	file, err := mockClient.GetFile(context.Background(), "fixture", "")
	require.NoError(t, err)
	assert.Equal(t, expectedFile, file)
	mockClient.AssertExpectations(t)
}
