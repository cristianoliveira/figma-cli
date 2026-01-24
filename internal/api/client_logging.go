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
func (lc *LoggingClient) GetFile(ctx context.Context, fileKey string, opts ...GetFileOption) (*File, error) {
	options := ApplyGetFileOptions(opts)
	start := time.Now()
	lc.logger.Debug(ctx, "GetFile request", logging.String("file_key", fileKey), logging.String("branch", options.Branch))

	file, err := lc.client.GetFile(ctx, fileKey, opts...)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.String("branch", options.Branch),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "GetFile failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetFile success", fields...)
	}
	return file, err
}

// GetFileMetadata implements Client.GetFileMetadata with logging.
func (lc *LoggingClient) GetFileMetadata(ctx context.Context, fileKey string, opts ...GetFileMetadataOption) (*FileMeta, error) {
	options := ApplyGetFileMetadataOptions(opts)
	start := time.Now()
	lc.logger.Debug(ctx, "GetFileMetadata request", logging.String("file_key", fileKey), logging.String("branch", options.Branch))

	meta, err := lc.client.GetFileMetadata(ctx, fileKey, opts...)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.String("branch", options.Branch),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "GetFileMetadata failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetFileMetadata success", fields...)
	}
	return meta, err
}

// GetFileVersions implements Client.GetFileVersions with logging.
func (lc *LoggingClient) GetFileVersions(ctx context.Context, fileKey string, opts ...GetFileVersionsOption) ([]*Version, error) {
	options := ApplyGetFileVersionsOptions(opts)
	start := time.Now()
	lc.logger.Debug(ctx, "GetFileVersions request",
		logging.String("file_key", fileKey),
		logging.Int("page_size", options.PageSize),
		logging.String("before", options.Before),
		logging.String("after", options.After),
		logging.String("branch", options.Branch),
	)

	versions, err := lc.client.GetFileVersions(ctx, fileKey, opts...)

	duration := time.Since(start)
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.Int("page_size", options.PageSize),
		logging.String("before", options.Before),
		logging.String("after", options.After),
		logging.String("branch", options.Branch),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	}
	if err != nil {
		lc.logger.Error(ctx, "GetFileVersions failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetFileVersions success", append(fields,
			logging.Int("version_count", len(versions)),
		)...)
	}
	return versions, err
}

// GetNode implements Client.GetNode with logging.
func (lc *LoggingClient) GetNode(ctx context.Context, fileKey, nodeID string, opts ...GetNodeOption) (*Node, error) {
	options := ApplyGetNodeOptions(opts)
	start := time.Now()
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.String("node_id", nodeID),
	}
	if options.Branch != "" {
		fields = append(fields, logging.String("branch", options.Branch))
	}
	if options.Depth != nil {
		fields = append(fields, logging.Int("depth", *options.Depth))
	}
	if options.Geometry != nil {
		fields = append(fields, logging.Bool("geometry", *options.Geometry))
	}
	lc.logger.Debug(ctx, "GetNode request", fields...)

	node, err := lc.client.GetNode(ctx, fileKey, nodeID, opts...)

	duration := time.Since(start)
	fields = append(fields, logging.Float64("duration_ms", duration.Seconds()*1000))
	if err != nil {
		lc.logger.Error(ctx, "GetNode failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetNode success", fields...)
	}
	return node, err
}

// GetFileNodes implements Client.GetFileNodes with logging.
func (lc *LoggingClient) GetFileNodes(ctx context.Context, fileKey string, nodeIDs []string, opts ...GetFileNodesOption) (*FileNodesResponse, error) {
	options := ApplyGetFileNodesOptions(opts)
	start := time.Now()
	fields := []logging.Field{
		logging.String("file_key", fileKey),
		logging.Any("node_ids", nodeIDs),
		logging.String("branch", options.Branch),
	}
	if options.Depth != nil {
		fields = append(fields, logging.Int("depth", *options.Depth))
	}
	if options.Geometry != nil {
		fields = append(fields, logging.Bool("geometry", *options.Geometry))
	}
	lc.logger.Debug(ctx, "GetFileNodes request", fields...)

	response, err := lc.client.GetFileNodes(ctx, fileKey, nodeIDs, opts...)

	duration := time.Since(start)
	fields = append(fields,
		logging.Int("node_count", len(nodeIDs)),
		logging.Float64("duration_ms", duration.Seconds()*1000),
	)
	if err != nil {
		lc.logger.Error(ctx, "GetFileNodes failed", append(fields, logging.Err(err))...)
	} else {
		lc.logger.Debug(ctx, "GetFileNodes success", append(fields,
			logging.Int("result_count", len(response.Nodes)),
		)...)
	}
	return response, err
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
