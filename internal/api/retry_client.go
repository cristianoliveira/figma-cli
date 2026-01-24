package api

import (
	"context"
)

// RetryClient wraps a Client and adds retry logic for transient errors.
type RetryClient struct {
	client Client
	config RetryConfig
}

// NewRetryClient creates a new RetryClient with default retry configuration.
func NewRetryClient(client Client) *RetryClient {
	return &RetryClient{
		client: client,
		config: DefaultRetryConfig(),
	}
}

// NewRetryClientWithConfig creates a new RetryClient with custom configuration.
func NewRetryClientWithConfig(client Client, config RetryConfig) *RetryClient {
	return &RetryClient{
		client: client,
		config: config,
	}
}

// GetFile implements Client.GetFile with retry logic.
func (rc *RetryClient) GetFile(ctx context.Context, fileKey string, opts ...GetFileOption) (*File, error) {
	var result *File
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		file, err := rc.client.GetFile(ctx, fileKey, opts...)
		if err != nil {
			return err
		}
		result = file
		return nil
	})
	return result, err
}

// GetNode implements Client.GetNode with retry logic.
func (rc *RetryClient) GetNode(ctx context.Context, fileKey, nodeID string) (*Node, error) {
	var result *Node
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		node, err := rc.client.GetNode(ctx, fileKey, nodeID)
		if err != nil {
			return err
		}
		result = node
		return nil
	})
	return result, err
}

// GetFileNodes implements Client.GetFileNodes with retry logic.
func (rc *RetryClient) GetFileNodes(ctx context.Context, fileKey string, nodeIDs []string, opts ...GetFileNodesOption) (*FileNodesResponse, error) {
	var result *FileNodesResponse
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		response, err := rc.client.GetFileNodes(ctx, fileKey, nodeIDs, opts...)
		if err != nil {
			return err
		}
		result = response
		return nil
	})
	return result, err
}

// GetImage implements Client.GetImage with retry logic.
func (rc *RetryClient) GetImage(ctx context.Context, fileKey string, nodeIDs []string, options *ImageOptions) (map[string]string, error) {
	var result map[string]string
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		images, err := rc.client.GetImage(ctx, fileKey, nodeIDs, options)
		if err != nil {
			return err
		}
		result = images
		return nil
	})
	return result, err
}

// GetComments implements Client.GetComments with retry logic.
func (rc *RetryClient) GetComments(ctx context.Context, fileKey string) ([]*Comment, error) {
	var result []*Comment
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		comments, err := rc.client.GetComments(ctx, fileKey)
		if err != nil {
			return err
		}
		result = comments
		return nil
	})
	return result, err
}

// PostComment implements Client.PostComment with retry logic.
func (rc *RetryClient) PostComment(ctx context.Context, fileKey string, comment *CommentRequest) (*Comment, error) {
	var result *Comment
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		cmt, err := rc.client.PostComment(ctx, fileKey, comment)
		if err != nil {
			return err
		}
		result = cmt
		return nil
	})
	return result, err
}

// DeleteComment implements Client.DeleteComment with retry logic.
func (rc *RetryClient) DeleteComment(ctx context.Context, fileKey, commentID string) error {
	return WithRetry(ctx, rc.config, func(ctx context.Context) error {
		return rc.client.DeleteComment(ctx, fileKey, commentID)
	})
}

// GetCommentReactions implements Client.GetCommentReactions with retry logic.
func (rc *RetryClient) GetCommentReactions(ctx context.Context, fileKey, commentID string) ([]*Reaction, error) {
	var result []*Reaction
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		reactions, err := rc.client.GetCommentReactions(ctx, fileKey, commentID)
		if err != nil {
			return err
		}
		result = reactions
		return nil
	})
	return result, err
}

// PostCommentReaction implements Client.PostCommentReaction with retry logic.
func (rc *RetryClient) PostCommentReaction(ctx context.Context, fileKey, commentID, emoji string) error {
	return WithRetry(ctx, rc.config, func(ctx context.Context) error {
		return rc.client.PostCommentReaction(ctx, fileKey, commentID, emoji)
	})
}

// DeleteCommentReaction implements Client.DeleteCommentReaction with retry logic.
func (rc *RetryClient) DeleteCommentReaction(ctx context.Context, fileKey, commentID, emoji string) error {
	return WithRetry(ctx, rc.config, func(ctx context.Context) error {
		return rc.client.DeleteCommentReaction(ctx, fileKey, commentID, emoji)
	})
}

// GetFileMetadata implements Client.GetFileMetadata with retry logic.
func (rc *RetryClient) GetFileMetadata(ctx context.Context, fileKey string, opts ...GetFileMetadataOption) (*FileMeta, error) {
	var result *FileMeta
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		meta, err := rc.client.GetFileMetadata(ctx, fileKey, opts...)
		if err != nil {
			return err
		}
		result = meta
		return nil
	})
	return result, err
}

// GetFileVersions implements Client.GetFileVersions with retry logic.
func (rc *RetryClient) GetFileVersions(ctx context.Context, fileKey string, opts ...GetFileVersionsOption) ([]*Version, error) {
	var result []*Version
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		versions, err := rc.client.GetFileVersions(ctx, fileKey, opts...)
		if err != nil {
			return err
		}
		result = versions
		return nil
	})
	return result, err
}

// GetTeamStyles implements Client.GetTeamStyles with retry logic.
func (rc *RetryClient) GetTeamStyles(ctx context.Context, teamID string) ([]*Style, error) {
	var result []*Style
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		styles, err := rc.client.GetTeamStyles(ctx, teamID)
		if err != nil {
			return err
		}
		result = styles
		return nil
	})
	return result, err
}

// GetStyle implements Client.GetStyle with retry logic.
func (rc *RetryClient) GetStyle(ctx context.Context, styleKey string) (*Style, error) {
	var result *Style
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		style, err := rc.client.GetStyle(ctx, styleKey)
		if err != nil {
			return err
		}
		result = style
		return nil
	})
	return result, err
}

// GetFileStyles implements Client.GetFileStyles with retry logic.
func (rc *RetryClient) GetFileStyles(ctx context.Context, fileKey string) ([]*Style, error) {
	var result []*Style
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		styles, err := rc.client.GetFileStyles(ctx, fileKey)
		if err != nil {
			return err
		}
		result = styles
		return nil
	})
	return result, err
}

// GetTeamComponents implements Client.GetTeamComponents with retry logic.
func (rc *RetryClient) GetTeamComponents(ctx context.Context, teamID string) ([]*Component, error) {
	var result []*Component
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		components, err := rc.client.GetTeamComponents(ctx, teamID)
		if err != nil {
			return err
		}
		result = components
		return nil
	})
	return result, err
}

// GetComponent implements Client.GetComponent with retry logic.
func (rc *RetryClient) GetComponent(ctx context.Context, componentKey string) (*Component, error) {
	var result *Component
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		component, err := rc.client.GetComponent(ctx, componentKey)
		if err != nil {
			return err
		}
		result = component
		return nil
	})
	return result, err
}

// GetTeamComponentSets implements Client.GetTeamComponentSets with retry logic.
func (rc *RetryClient) GetTeamComponentSets(ctx context.Context, teamID string) ([]*ComponentSet, error) {
	var result []*ComponentSet
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		componentSets, err := rc.client.GetTeamComponentSets(ctx, teamID)
		if err != nil {
			return err
		}
		result = componentSets
		return nil
	})
	return result, err
}

// GetComponentSet implements Client.GetComponentSet with retry logic.
func (rc *RetryClient) GetComponentSet(ctx context.Context, componentSetKey string) (*ComponentSet, error) {
	var result *ComponentSet
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		componentSet, err := rc.client.GetComponentSet(ctx, componentSetKey)
		if err != nil {
			return err
		}
		result = componentSet
		return nil
	})
	return result, err
}

// GetFileComponents implements Client.GetFileComponents with retry logic.
func (rc *RetryClient) GetFileComponents(ctx context.Context, fileKey string) ([]*Component, error) {
	var result []*Component
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		components, err := rc.client.GetFileComponents(ctx, fileKey)
		if err != nil {
			return err
		}
		result = components
		return nil
	})
	return result, err
}

// GetTeamProjects implements Client.GetTeamProjects with retry logic.
func (rc *RetryClient) GetTeamProjects(ctx context.Context, teamID string) ([]*Project, error) {
	var result []*Project
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		projects, err := rc.client.GetTeamProjects(ctx, teamID)
		if err != nil {
			return err
		}
		result = projects
		return nil
	})
	return result, err
}

// GetProjectFiles implements Client.GetProjectFiles with retry logic.
func (rc *RetryClient) GetProjectFiles(ctx context.Context, projectID string) ([]*File, error) {
	var result []*File
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		files, err := rc.client.GetProjectFiles(ctx, projectID)
		if err != nil {
			return err
		}
		result = files
		return nil
	})
	return result, err
}

// GetMe implements Client.GetMe with retry logic.
func (rc *RetryClient) GetMe(ctx context.Context) (*User, error) {
	var result *User
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		user, err := rc.client.GetMe(ctx)
		if err != nil {
			return err
		}
		result = user
		return nil
	})
	return result, err
}

// GetFileVariables implements Client.GetFileVariables with retry logic.
func (rc *RetryClient) GetFileVariables(ctx context.Context, fileKey string) ([]*Variable, error) {
	var result []*Variable
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		variables, err := rc.client.GetFileVariables(ctx, fileKey)
		if err != nil {
			return err
		}
		result = variables
		return nil
	})
	return result, err
}

// GetTeamPublishedVariables implements Client.GetTeamPublishedVariables with retry logic.
func (rc *RetryClient) GetTeamPublishedVariables(ctx context.Context, teamID string) ([]*Variable, error) {
	var result []*Variable
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		variables, err := rc.client.GetTeamPublishedVariables(ctx, teamID)
		if err != nil {
			return err
		}
		result = variables
		return nil
	})
	return result, err
}

// PostFileVariables implements Client.PostFileVariables with retry logic.
func (rc *RetryClient) PostFileVariables(ctx context.Context, fileKey string, variables []*Variable) error {
	return WithRetry(ctx, rc.config, func(ctx context.Context) error {
		return rc.client.PostFileVariables(ctx, fileKey, variables)
	})
}

// GetFileDevResources implements Client.GetFileDevResources with retry logic.
func (rc *RetryClient) GetFileDevResources(ctx context.Context, fileKey string) ([]*DevResource, error) {
	var result []*DevResource
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		resources, err := rc.client.GetFileDevResources(ctx, fileKey)
		if err != nil {
			return err
		}
		result = resources
		return nil
	})
	return result, err
}

// PostDevResources implements Client.PostDevResources with retry logic.
func (rc *RetryClient) PostDevResources(ctx context.Context, resources []*DevResource) error {
	return WithRetry(ctx, rc.config, func(ctx context.Context) error {
		return rc.client.PostDevResources(ctx, resources)
	})
}

// PatchDevResources implements Client.PatchDevResources with retry logic.
func (rc *RetryClient) PatchDevResources(ctx context.Context, resources []*DevResource) error {
	return WithRetry(ctx, rc.config, func(ctx context.Context) error {
		return rc.client.PatchDevResources(ctx, resources)
	})
}

// GetWebhooks implements Client.GetWebhooks with retry logic.
func (rc *RetryClient) GetWebhooks(ctx context.Context) ([]*Webhook, error) {
	var result []*Webhook
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		webhooks, err := rc.client.GetWebhooks(ctx)
		if err != nil {
			return err
		}
		result = webhooks
		return nil
	})
	return result, err
}

// CreateWebhook implements Client.CreateWebhook with retry logic.
func (rc *RetryClient) CreateWebhook(ctx context.Context, webhook *WebhookRequest) (*Webhook, error) {
	var result *Webhook
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		wh, err := rc.client.CreateWebhook(ctx, webhook)
		if err != nil {
			return err
		}
		result = wh
		return nil
	})
	return result, err
}

// DeleteWebhook implements Client.DeleteWebhook with retry logic.
func (rc *RetryClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	return WithRetry(ctx, rc.config, func(ctx context.Context) error {
		return rc.client.DeleteWebhook(ctx, webhookID)
	})
}

// UpdateWebhook implements Client.UpdateWebhook with retry logic.
func (rc *RetryClient) UpdateWebhook(ctx context.Context, webhookID string, updates *WebhookUpdate) error {
	return WithRetry(ctx, rc.config, func(ctx context.Context) error {
		return rc.client.UpdateWebhook(ctx, webhookID, updates)
	})
}
