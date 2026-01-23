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
func (rc *RetryClient) GetFile(ctx context.Context, fileKey string) (*File, error) {
	var result *File
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		file, err := rc.client.GetFile(ctx, fileKey)
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

// GetNodes implements Client.GetNodes with retry logic.
func (rc *RetryClient) GetNodes(ctx context.Context, fileKey string, nodeIDs []string) (map[string]*Node, error) {
	var result map[string]*Node
	err := WithRetry(ctx, rc.config, func(ctx context.Context) error {
		nodes, err := rc.client.GetNodes(ctx, fileKey, nodeIDs)
		if err != nil {
			return err
		}
		result = nodes
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
