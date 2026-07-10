package cli

import (
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
)

// ScopeComments keeps comments attached to selected nodes and, optionally,
// their descendants and ancestors. It also enriches comments with node paths.
func ScopeComments(client *figma.Client, fileID string, nodeIDs []string, comments []extract.CommentOutput, recursive, includeAncestors bool) ([]extract.CommentOutput, error) {
	documents, err := figma.FetchNodeDocuments(client, fileID, nodeIDs)
	if err != nil {
		return nil, err
	}
	extract.AttachCommentNodePaths(comments, extract.CommentNodePaths(documents))
	scopeIDs := extract.CommentNodeIDs(documents, recursive)
	if includeAncestors {
		fileDocument, err := figma.FetchDocument(client, fileID, nodeIDs, "", "")
		if err != nil {
			return nil, err
		}
		extract.AttachCommentNodePaths(comments, extract.CommentNodePaths([]any{fileDocument}))
		for _, nodeID := range nodeIDs {
			for ancestorID := range extract.AncestorNodeIDs(fileDocument, nodeID) {
				scopeIDs[ancestorID] = struct{}{}
			}
		}
	}
	return extract.FilterCommentsByNodeIDs(comments, scopeIDs), nil
}
