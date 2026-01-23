package api

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Client interface.
type MockClient struct {
	mock.Mock
}

// GetFile mocks the GetFile method.
func (m *MockClient) GetFile(ctx context.Context, fileKey string) (*File, error) {
	args := m.Called(ctx, fileKey)
	if args.Get(0) != nil {
		return args.Get(0).(*File), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetNode mocks the GetNode method.
func (m *MockClient) GetNode(ctx context.Context, fileKey, nodeID string) (*Node, error) {
	args := m.Called(ctx, fileKey, nodeID)
	if args.Get(0) != nil {
		return args.Get(0).(*Node), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetNodes mocks the GetNodes method.
func (m *MockClient) GetNodes(ctx context.Context, fileKey string, nodeIDs []string) (map[string]*Node, error) {
	args := m.Called(ctx, fileKey, nodeIDs)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]*Node), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetImage mocks the GetImage method.
func (m *MockClient) GetImage(ctx context.Context, fileKey string, nodeIDs []string, options *ImageOptions) (map[string]string, error) {
	args := m.Called(ctx, fileKey, nodeIDs, options)
	if args.Get(0) != nil {
		return args.Get(0).(map[string]string), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetComments mocks the GetComments method.
func (m *MockClient) GetComments(ctx context.Context, fileKey string) ([]*Comment, error) {
	args := m.Called(ctx, fileKey)
	if args.Get(0) != nil {
		return args.Get(0).([]*Comment), args.Error(1)
	}
	return nil, args.Error(1)
}

// PostComment mocks the PostComment method.
func (m *MockClient) PostComment(ctx context.Context, fileKey string, comment *CommentRequest) (*Comment, error) {
	args := m.Called(ctx, fileKey, comment)
	if args.Get(0) != nil {
		return args.Get(0).(*Comment), args.Error(1)
	}
	return nil, args.Error(1)
}
