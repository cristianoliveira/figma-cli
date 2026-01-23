package api

import (
	"context"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/logging"
)

// LoggingClient wraps a Client and logs all requests and responses.
type LoggingClient struct {
	client Client
	logger logging.Logger
}

// NewLoggingClient creates a new LoggingClient.
func NewLoggingClient(client Client, logger logging.Logger) *LoggingClient {
	return &LoggingClient{
		client: client,
		logger: logger,
	}
}

// GetFile implements Client.GetFile with logging.
func (lc *LoggingClient) GetFile(ctx context.Context, fileKey string) (*File, error) {
	start := time.Now()
	lc.logger.Debug(ctx, "GetFile request", logging.String("file_key", fileKey))

	file, err := lc.client.GetFile(ctx, fileKey)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "GetFile failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetFile success", fields...)
	}
	return file, err
}

// GetNode implements Client.GetNode with logging.
func (lc *LoggingClient) GetNode(ctx context.Context, fileKey, nodeID string) (*Node, error) {
	start := time.Now()
	lc.logger.Debug(ctx, "GetNode request",
		logging.String("file_key", fileKey),
		logging.String("node_id", nodeID),
	)

	node, err := lc.client.GetNode(ctx, fileKey, nodeID)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.String("node_id", nodeID),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "GetNode failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetNode success", fields...)
	}
	return node, err
}

// GetNodes implements Client.GetNodes with logging.
func (lc *LoggingClient) GetNodes(ctx context.Context, fileKey string, nodeIDs []string) (map[string]*Node, error) {
	start := time.Now()
	lc.logger.Debug(ctx, "GetNodes request",
		logging.String("file_key", fileKey),
		logging.Any("node_ids", nodeIDs),
	)

	nodes, err := lc.client.GetNodes(ctx, fileKey, nodeIDs)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.Any("node_ids", nodeIDs),
		logging.Int("node_count", len(nodeIDs)),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "GetNodes failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetNodes success", append(fields,
			logging.Int("result_count", len(nodes)),
		)...)
	}
	return nodes, err
}

// GetImage implements Client.GetImage with logging.
func (lc *LoggingClient) GetImage(ctx context.Context, fileKey string, nodeIDs []string, options *ImageOptions) (map[string]string, error) {
	start := time.Now()
	lc.logger.Debug(ctx, "GetImage request",
		logging.String("file_key", fileKey),
		logging.Any("node_ids", nodeIDs),
		logging.Any("options", options),
	)

	images, err := lc.client.GetImage(ctx, fileKey, nodeIDs, options)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.Any("node_ids", nodeIDs),
		logging.Any("options", options),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "GetImage failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetImage success", append(fields,
			logging.Int("image_count", len(images)),
		)...)
	}
	return images, err
}

// GetComments implements Client.GetComments with logging.
func (lc *LoggingClient) GetComments(ctx context.Context, fileKey string) ([]*Comment, error) {
	start := time.Now()
	lc.logger.Debug(ctx, "GetComments request", logging.String("file_key", fileKey))

	comments, err := lc.client.GetComments(ctx, fileKey)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "GetComments failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetComments success", append(fields,
			logging.Int("comment_count", len(comments)),
		)...)
	}
	return comments, err
}

// PostComment implements Client.PostComment with logging.
func (lc *LoggingClient) PostComment(ctx context.Context, fileKey string, comment *CommentRequest) (*Comment, error) {
	start := time.Now()
	lc.logger.Debug(ctx, "PostComment request",
		logging.String("file_key", fileKey),
		logging.Any("comment", comment),
	)

	result, err := lc.client.PostComment(ctx, fileKey, comment)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.Any("comment", comment),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "PostComment failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "PostComment success", fields...)
	}
	return result, err
}
