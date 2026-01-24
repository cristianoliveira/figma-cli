package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/api"
)

// HTTPClient implements api.Client using a Transport and RequestBuilder.
type HTTPClient struct {
	transport      Transport
	requestBuilder RequestBuilder
}

// NewHTTPClient creates a new HTTPClient with the given transport and request builder.
func NewHTTPClient(transport Transport, requestBuilder RequestBuilder) *HTTPClient {
	return &HTTPClient{
		transport:      transport,
		requestBuilder: requestBuilder,
	}
}

// parseNodesResponse parses the response from the /v1/files/{file}/nodes endpoint.
// It supports both the nested document format (nodes[id].document) and direct node format.
func parseNodesResponse(body []byte) (map[string]*api.Node, error) {
	var response struct {
		Nodes map[string]json.RawMessage `json:"nodes"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, api.NewParseError(err, string(body), "decode nodes response")
	}

	nodes := make(map[string]*api.Node, len(response.Nodes))
	for id, raw := range response.Nodes {
		// First try to unmarshal as nested document wrapper
		var wrapper struct {
			Document *api.Node `json:"document"`
		}
		if err := json.Unmarshal(raw, &wrapper); err == nil && wrapper.Document != nil {
			nodes[id] = wrapper.Document
			continue
		}
		// Otherwise try to unmarshal as direct node
		var node api.Node
		if err := json.Unmarshal(raw, &node); err != nil {
			return nil, api.NewParseError(err, string(raw), "decode node")
		}
		nodes[id] = &node
	}
	return nodes, nil
}

// doGetRequest performs a GET request to the given path with optional branch query parameter.
func (c *HTTPClient) doGetRequest(ctx context.Context, path string, branch string) ([]byte, error) {
	if branch != "" {
		path = fmt.Sprintf("%s?branch_data=%s", path, url.QueryEscape(branch))
	}
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, body)
	}
	return body, nil
}

// doPostRequest performs a POST request to the given path with JSON body.
func (c *HTTPClient) doPostRequest(ctx context.Context, path string, body interface{}) ([]byte, error) {
	req, err := c.requestBuilder.Build(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, respBody)
	}
	return respBody, nil
}

// doDeleteRequest performs a DELETE request to the given path.
func (c *HTTPClient) doDeleteRequest(ctx context.Context, path string) error {
	req, err := c.requestBuilder.Build(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return api.ErrorFromResponse(resp, body)
	}
	return nil
}

// GetFile retrieves a Figma file by its key. If branch is not empty, fetches from the specified branch using the branch_data query parameter.
func (c *HTTPClient) GetFile(ctx context.Context, fileKey string, opts ...api.GetFileOption) (*api.File, error) {
	options := api.ApplyGetFileOptions(opts)
	path := fmt.Sprintf("/files/%s", fileKey)
	body, err := c.doGetRequest(ctx, path, options.Branch)
	if err != nil {
		return nil, err
	}

	var file api.File
	if err := json.Unmarshal(body, &file); err != nil {
		return nil, api.NewParseError(err, string(body), "decode file response")
	}
	return &file, nil
}

// GetNode retrieves a specific node within a file.
func (c *HTTPClient) GetNode(ctx context.Context, fileKey, nodeID string, opts ...api.GetNodeOption) (*api.Node, error) {
	options := api.ApplyGetNodeOptions(opts)

	// Build query parameters
	queryParams := make([]string, 0, 4)
	queryParams = append(queryParams, fmt.Sprintf("ids=%s", nodeID))
	if options.Branch != "" {
		queryParams = append(queryParams, fmt.Sprintf("branch_data=%s", url.QueryEscape(options.Branch)))
	}
	if options.Depth != nil {
		depth := *options.Depth
		if depth < 1 || depth > 10 {
			return nil, fmt.Errorf("depth must be between 1 and 10, got %d", depth)
		}
		queryParams = append(queryParams, fmt.Sprintf("depth=%d", depth))
	}
	if options.Geometry != nil && *options.Geometry {
		// geometry=true means minimal geometry (paths)
		queryParams = append(queryParams, "geometry=paths")
	}

	path := fmt.Sprintf("/files/%s/nodes", fileKey)
	if len(queryParams) > 0 {
		path = fmt.Sprintf("%s?%s", path, strings.Join(queryParams, "&"))
	}

	req, err := c.requestBuilder.Build(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, body)
	}

	nodes, err := parseNodesResponse(body)
	if err != nil {
		return nil, err
	}
	node, ok := nodes[nodeID]
	if !ok || node == nil {
		return nil, api.NewAPIError(http.StatusNotFound, "node not found", "")
	}
	return node, nil
}

// GetFileNodes retrieves multiple nodes within a file.
func (c *HTTPClient) GetFileNodes(ctx context.Context, fileKey string, nodeIDs []string, opts ...api.GetFileNodesOption) (*api.FileNodesResponse, error) {
	options := api.ApplyGetFileNodesOptions(opts)
	// Build query parameters
	var queryParams []string
	// node IDs (required)
	if len(nodeIDs) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("ids=%s", strings.Join(nodeIDs, ",")))
	}
	// depth (1-10)
	if options.Depth != nil {
		depth := *options.Depth
		if depth < 1 || depth > 10 {
			return nil, fmt.Errorf("depth must be between 1 and 10, got %d", depth)
		}
		queryParams = append(queryParams, fmt.Sprintf("depth=%d", depth))
	}
	// geometry (true for minimal geometry, false/nil for full)
	if options.Geometry != nil && *options.Geometry {
		queryParams = append(queryParams, "geometry=paths")
	}
	// branch
	if options.Branch != "" {
		queryParams = append(queryParams, fmt.Sprintf("branch_data=%s", url.QueryEscape(options.Branch)))
	}
	// Build path
	path := fmt.Sprintf("/files/%s/nodes", fileKey)
	if len(queryParams) > 0 {
		path = fmt.Sprintf("%s?%s", path, strings.Join(queryParams, "&"))
	}
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, body)
	}

	nodes, err := parseNodesResponse(body)
	if err != nil {
		return nil, err
	}
	return &api.FileNodesResponse{Nodes: nodes}, nil
}

// GetImage retrieves an image representation of a node.
func (c *HTTPClient) GetImage(ctx context.Context, fileKey string, nodeIDs []string, options *api.ImageOptions) (map[string]string, error) {
	// Build query parameters
	var queryParams []string
	if len(nodeIDs) > 0 {
		queryParams = append(queryParams, fmt.Sprintf("ids=%s", strings.Join(nodeIDs, ",")))
	}
	if options != nil {
		if options.Scale != 0 {
			queryParams = append(queryParams, fmt.Sprintf("scale=%g", options.Scale))
		}
		if options.Format != "" {
			queryParams = append(queryParams, fmt.Sprintf("format=%s", options.Format))
		}
		if options.Constraint != "" {
			queryParams = append(queryParams, fmt.Sprintf("constraint=%s", options.Constraint))
		}
		if options.Value != 0 {
			queryParams = append(queryParams, fmt.Sprintf("value=%g", options.Value))
		}
	}
	path := fmt.Sprintf("/images/%s", fileKey)
	if len(queryParams) > 0 {
		path = fmt.Sprintf("%s?%s", path, strings.Join(queryParams, "&"))
	}
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, body)
	}
	var response struct {
		Images map[string]string `json:"images"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, api.NewParseError(err, string(body), "decode image response")
	}
	return response.Images, nil
}

// GetComments retrieves comments for a file.
func (c *HTTPClient) GetComments(ctx context.Context, fileKey string) ([]*api.Comment, error) {
	path := fmt.Sprintf("/files/%s/comments", fileKey)
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, body)
	}
	var comments []*api.Comment
	if err := json.Unmarshal(body, &comments); err != nil {
		return nil, api.NewParseError(err, string(body), "decode comments response")
	}
	return comments, nil
}

// PostComment posts a comment to a file.
func (c *HTTPClient) PostComment(ctx context.Context, fileKey string, comment *api.CommentRequest) (*api.Comment, error) {
	path := fmt.Sprintf("/files/%s/comments", fileKey)
	body, err := c.doPostRequest(ctx, path, comment)
	if err != nil {
		return nil, err
	}
	var createdComment api.Comment
	if err := json.Unmarshal(body, &createdComment); err != nil {
		return nil, api.NewParseError(err, string(body), "decode comment response")
	}
	return &createdComment, nil
}

// DeleteComment deletes a comment from a file.
func (c *HTTPClient) DeleteComment(ctx context.Context, fileKey, commentID string) error {
	path := fmt.Sprintf("/files/%s/comments/%s", fileKey, commentID)
	return c.doDeleteRequest(ctx, path)
}

// GetCommentReactions retrieves reactions for a comment.
func (c *HTTPClient) GetCommentReactions(ctx context.Context, fileKey, commentID string) ([]*api.Reaction, error) {
	// TODO: implement
	return nil, fmt.Errorf("not implemented")
}

// PostCommentReaction adds a reaction to a comment.
func (c *HTTPClient) PostCommentReaction(ctx context.Context, fileKey, commentID, emoji string) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

// DeleteCommentReaction removes a reaction from a comment.
func (c *HTTPClient) DeleteCommentReaction(ctx context.Context, fileKey, commentID, emoji string) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

// GetFileMetadata retrieves metadata about a file.
func (c *HTTPClient) GetFileMetadata(ctx context.Context, fileKey string, opts ...api.GetFileMetadataOption) (*api.FileMeta, error) {
	options := api.ApplyGetFileMetadataOptions(opts)
	path := fmt.Sprintf("/files/%s/meta", fileKey)
	body, err := c.doGetRequest(ctx, path, options.Branch)
	if err != nil {
		return nil, err
	}

	var meta api.FileMeta
	if err := json.Unmarshal(body, &meta); err != nil {
		return nil, api.NewParseError(err, string(body), "decode file meta response")
	}
	return &meta, nil
}

// GetFileVersions retrieves version history of a file with pagination support.
func (c *HTTPClient) GetFileVersions(ctx context.Context, fileKey string, opts ...api.GetFileVersionsOption) ([]*api.Version, error) {
	options := api.ApplyGetFileVersionsOptions(opts)
	// Build path with query parameters
	path := fmt.Sprintf("/files/%s/versions", fileKey)
	var queryParams []string
	if options.PageSize > 0 {
		queryParams = append(queryParams, fmt.Sprintf("page_size=%d", options.PageSize))
	}
	if options.Before != "" {
		queryParams = append(queryParams, fmt.Sprintf("before=%s", options.Before))
	}
	if options.After != "" {
		queryParams = append(queryParams, fmt.Sprintf("after=%s", options.After))
	}
	if options.Branch != "" {
		queryParams = append(queryParams, fmt.Sprintf("branch_data=%s", url.QueryEscape(options.Branch)))
	}
	if len(queryParams) > 0 {
		path = fmt.Sprintf("%s?%s", path, strings.Join(queryParams, "&"))
	}

	req, err := c.requestBuilder.Build(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, body)
	}

	// Try to parse as a paginated response first
	type paginatedResponse struct {
		Versions   []*api.Version `json:"versions"`
		Pagination *struct {
			Before *string `json:"before"`
			After  *string `json:"after"`
		} `json:"pagination,omitempty"`
	}
	var paginated paginatedResponse
	if err := json.Unmarshal(body, &paginated); err == nil && paginated.Versions != nil {
		// Successfully parsed as paginated response
		// Pagination metadata is available in paginated.Pagination if needed
		return paginated.Versions, nil
	}

	// Fall back to direct array of versions
	var versions []*api.Version
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, api.NewParseError(err, string(body), "decode versions response")
	}
	return versions, nil
}

// GetTeams retrieves teams accessible to the authenticated user.
func (c *HTTPClient) GetTeams(ctx context.Context) ([]*api.Team, error) {
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, "/teams", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, body)
	}

	var teams []*api.Team
	if err := json.Unmarshal(body, &teams); err != nil {
		return nil, api.NewParseError(err, string(body), "decode teams response")
	}
	return teams, nil
}

// GetTeamStyles retrieves published styles for a team.
func (c *HTTPClient) GetTeamStyles(ctx context.Context, teamID string) ([]*api.Style, error) {
	path := fmt.Sprintf("/teams/%s/styles", teamID)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var styles []*api.Style
	if err := json.Unmarshal(body, &styles); err != nil {
		return nil, api.NewParseError(err, string(body), "decode team styles response")
	}
	return styles, nil
}

// GetStyle retrieves a specific style by its key.
func (c *HTTPClient) GetStyle(ctx context.Context, styleKey string) (*api.Style, error) {
	path := fmt.Sprintf("/styles/%s", styleKey)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var style api.Style
	if err := json.Unmarshal(body, &style); err != nil {
		return nil, api.NewParseError(err, string(body), "decode style response")
	}
	return &style, nil
}

// GetFileStyles retrieves styles defined in a file.
func (c *HTTPClient) GetFileStyles(ctx context.Context, fileKey string) ([]*api.Style, error) {
	path := fmt.Sprintf("/files/%s/styles", fileKey)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var styles []*api.Style
	if err := json.Unmarshal(body, &styles); err != nil {
		return nil, api.NewParseError(err, string(body), "decode file styles response")
	}
	return styles, nil
}

// GetTeamComponents retrieves published components for a team.
func (c *HTTPClient) GetTeamComponents(ctx context.Context, teamID string) ([]*api.Component, error) {
	path := fmt.Sprintf("/teams/%s/components", teamID)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var components []*api.Component
	if err := json.Unmarshal(body, &components); err != nil {
		return nil, api.NewParseError(err, string(body), "decode team components response")
	}
	return components, nil
}

// GetComponent retrieves a specific component by its key.
func (c *HTTPClient) GetComponent(ctx context.Context, componentKey string) (*api.Component, error) {
	path := fmt.Sprintf("/components/%s", componentKey)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var component api.Component
	if err := json.Unmarshal(body, &component); err != nil {
		return nil, api.NewParseError(err, string(body), "decode component response")
	}
	return &component, nil
}

// GetTeamComponentSets retrieves published component sets for a team.
func (c *HTTPClient) GetTeamComponentSets(ctx context.Context, teamID string) ([]*api.ComponentSet, error) {
	path := fmt.Sprintf("/teams/%s/component_sets", teamID)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var componentSets []*api.ComponentSet
	if err := json.Unmarshal(body, &componentSets); err != nil {
		return nil, api.NewParseError(err, string(body), "decode team component sets response")
	}
	return componentSets, nil
}

// GetComponentSet retrieves a specific component set by its key.
func (c *HTTPClient) GetComponentSet(ctx context.Context, componentSetKey string) (*api.ComponentSet, error) {
	path := fmt.Sprintf("/component_sets/%s", componentSetKey)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var componentSet api.ComponentSet
	if err := json.Unmarshal(body, &componentSet); err != nil {
		return nil, api.NewParseError(err, string(body), "decode component set response")
	}
	return &componentSet, nil
}

// GetFileComponents retrieves components defined in a file.
func (c *HTTPClient) GetFileComponents(ctx context.Context, fileKey string) ([]*api.Component, error) {
	path := fmt.Sprintf("/files/%s/components", fileKey)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var components []*api.Component
	if err := json.Unmarshal(body, &components); err != nil {
		return nil, api.NewParseError(err, string(body), "decode file components response")
	}
	return components, nil
}

// GetTeamProjects retrieves projects for a team.
func (c *HTTPClient) GetTeamProjects(ctx context.Context, teamID string) ([]*api.Project, error) {
	path := fmt.Sprintf("/teams/%s/projects", teamID)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}

	var projects []*api.Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, api.NewParseError(err, string(body), "decode projects response")
	}
	return projects, nil
}

// GetProjectFiles retrieves files in a project.
func (c *HTTPClient) GetProjectFiles(ctx context.Context, projectID string) ([]*api.File, error) {
	path := fmt.Sprintf("/projects/%s/files", projectID)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}

	var files []*api.File
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, api.NewParseError(err, string(body), "decode project files response")
	}
	return files, nil
}

// GetMe retrieves the current authenticated user.
func (c *HTTPClient) GetMe(ctx context.Context) (*api.User, error) {
	req, err := c.requestBuilder.Build(ctx, http.MethodGet, "/me", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.transport.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, api.ErrorFromResponse(resp, body)
	}

	var user api.User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, api.NewParseError(err, string(body), "decode user response")
	}
	return &user, nil
}

// GetFileVariables retrieves local variables in a file (Enterprise).
func (c *HTTPClient) GetFileVariables(ctx context.Context, fileKey string) ([]*api.Variable, error) {
	path := fmt.Sprintf("/files/%s/variables", fileKey)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var variables []*api.Variable
	if err := json.Unmarshal(body, &variables); err != nil {
		return nil, api.NewParseError(err, string(body), "decode file variables response")
	}
	return variables, nil
}

// GetTeamPublishedVariables retrieves published variables for a team (Enterprise).
func (c *HTTPClient) GetTeamPublishedVariables(ctx context.Context, teamID string) ([]*api.Variable, error) {
	path := fmt.Sprintf("/teams/%s/variables", teamID)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var variables []*api.Variable
	if err := json.Unmarshal(body, &variables); err != nil {
		return nil, api.NewParseError(err, string(body), "decode team published variables response")
	}
	return variables, nil
}

// PostFileVariables bulk creates/updates/deletes variables in a file (Enterprise).
func (c *HTTPClient) PostFileVariables(ctx context.Context, fileKey string, variables []*api.Variable) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

// GetFileDevResources retrieves developer resources for a file.
func (c *HTTPClient) GetFileDevResources(ctx context.Context, fileKey string) ([]*api.DevResource, error) {
	path := fmt.Sprintf("/files/%s/dev_resources", fileKey)
	body, err := c.doGetRequest(ctx, path, "")
	if err != nil {
		return nil, err
	}
	var devResources []*api.DevResource
	if err := json.Unmarshal(body, &devResources); err != nil {
		return nil, api.NewParseError(err, string(body), "decode dev resources response")
	}
	return devResources, nil
}

// PostDevResources bulk creates developer resources.
func (c *HTTPClient) PostDevResources(ctx context.Context, resources []*api.DevResource) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

// PatchDevResources bulk updates developer resources.
func (c *HTTPClient) PatchDevResources(ctx context.Context, resources []*api.DevResource) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

// GetWebhooks retrieves configured webhooks (v2).
func (c *HTTPClient) GetWebhooks(ctx context.Context) ([]*api.Webhook, error) {
	// TODO: implement
	return nil, fmt.Errorf("not implemented")
}

// CreateWebhook creates a new webhook (v2).
func (c *HTTPClient) CreateWebhook(ctx context.Context, webhook *api.WebhookRequest) (*api.Webhook, error) {
	// TODO: implement
	return nil, fmt.Errorf("not implemented")
}

// DeleteWebhook deletes a webhook (v2).
func (c *HTTPClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

// UpdateWebhook updates an existing webhook (v2).
func (c *HTTPClient) UpdateWebhook(ctx context.Context, webhookID string, updates *api.WebhookUpdate) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}
