// Package comments contains comment retrieval and node-scoping workflows.
package comments

import (
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// Fetch retrieves a file's comments and converts them to command output.
func Fetch(client *figma.Client, fileID string) ([]extract.CommentOutput, error) {
	var response api.GetCommentsResponse
	if err := client.Fetch(figma.BuildCommentsURL(fileID, ""), &response); err != nil {
		return nil, err
	}
	outputs := extract.CommentOutputs(response.Comments)
	for index := range outputs {
		outputs[index].URL = figma.BuildCommentWebURL(fileID, outputs[index].NodeID, outputs[index].ID)
	}
	return outputs, nil
}

// Scope keeps comments attached to selected nodes and, optionally, their
// descendants and ancestors. It enriches the returned comments with node paths.
func Scope(client *figma.Client, fileID string, nodeIDs []string, comments []extract.CommentOutput, recursive, includeAncestors bool) ([]extract.CommentOutput, error) {
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
