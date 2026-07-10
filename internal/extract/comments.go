package extract

import (
	"sort"
	"strings"
)

// CommentOutput is one comment, used by `figma comments`.
type CommentOutput struct {
	ID        string   `json:"id"`
	Message   string   `json:"message"`
	CreatedAt string   `json:"created_at"`
	Resolved  bool     `json:"resolved"`
	NodeID    string   `json:"node_id,omitempty"`
	NodePath  []string `json:"node_path,omitempty"`
	User      string   `json:"user"`
	ParentID  string   `json:"parent_id,omitempty"`
	URL       string   `json:"url,omitempty"`
}

// CommentThreadOutput groups one review thread with replies in creation order.
type CommentThreadOutput struct {
	Root    CommentOutput   `json:"root"`
	Replies []CommentOutput `json:"replies"`
}

// GroupCommentThreads groups flat API comments and retains replies whose parents were deleted.
func GroupCommentThreads(comments []CommentOutput) []CommentThreadOutput {
	byID := make(map[string]CommentOutput, len(comments))
	for _, comment := range comments {
		byID[comment.ID] = comment
	}
	threadsByRoot := make(map[string]*CommentThreadOutput)
	for _, comment := range comments {
		rootID := commentRootID(comment, byID)
		thread := threadsByRoot[rootID]
		if thread == nil {
			root := byID[rootID]
			thread = &CommentThreadOutput{Root: root, Replies: make([]CommentOutput, 0)}
			threadsByRoot[rootID] = thread
		}
		if comment.ID != rootID {
			thread.Replies = append(thread.Replies, comment)
		}
	}
	threads := make([]CommentThreadOutput, 0, len(threadsByRoot))
	for _, thread := range threadsByRoot {
		sort.SliceStable(thread.Replies, func(i, j int) bool {
			return thread.Replies[i].CreatedAt < thread.Replies[j].CreatedAt
		})
		threads = append(threads, *thread)
	}
	sort.SliceStable(threads, func(i, j int) bool {
		return threads[i].Root.CreatedAt < threads[j].Root.CreatedAt
	})
	return threads
}

func commentRootID(comment CommentOutput, byID map[string]CommentOutput) string {
	current := comment
	visited := map[string]struct{}{current.ID: {}}
	for current.ParentID != "" {
		parent, ok := byID[current.ParentID]
		if !ok {
			return comment.ID
		}
		if _, seen := visited[parent.ID]; seen {
			return comment.ID
		}
		visited[parent.ID] = struct{}{}
		current = parent
	}
	return current.ID
}

// FilterCommentThreads applies review filters while retaining complete matching threads.
func FilterCommentThreads(threads []CommentThreadOutput, state, author, after, before string) []CommentThreadOutput {
	filtered := make([]CommentThreadOutput, 0, len(threads))
	for _, thread := range threads {
		if state == "open" && thread.Root.Resolved || state == "resolved" && !thread.Root.Resolved {
			continue
		}
		if author != "" && !threadHasAuthor(thread, author) {
			continue
		}
		if after != "" && thread.Root.CreatedAt < after || before != "" && thread.Root.CreatedAt > before {
			continue
		}
		filtered = append(filtered, thread)
	}
	return filtered
}

func threadHasAuthor(thread CommentThreadOutput, author string) bool {
	wanted := strings.ToLower(author)
	if strings.Contains(strings.ToLower(thread.Root.User), wanted) {
		return true
	}
	for _, reply := range thread.Replies {
		if strings.Contains(strings.ToLower(reply.User), wanted) {
			return true
		}
	}
	return false
}

// CommentNodePaths maps node IDs to readable name paths within selected subtrees.
func CommentNodePaths(documents []any) map[string][]string {
	paths := make(map[string][]string)
	for _, document := range documents {
		collectCommentNodePaths(document, nil, paths)
	}
	return paths
}

func collectCommentNodePaths(value any, parentPath []string, paths map[string][]string) {
	node, ok := value.(map[string]any)
	if !ok {
		return
	}
	path := append([]string(nil), parentPath...)
	if name := StringValue(node["name"]); name != "" {
		path = append(path, name)
	}
	if id := StringValue(node["id"]); id != "" {
		paths[id] = path
	}
	children, _ := node["children"].([]any)
	for _, child := range children {
		collectCommentNodePaths(child, path, paths)
	}
}

// AttachCommentNodePaths enriches anchored comments with selected-tree context.
func AttachCommentNodePaths(comments []CommentOutput, paths map[string][]string) {
	for index := range comments {
		if path := paths[comments[index].NodeID]; len(path) > 0 {
			comments[index].NodePath = append([]string(nil), path...)
		}
	}
}

// CommentNodeIDs returns IDs eligible for node-scoped comments.
func CommentNodeIDs(documents []any, recursive bool) map[string]struct{} {
	ids := make(map[string]struct{})
	for _, document := range documents {
		collectCommentNodeIDs(document, recursive, ids)
	}
	return ids
}

func collectCommentNodeIDs(value any, recursive bool, ids map[string]struct{}) {
	node, ok := value.(map[string]any)
	if !ok {
		return
	}
	if id, ok := node["id"].(string); ok && id != "" {
		ids[id] = struct{}{}
	}
	if !recursive {
		return
	}
	children, _ := node["children"].([]any)
	for _, child := range children {
		collectCommentNodeIDs(child, true, ids)
	}
}

// FilterCommentsByNodeIDs keeps comments anchored to selected nodes and every
// reply in those comment threads while preserving API order.
func FilterCommentsByNodeIDs(comments []CommentOutput, nodeIDs map[string]struct{}) []CommentOutput {
	included := make(map[string]struct{})
	for _, comment := range comments {
		if _, ok := nodeIDs[comment.NodeID]; ok {
			included[comment.ID] = struct{}{}
		}
	}
	for changed := true; changed; {
		changed = false
		for _, comment := range comments {
			if _, ok := included[comment.ParentID]; !ok {
				continue
			}
			if _, ok := included[comment.ID]; ok {
				continue
			}
			included[comment.ID] = struct{}{}
			changed = true
		}
	}

	filtered := make([]CommentOutput, 0)
	for _, comment := range comments {
		if _, ok := included[comment.ID]; ok {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}

// FilterCommentsByID returns the exact comment matching id.
func FilterCommentsByID(comments []CommentOutput, id string) []CommentOutput {
	for _, comment := range comments {
		if comment.ID == id {
			return []CommentOutput{comment}
		}
	}
	return nil
}

// FilterUnresolvedComments removes resolved comments while preserving order.
func FilterUnresolvedComments(comments []CommentOutput) []CommentOutput {
	filtered := make([]CommentOutput, 0, len(comments))
	for _, comment := range comments {
		if !comment.Resolved {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}

// AncestorNodeIDs returns target and its ancestor path from a full document.
func AncestorNodeIDs(document any, targetID string) map[string]struct{} {
	var path []string
	if !findNodePath(document, targetID, &path) {
		return nil
	}
	ids := make(map[string]struct{}, len(path))
	for _, id := range path {
		ids[id] = struct{}{}
	}
	return ids
}

func findNodePath(value any, targetID string, path *[]string) bool {
	node, ok := value.(map[string]any)
	if !ok {
		return false
	}
	id := StringValue(node["id"])
	if id != "" {
		*path = append(*path, id)
	}
	if id == targetID {
		return true
	}
	children, _ := node["children"].([]any)
	for _, child := range children {
		if findNodePath(child, targetID, path) {
			return true
		}
	}
	if id != "" {
		*path = (*path)[:len(*path)-1]
	}
	return false
}
