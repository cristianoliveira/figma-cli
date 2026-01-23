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

// DeleteComment mocks the DeleteComment method.
func (m *MockClient) DeleteComment(ctx context.Context, fileKey, commentID string) error {
	args := m.Called(ctx, fileKey, commentID)
	return args.Error(0)
}

// GetCommentReactions mocks the GetCommentReactions method.
func (m *MockClient) GetCommentReactions(ctx context.Context, fileKey, commentID string) ([]*Reaction, error) {
	args := m.Called(ctx, fileKey, commentID)
	if args.Get(0) != nil {
		return args.Get(0).([]*Reaction), args.Error(1)
	}
	return nil, args.Error(1)
}

// PostCommentReaction mocks the PostCommentReaction method.
func (m *MockClient) PostCommentReaction(ctx context.Context, fileKey, commentID, emoji string) error {
	args := m.Called(ctx, fileKey, commentID, emoji)
	return args.Error(0)
}

// DeleteCommentReaction mocks the DeleteCommentReaction method.
func (m *MockClient) DeleteCommentReaction(ctx context.Context, fileKey, commentID, emoji string) error {
	args := m.Called(ctx, fileKey, commentID, emoji)
	return args.Error(0)
}

// GetFileMeta mocks the GetFileMeta method.
func (m *MockClient) GetFileMeta(ctx context.Context, fileKey string) (*FileMeta, error) {
	args := m.Called(ctx, fileKey)
	if args.Get(0) != nil {
		return args.Get(0).(*FileMeta), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetFileVersions mocks the GetFileVersions method.
func (m *MockClient) GetFileVersions(ctx context.Context, fileKey string) ([]*Version, error) {
	args := m.Called(ctx, fileKey)
	if args.Get(0) != nil {
		return args.Get(0).([]*Version), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetTeamStyles mocks the GetTeamStyles method.
func (m *MockClient) GetTeamStyles(ctx context.Context, teamID string) ([]*Style, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) != nil {
		return args.Get(0).([]*Style), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetStyle mocks the GetStyle method.
func (m *MockClient) GetStyle(ctx context.Context, styleKey string) (*Style, error) {
	args := m.Called(ctx, styleKey)
	if args.Get(0) != nil {
		return args.Get(0).(*Style), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetFileStyles mocks the GetFileStyles method.
func (m *MockClient) GetFileStyles(ctx context.Context, fileKey string) ([]*Style, error) {
	args := m.Called(ctx, fileKey)
	if args.Get(0) != nil {
		return args.Get(0).([]*Style), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetTeamComponents mocks the GetTeamComponents method.
func (m *MockClient) GetTeamComponents(ctx context.Context, teamID string) ([]*Component, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) != nil {
		return args.Get(0).([]*Component), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetComponent mocks the GetComponent method.
func (m *MockClient) GetComponent(ctx context.Context, componentKey string) (*Component, error) {
	args := m.Called(ctx, componentKey)
	if args.Get(0) != nil {
		return args.Get(0).(*Component), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetTeamComponentSets mocks the GetTeamComponentSets method.
func (m *MockClient) GetTeamComponentSets(ctx context.Context, teamID string) ([]*ComponentSet, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) != nil {
		return args.Get(0).([]*ComponentSet), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetComponentSet mocks the GetComponentSet method.
func (m *MockClient) GetComponentSet(ctx context.Context, componentSetKey string) (*ComponentSet, error) {
	args := m.Called(ctx, componentSetKey)
	if args.Get(0) != nil {
		return args.Get(0).(*ComponentSet), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetFileComponents mocks the GetFileComponents method.
func (m *MockClient) GetFileComponents(ctx context.Context, fileKey string) ([]*Component, error) {
	args := m.Called(ctx, fileKey)
	if args.Get(0) != nil {
		return args.Get(0).([]*Component), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetTeamProjects mocks the GetTeamProjects method.
func (m *MockClient) GetTeamProjects(ctx context.Context, teamID string) ([]*Project, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) != nil {
		return args.Get(0).([]*Project), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetProjectFiles mocks the GetProjectFiles method.
func (m *MockClient) GetProjectFiles(ctx context.Context, projectID string) ([]*File, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) != nil {
		return args.Get(0).([]*File), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetMe mocks the GetMe method.
func (m *MockClient) GetMe(ctx context.Context) (*User, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).(*User), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetFileVariables mocks the GetFileVariables method.
func (m *MockClient) GetFileVariables(ctx context.Context, fileKey string) ([]*Variable, error) {
	args := m.Called(ctx, fileKey)
	if args.Get(0) != nil {
		return args.Get(0).([]*Variable), args.Error(1)
	}
	return nil, args.Error(1)
}

// GetTeamPublishedVariables mocks the GetTeamPublishedVariables method.
func (m *MockClient) GetTeamPublishedVariables(ctx context.Context, teamID string) ([]*Variable, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) != nil {
		return args.Get(0).([]*Variable), args.Error(1)
	}
	return nil, args.Error(1)
}

// PostFileVariables mocks the PostFileVariables method.
func (m *MockClient) PostFileVariables(ctx context.Context, fileKey string, variables []*Variable) error {
	args := m.Called(ctx, fileKey, variables)
	return args.Error(0)
}

// GetFileDevResources mocks the GetFileDevResources method.
func (m *MockClient) GetFileDevResources(ctx context.Context, fileKey string) ([]*DevResource, error) {
	args := m.Called(ctx, fileKey)
	if args.Get(0) != nil {
		return args.Get(0).([]*DevResource), args.Error(1)
	}
	return nil, args.Error(1)
}

// PostDevResources mocks the PostDevResources method.
func (m *MockClient) PostDevResources(ctx context.Context, resources []*DevResource) error {
	args := m.Called(ctx, resources)
	return args.Error(0)
}

// PatchDevResources mocks the PatchDevResources method.
func (m *MockClient) PatchDevResources(ctx context.Context, resources []*DevResource) error {
	args := m.Called(ctx, resources)
	return args.Error(0)
}

// GetWebhooks mocks the GetWebhooks method.
func (m *MockClient) GetWebhooks(ctx context.Context) ([]*Webhook, error) {
	args := m.Called(ctx)
	if args.Get(0) != nil {
		return args.Get(0).([]*Webhook), args.Error(1)
	}
	return nil, args.Error(1)
}

// CreateWebhook mocks the CreateWebhook method.
func (m *MockClient) CreateWebhook(ctx context.Context, webhook *WebhookRequest) (*Webhook, error) {
	args := m.Called(ctx, webhook)
	if args.Get(0) != nil {
		return args.Get(0).(*Webhook), args.Error(1)
	}
	return nil, args.Error(1)
}

// DeleteWebhook mocks the DeleteWebhook method.
func (m *MockClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	args := m.Called(ctx, webhookID)
	return args.Error(0)
}

// UpdateWebhook mocks the UpdateWebhook method.
func (m *MockClient) UpdateWebhook(ctx context.Context, webhookID string, updates *WebhookUpdate) error {
	args := m.Called(ctx, webhookID, updates)
	return args.Error(0)
}
