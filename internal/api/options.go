package api

// GetFileOptions holds optional parameters for GetFile.
type GetFileOptions struct {
	Branch string
}

// GetFileOption configures GetFileOptions.
type GetFileOption func(*GetFileOptions)

// WithBranch sets the branch parameter for file operations.
func WithBranch(branch string) GetFileOption {
	return func(opts *GetFileOptions) {
		opts.Branch = branch
	}
}

// ApplyGetFileOptions applies options to GetFileOptions.
func ApplyGetFileOptions(opts []GetFileOption) GetFileOptions {
	config := GetFileOptions{}
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

// GetFileMetadataOptions holds optional parameters for GetFileMetadata.
type GetFileMetadataOptions struct {
	Branch string
}

// GetFileMetadataOption configures GetFileMetadataOptions.
type GetFileMetadataOption func(*GetFileMetadataOptions)

// WithBranchForMetadata sets the branch parameter for metadata operations.
func WithBranchForMetadata(branch string) GetFileMetadataOption {
	return func(opts *GetFileMetadataOptions) {
		opts.Branch = branch
	}
}

// ApplyGetFileMetadataOptions applies options to GetFileMetadataOptions.
func ApplyGetFileMetadataOptions(opts []GetFileMetadataOption) GetFileMetadataOptions {
	config := GetFileMetadataOptions{}
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

// GetFileVersionsOptions holds optional parameters for GetFileVersions.
type GetFileVersionsOptions struct {
	PageSize int
	Before   string
	After    string
	Branch   string
}

// GetFileVersionsOption configures GetFileVersionsOptions.
type GetFileVersionsOption func(*GetFileVersionsOptions)

// WithPageSize sets the page size for pagination.
func WithPageSize(pageSize int) GetFileVersionsOption {
	return func(opts *GetFileVersionsOptions) {
		opts.PageSize = pageSize
	}
}

// WithBefore sets the before cursor for pagination.
func WithBefore(before string) GetFileVersionsOption {
	return func(opts *GetFileVersionsOptions) {
		opts.Before = before
	}
}

// WithAfter sets the after cursor for pagination.
func WithAfter(after string) GetFileVersionsOption {
	return func(opts *GetFileVersionsOptions) {
		opts.After = after
	}
}

// WithBranchForVersions sets the branch parameter for version operations.
func WithBranchForVersions(branch string) GetFileVersionsOption {
	return func(opts *GetFileVersionsOptions) {
		opts.Branch = branch
	}
}

// ApplyGetFileVersionsOptions applies options to GetFileVersionsOptions.
func ApplyGetFileVersionsOptions(opts []GetFileVersionsOption) GetFileVersionsOptions {
	config := GetFileVersionsOptions{}
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

// GetFileNodesOptions holds optional parameters for GetFileNodes.
type GetFileNodesOptions struct {
	Branch string
}

// GetFileNodesOption configures GetFileNodesOptions.
type GetFileNodesOption func(*GetFileNodesOptions)

// WithBranchForNodes sets the branch parameter for node operations.
func WithBranchForNodes(branch string) GetFileNodesOption {
	return func(opts *GetFileNodesOptions) {
		opts.Branch = branch
	}
}

// ApplyGetFileNodesOptions applies options to GetFileNodesOptions.
func ApplyGetFileNodesOptions(opts []GetFileNodesOption) GetFileNodesOptions {
	config := GetFileNodesOptions{}
	for _, opt := range opts {
		opt(&config)
	}
	return config
}

// FileNodesResponse represents the response from GetFileNodes.
type FileNodesResponse struct {
	Nodes map[string]*Node `json:"nodes"`
}
