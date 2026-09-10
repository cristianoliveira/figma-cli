package figma

import "github.com/cristianoliveira/figma-cli/internal/figma/api"

// MapUser converts a generated api.User into the stable application-facing DTO.
func MapUser(u api.User) User {
	return User{
		Handle:   u.Handle,
		ID:       u.Id,
		ImageURL: u.ImgUrl,
	}
}

// MapProjectFiles converts the generated /v1/projects/:id/files response into
// the stable ProjectFiles DTO. Optional fields such as ThumbnailURL are
// preserved as pointers so callers can distinguish missing from empty.
func MapProjectFiles(raw api.GetProjectFilesResponse) ProjectFiles {
	files := make([]ProjectFile, 0, len(raw.Files))
	for _, file := range raw.Files {
		files = append(files, ProjectFile{
			Key:          file.Key,
			Name:         file.Name,
			LastModified: file.LastModified,
			ThumbnailURL: file.ThumbnailUrl,
		})
	}
	return ProjectFiles{Name: raw.Name, Files: files}
}

// MapTeamProjects converts the generated /v1/teams/:id/projects response into
// the stable TeamProjects DTO.
func MapTeamProjects(raw api.GetTeamProjectsResponse) TeamProjects {
	projects := make([]TeamProject, 0, len(raw.Projects))
	for _, project := range raw.Projects {
		projects = append(projects, TeamProject{
			ID:   project.Id,
			Name: project.Name,
		})
	}
	return TeamProjects{Name: raw.Name, Projects: projects}
}

// MapFileVersions converts the generated /v1/files/:key/versions response into
// the stable FileVersions DTO. Pagination cursors collapse into presence flags.
func MapFileVersions(raw api.GetFileVersionsResponse) FileVersions {
	versions := make([]FileVersion, 0, len(raw.Versions))
	for _, v := range raw.Versions {
		versions = append(versions, FileVersion{
			ID:           v.Id,
			CreatedAt:    v.CreatedAt,
			Label:        v.Label,
			Description:  v.Description,
			User:         MapUser(v.User),
			ThumbnailURL: v.ThumbnailUrl,
		})
	}
	return FileVersions{
		Versions: versions,
		Pagination: Pagination{
			HasNextPage: raw.Pagination.NextPage != nil,
			HasPrevPage: raw.Pagination.PrevPage != nil,
		},
	}
}

// MapComments converts the generated /v1/files/:key/comments payload into
// the stable Comment DTOs. Resolved collapses the optional ResolvedAt
// timestamp into a boolean so callers do not need to inspect generated
// nullable fields. ClientMeta carries the raw JSON of the original union so
// downstream packages can decode pin coordinates without depending on the
// generated union type.
func MapComments(raw []api.Comment) []Comment {
	out := make([]Comment, 0, len(raw))
	for _, comment := range raw {
		out = append(out, Comment{
			ID:         comment.Id,
			Message:    comment.Message,
			CreatedAt:  comment.CreatedAt,
			ParentID:   comment.ParentId,
			Resolved:   comment.ResolvedAt != nil,
			User:       MapUser(comment.User),
			ClientMeta: clientMetaJSON(comment.ClientMeta),
		})
	}
	return out
}

// clientMetaJSON returns the raw JSON of the generated client_meta union, or
// nil when the union is empty. Marshalling then re-encoding avoids forcing
// callers to depend on the generated Comment_ClientMeta wrapper.
func clientMetaJSON(meta api.Comment_ClientMeta) []byte {
	data, err := meta.MarshalJSON()
	if err != nil {
		return nil
	}
	// MarshalJSON on an empty json.RawMessage returns the literal "null";
	// collapse that to nil so callers can detect absence with a nil check.
	if string(data) == "null" {
		return nil
	}
	return data
}
