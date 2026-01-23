package api

// FileMeta represents metadata about a Figma file.
type FileMeta struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	LastModified string `json:"lastModified"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Version      string `json:"version"`
	// Add other fields as needed
}

// Version represents a version in a file's version history.
type Version struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"createdAt"`
	Label       string `json:"label"`
	Description string `json:"description"`
	User        User   `json:"user"`
}

// Reaction represents a reaction to a comment.
type Reaction struct {
	Emoji string `json:"emoji"`
	User  User   `json:"user"`
}

// Component represents a published component.
type Component struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	User            User   `json:"user"`
	ContainingFrame *Node  `json:"containingFrame,omitempty"`
	// Additional fields can be added
}

// ComponentSet represents a collection of component variants.
type ComponentSet struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	User        User   `json:"user"`
}

// Project represents a Figma project.
type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"createdAt"`
	ModifiedAt string `json:"modifiedAt"`
}

// Variable represents a design variable (Enterprise feature).
type Variable struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Key                  string `json:"key"`
	VariableCollectionID string `json:"variableCollectionId"`
	// Additional fields as needed
}

// DevResource represents a developer resource.
type DevResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Additional fields as needed
}

// Webhook represents a webhook configuration.
type Webhook struct {
	ID        string `json:"id"`
	EventType string `json:"eventType"`
	URL       string `json:"url"`
	Status    string `json:"status"`
	// Additional fields as needed
}

// WebhookRequest represents a request to create a webhook.
type WebhookRequest struct {
	EventType string `json:"eventType"`
	URL       string `json:"url"`
	// Additional fields as needed
}

// WebhookUpdate represents updates to a webhook.
type WebhookUpdate struct {
	Status string `json:"status,omitempty"`
	// Additional fields as needed
}

// ReactionRequest represents a request to add a reaction.
type ReactionRequest struct {
	Emoji string `json:"emoji"`
}
