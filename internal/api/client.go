package api

import (
	"context"
)

// Client defines the interface for interacting with the Figma API.
type Client interface {
	// GetFile retrieves a Figma file by its key. If branch is not empty, fetches from the specified branch.
	GetFile(ctx context.Context, fileKey string, opts ...GetFileOption) (*File, error)

	// GetNode retrieves a specific node within a file.
	GetNode(ctx context.Context, fileKey, nodeID string, opts ...GetNodeOption) (*Node, error)

	// GetFileNodes retrieves multiple nodes within a file.
	GetFileNodes(ctx context.Context, fileKey string, nodeIDs []string, opts ...GetFileNodesOption) (*FileNodesResponse, error)

	// GetImage retrieves an image representation of a node.
	GetImage(ctx context.Context, fileKey string, nodeIDs []string, options *ImageOptions) (map[string]string, error)

	// GetComments retrieves comments for a file.
	GetComments(ctx context.Context, fileKey string) ([]*Comment, error)

	// PostComment posts a comment to a file.
	PostComment(ctx context.Context, fileKey string, comment *CommentRequest) (*Comment, error)

	// DeleteComment deletes a comment from a file.
	DeleteComment(ctx context.Context, fileKey, commentID string) error

	// GetCommentReactions retrieves reactions for a comment.
	GetCommentReactions(ctx context.Context, fileKey, commentID string) ([]*Reaction, error)

	// PostCommentReaction adds a reaction to a comment.
	PostCommentReaction(ctx context.Context, fileKey, commentID, emoji string) error

	// DeleteCommentReaction removes a reaction from a comment.
	DeleteCommentReaction(ctx context.Context, fileKey, commentID, emoji string) error

	// GetFileMetadata retrieves metadata about a file.
	GetFileMetadata(ctx context.Context, fileKey string, opts ...GetFileMetadataOption) (*FileMeta, error)

	// GetFileVersions retrieves version history of a file with pagination support.
	GetFileVersions(ctx context.Context, fileKey string, opts ...GetFileVersionsOption) ([]*Version, error)

	// GetTeamStyles retrieves published styles for a team.
	GetTeamStyles(ctx context.Context, teamID string) ([]*Style, error)

	// GetStyle retrieves a specific style by its key.
	GetStyle(ctx context.Context, styleKey string) (*Style, error)

	// GetFileStyles retrieves styles defined in a file.
	GetFileStyles(ctx context.Context, fileKey string) ([]*Style, error)

	// GetTeamComponents retrieves published components for a team.
	GetTeamComponents(ctx context.Context, teamID string) ([]*Component, error)

	// GetComponent retrieves a specific component by its key.
	GetComponent(ctx context.Context, componentKey string) (*Component, error)

	// GetTeamComponentSets retrieves published component sets for a team.
	GetTeamComponentSets(ctx context.Context, teamID string) ([]*ComponentSet, error)

	// GetComponentSet retrieves a specific component set by its key.
	GetComponentSet(ctx context.Context, componentSetKey string) (*ComponentSet, error)

	// GetFileComponents retrieves components defined in a file.
	GetFileComponents(ctx context.Context, fileKey string) ([]*Component, error)

	// GetTeamProjects retrieves projects for a team.
	GetTeamProjects(ctx context.Context, teamID string) ([]*Project, error)

	// GetProjectFiles retrieves files in a project.
	GetProjectFiles(ctx context.Context, projectID string) ([]*File, error)

	// GetMe retrieves the current authenticated user.
	GetMe(ctx context.Context) (*User, error)

	// GetFileVariables retrieves local variables in a file (Enterprise).
	GetFileVariables(ctx context.Context, fileKey string) ([]*Variable, error)

	// GetTeamPublishedVariables retrieves published variables for a team (Enterprise).
	GetTeamPublishedVariables(ctx context.Context, teamID string) ([]*Variable, error)

	// PostFileVariables bulk creates/updates/deletes variables in a file (Enterprise).
	PostFileVariables(ctx context.Context, fileKey string, variables []*Variable) error

	// GetFileDevResources retrieves developer resources for a file.
	GetFileDevResources(ctx context.Context, fileKey string) ([]*DevResource, error)

	// PostDevResources bulk creates developer resources.
	PostDevResources(ctx context.Context, resources []*DevResource) error

	// PatchDevResources bulk updates developer resources.
	PatchDevResources(ctx context.Context, resources []*DevResource) error

	// GetWebhooks retrieves configured webhooks (v2).
	GetWebhooks(ctx context.Context) ([]*Webhook, error)

	// CreateWebhook creates a new webhook (v2).
	CreateWebhook(ctx context.Context, webhook *WebhookRequest) (*Webhook, error)

	// DeleteWebhook deletes a webhook (v2).
	DeleteWebhook(ctx context.Context, webhookID string) error

	// UpdateWebhook updates an existing webhook (v2).
	UpdateWebhook(ctx context.Context, webhookID string, updates *WebhookUpdate) error
}

// File represents a Figma file.
type File struct {
	Key           string                  `json:"key"`
	Name          string                  `json:"name"`
	LastModified  string                  `json:"lastModified"`
	ThumbnailURL  string                  `json:"thumbnailUrl"`
	Version       string                  `json:"version"`
	Document      *Node                   `json:"document"`
	Components    map[string]Component    `json:"components"`
	ComponentSets map[string]ComponentSet `json:"componentSets,omitempty"`
	Styles        map[string]Style        `json:"styles,omitempty"`
	// TODO: add schemaVersion, role, linkAccess, etc.
}

// Node represents a Figma node.
type Node struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Visible  bool              `json:"visible"`
	Children []Node            `json:"children,omitempty"`
	Styles   map[string]string `json:"styles,omitempty"`
	// Additional fields can be added as needed.
	AbsoluteBoundingBox         *Rectangle        `json:"absoluteBoundingBox,omitempty"`
	RelativeTransform           Transform         `json:"relativeTransform,omitempty"`
	IsFixed                     bool              `json:"isFixed,omitempty"`
	ScrollBehavior              string            `json:"scrollBehavior,omitempty"`
	Rotation                    float64           `json:"rotation,omitempty"`
	ComponentPropertyReferences map[string]string `json:"componentPropertyReferences,omitempty"`
	PluginData                  interface{}       `json:"pluginData,omitempty"`
	SharedPluginData            interface{}       `json:"sharedPluginData,omitempty"`
	ExplicitVariableModes       map[string]string `json:"explicitVariableModes,omitempty"`
	Characters                  string            `json:"characters,omitempty"`
	Style                       *TextStyle        `json:"style,omitempty"`
	Fills                       []Paint           `json:"fills,omitempty"`
	Strokes                     []Paint           `json:"strokes,omitempty"`
	Effects                     []Effect          `json:"effects,omitempty"`
	LayoutGrids                 []LayoutGrid      `json:"layoutGrids,omitempty"`
	StrokeWeight                float64           `json:"strokeWeight,omitempty"`
	StrokeAlign                 string            `json:"strokeAlign,omitempty"`
	CornerRadius                float64           `json:"cornerRadius,omitempty"`
	RectangleCornerRadii        []float64         `json:"rectangleCornerRadii,omitempty"`
	Constraints                 *LayoutConstraint `json:"constraints,omitempty"`
	Opacity                     float64           `json:"opacity,omitempty"`
	BlendMode                   string            `json:"blendMode,omitempty"`
	IsMask                      bool              `json:"isMask,omitempty"`
	Locked                      bool              `json:"locked,omitempty"`
	LayoutMode                  string            `json:"layoutMode,omitempty"`
	PrimaryAxisAlignItems       string            `json:"primaryAxisAlignItems,omitempty"`
	CounterAxisAlignItems       string            `json:"counterAxisAlignItems,omitempty"`
	ItemSpacing                 float64           `json:"itemSpacing,omitempty"`
	PaddingLeft                 float64           `json:"paddingLeft,omitempty"`
	PaddingRight                float64           `json:"paddingRight,omitempty"`
	PaddingTop                  float64           `json:"paddingTop,omitempty"`
	PaddingBottom               float64           `json:"paddingBottom,omitempty"`
	TransitionNodeID            string            `json:"transitionNodeID,omitempty"`
	TransitionDuration          float64           `json:"transitionDuration,omitempty"`
	TransitionEasing            string            `json:"transitionEasing,omitempty"`
	// TODO: add effects, exportSettings, background color
}

// Style represents a Figma style.
type Style struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Remote      bool   `json:"remote"`
	StyleType   string `json:"styleType"` // "FILL", "TEXT", "EFFECT", "GRID"
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
