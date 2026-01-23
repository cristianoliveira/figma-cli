package api

import (
	"context"
)

// Client defines the interface for interacting with the Figma API.
type Client interface {
	// GetFile retrieves a Figma file by its key.
	GetFile(ctx context.Context, fileKey string) (*File, error)

	// GetNode retrieves a specific node within a file.
	GetNode(ctx context.Context, fileKey, nodeID string) (*Node, error)

	// GetNodes retrieves multiple nodes within a file.
	GetNodes(ctx context.Context, fileKey string, nodeIDs []string) (map[string]*Node, error)

	// GetImage retrieves an image representation of a node.
	GetImage(ctx context.Context, fileKey string, nodeIDs []string, options *ImageOptions) (map[string]string, error)

	// GetComments retrieves comments for a file.
	GetComments(ctx context.Context, fileKey string) ([]*Comment, error)

	// PostComment posts a comment to a file.
	PostComment(ctx context.Context, fileKey string, comment *CommentRequest) (*Comment, error)
}

// File represents a Figma file.
type File struct {
	Key          string          `json:"key"`
	Name         string          `json:"name"`
	LastModified string          `json:"lastModified"`
	ThumbnailURL string          `json:"thumbnailUrl"`
	Version      string          `json:"version"`
	Document     *Node           `json:"document"`
	Components   map[string]Node `json:"components"`
}

// Node represents a Figma node.
type Node struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Type     string           `json:"type"`
	Visible  bool             `json:"visible"`
	Children []Node           `json:"children,omitempty"`
	Styles   map[string]Style `json:"styles,omitempty"`
	// Additional fields can be added as needed.
}

// Style represents a Figma style.
type Style struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// ImageOptions defines options for image export.
type ImageOptions struct {
	Scale      float64 `json:"scale"`
	Format     string  `json:"format"`     // "png", "jpg", "svg", "pdf"
	Constraint string  `json:"constraint"` // "scale", "width", "height"
	Value      float64 `json:"value"`      // constraint value
}

// Comment represents a comment on a Figma file.
type Comment struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
	User      User   `json:"user"`
}

// User represents a Figma user.
type User struct {
	ID     string `json:"id"`
	Handle string `json:"handle"`
	ImgURL string `json:"imgUrl"`
}

// CommentRequest represents a request to post a comment.
type CommentRequest struct {
	Message string `json:"message"`
	// Optional: position data
}

// Common errors
var (
	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = &apiError{"not found"}
	// ErrInvalidRequest is returned when the request is malformed.
	ErrInvalidRequest = &apiError{"invalid request"}
)

type apiError struct {
	msg string
}

func (e *apiError) Error() string { return e.msg }
