package extract

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommentNodePathsIncludesReadableHierarchy(t *testing.T) {
	document := map[string]any{"id": "1:1", "name": "Checkout", "children": []any{
		map[string]any{"id": "1:2", "name": "Payment", "children": []any{
			map[string]any{"id": "1:3", "name": "Pay button"},
		}},
	}}

	paths := CommentNodePaths([]any{document})

	assert.Equal(t, []string{"Checkout", "Payment", "Pay button"}, paths["1:3"])
}

func TestAttachCommentNodePaths(t *testing.T) {
	comments := []CommentOutput{{ID: "root", NodeID: "1:2"}, {ID: "reply", ParentID: "root"}}

	AttachCommentNodePaths(comments, map[string][]string{"1:2": {"Screen", "Button"}})

	assert.Equal(t, []string{"Screen", "Button"}, comments[0].NodePath)
	assert.Empty(t, comments[1].NodePath)
}

func TestCommentNodeIDs_RecursiveControlsDescendants(t *testing.T) {
	document := map[string]any{
		"id": "1:1",
		"children": []any{
			map[string]any{"id": "1:2"},
		},
	}

	assert.Equal(t, map[string]struct{}{"1:1": {}}, CommentNodeIDs([]any{document}, false))
	assert.Equal(t, map[string]struct{}{"1:1": {}, "1:2": {}}, CommentNodeIDs([]any{document}, true))
}

func TestFilterCommentsByNodeIDs_IncludesThreadReplies(t *testing.T) {
	comments := []CommentOutput{
		{ID: "reply", ParentID: "root"},
		{ID: "other", NodeID: "9:9"},
		{ID: "root", NodeID: "1:2"},
		{ID: "nested-reply", ParentID: "reply"},
	}

	filtered := FilterCommentsByNodeIDs(comments, map[string]struct{}{"1:2": {}})

	assert.Equal(t, []string{"reply", "root", "nested-reply"}, []string{filtered[0].ID, filtered[1].ID, filtered[2].ID})
}

func TestFilterCommentsByNodeIDs_NoMatches(t *testing.T) {
	filtered := FilterCommentsByNodeIDs([]CommentOutput{{ID: "other", NodeID: "9:9"}}, map[string]struct{}{"1:2": {}})

	assert.Empty(t, filtered)
}

func TestFilterCommentsByID(t *testing.T) {
	comments := []CommentOutput{{ID: "other"}, {ID: "1838610593", Message: "target"}}

	assert.Equal(t, []CommentOutput{{ID: "1838610593", Message: "target"}}, FilterCommentsByID(comments, "1838610593"))
}

func TestGroupCommentThreadsOrdersRootsAndNestedReplies(t *testing.T) {
	comments := []CommentOutput{
		{ID: "reply-2", ParentID: "reply-1", CreatedAt: "2026-01-03T00:00:00Z"},
		{ID: "root-2", CreatedAt: "2026-01-04T00:00:00Z", Resolved: true},
		{ID: "reply-1", ParentID: "root-1", CreatedAt: "2026-01-02T00:00:00Z"},
		{ID: "root-1", CreatedAt: "2026-01-01T00:00:00Z"},
	}

	threads := GroupCommentThreads(comments)

	require.Len(t, threads, 2)
	assert.Equal(t, "root-1", threads[0].Root.ID)
	assert.Equal(t, []string{"reply-1", "reply-2"}, []string{threads[0].Replies[0].ID, threads[0].Replies[1].ID})
	assert.Equal(t, "root-2", threads[1].Root.ID)
}

func TestGroupCommentThreadsKeepsOrphanedReply(t *testing.T) {
	threads := GroupCommentThreads([]CommentOutput{{ID: "orphan", ParentID: "deleted"}})

	require.Len(t, threads, 1)
	assert.Equal(t, "orphan", threads[0].Root.ID)
}

func TestFilterCommentThreadsByStateAndAuthor(t *testing.T) {
	threads := []CommentThreadOutput{
		{Root: CommentOutput{ID: "open", User: "Ada"}, Replies: []CommentOutput{{User: "Cristian"}}},
		{Root: CommentOutput{ID: "done", User: "Linus", Resolved: true}},
	}

	assert.Equal(t, []CommentThreadOutput{threads[0]}, FilterCommentThreads(threads, "open", "cristian", "", ""))
	assert.Equal(t, []CommentThreadOutput{threads[1]}, FilterCommentThreads(threads, "resolved", "", "", ""))
}

func TestFilterCommentThreadsByCreationRange(t *testing.T) {
	threads := []CommentThreadOutput{
		{Root: CommentOutput{ID: "old", CreatedAt: "2026-01-01T00:00:00Z"}},
		{Root: CommentOutput{ID: "current", CreatedAt: "2026-01-10T00:00:00Z"}},
		{Root: CommentOutput{ID: "future", CreatedAt: "2026-02-01T00:00:00Z"}},
	}

	assert.Equal(t, []CommentThreadOutput{threads[1]}, FilterCommentThreads(
		threads, "all", "", "2026-01-05T00:00:00Z", "2026-01-31T00:00:00Z",
	))
}

func TestAncestorNodeIDsIncludesTargetAndParents(t *testing.T) {
	document := map[string]any{"id": "0:0", "children": []any{
		map[string]any{"id": "1:1", "children": []any{
			map[string]any{"id": "1:2"},
		}},
	}}

	assert.Equal(t, map[string]struct{}{"0:0": {}, "1:1": {}, "1:2": {}}, AncestorNodeIDs(document, "1:2"))
	assert.Empty(t, AncestorNodeIDs(document, "missing"))
}
