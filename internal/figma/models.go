package figma

import (
	"encoding/json"
	"time"
)

// User is the stable, application-facing representation of a Figma user.
// The generated API types use snake_case field names such as img_url;
// this DTO shields callers from upstream field renames.
type User struct {
	Handle   string
	ID       string
	ImageURL string
}

// Pagination captures the presence of pagination cursors as a boolean
// signal. The Figma API returns opaque next_page/prev_page URLs that we
// translate into cursor IDs in the command layer.
type Pagination struct {
	HasNextPage bool
	HasPrevPage bool
}

// ProjectFile is a stable file descriptor returned by /v1/projects/:id/files.
type ProjectFile struct {
	Key          string
	Name         string
	LastModified time.Time
	ThumbnailURL *string
}

// ProjectFiles is the result of listing files in a Figma project.
type ProjectFiles struct {
	Name  string
	Files []ProjectFile
}

// TeamProject is a stable project entry returned by /v1/teams/:id/projects.
type TeamProject struct {
	ID   string
	Name string
}

// TeamProjects is the result of listing projects in a Figma team.
type TeamProjects struct {
	Name     string
	Projects []TeamProject
}

// FileVersion is a stable version descriptor returned by /v1/files/:key/versions.
type FileVersion struct {
	ID           string
	CreatedAt    time.Time
	Label        *string
	Description  *string
	User         User
	ThumbnailURL *string
}

// FileVersions is the paginated result of fetching version history.
type FileVersions struct {
	Versions   []FileVersion
	Pagination Pagination
}

// Comment is a stable comment descriptor returned by /v1/files/:key/comments.
// ClientMeta holds the raw JSON payload of the original client_meta union so
// callers can decode pin coordinates without depending on the generated
// union type, which is rewritten whenever OpenAPI changes.
type Comment struct {
	ID         string
	Message    string
	CreatedAt  time.Time
	ParentID   *string
	Resolved   bool
	User       User
	ClientMeta json.RawMessage
}
