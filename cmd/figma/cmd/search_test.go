package cmd

import (
	"regexp"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/api"
)

func TestTextMatches(t *testing.T) {
	tests := []struct {
		text          string
		query         string
		caseSensitive bool
		exact         bool
		want          bool
	}{
		// case-insensitive substring matches
		{"Hello World", "hello", false, false, true},
		{"Hello World", "world", false, false, true},
		{"Hello World", "lo wo", false, false, true},
		{"Hello World", "foo", false, false, false},
		// case-sensitive substring matches
		{"Hello World", "Hello", true, false, true},
		{"Hello World", "hello", true, false, false},
		// exact word matches (case-insensitive)
		{"Hello World", "hello", false, true, true},
		{"Hello World", "world", false, true, true},
		{"Hello World", "lo", false, true, false},         // not a whole word
		{"Hello World!", "world", false, true, true},      // punctuation after word
		{"Hello World", "Hello World", false, true, true}, // entire string
		// exact word matches (case-sensitive)
		{"Hello World", "Hello", true, true, true},
		{"Hello World", "hello", true, true, false},
	}
	for _, tt := range tests {
		got := textMatches(tt.text, tt.query, tt.caseSensitive, tt.exact)
		if got != tt.want {
			t.Errorf("textMatches(%q, %q, case=%v, exact=%v) = %v, want %v",
				tt.text, tt.query, tt.caseSensitive, tt.exact, got, tt.want)
		}
	}
}

func TestMatchesSearch(t *testing.T) {
	// Helper to create a node with given properties
	mknode := func(id, name, typ, characters string) *api.Node {
		node := &api.Node{
			ID:   id,
			Name: name,
			Type: typ,
		}
		if characters != "" {
			node.Characters = characters
		}
		return node
	}

	tests := []struct {
		name      string
		node      *api.Node
		opts      *SearchOptions
		nameRegex *regexp.Regexp
		want      bool
	}{
		// Type filter
		{
			name: "type match",
			node: mknode("1", "Button", "FRAME", ""),
			opts: &SearchOptions{TypeFilter: "FRAME"},
			want: true,
		},
		{
			name: "type mismatch",
			node: mknode("1", "Button", "FRAME", ""),
			opts: &SearchOptions{TypeFilter: "COMPONENT"},
			want: false,
		},
		// Name regex filter
		{
			name:      "name regex match",
			node:      mknode("1", "SubmitButton", "FRAME", ""),
			opts:      &SearchOptions{NamePattern: ".*Button"},
			nameRegex: regexp.MustCompile(".*Button"),
			want:      true,
		},
		{
			name:      "name regex mismatch",
			node:      mknode("1", "SubmitButton", "FRAME", ""),
			opts:      &SearchOptions{NamePattern: "^Button"},
			nameRegex: regexp.MustCompile("^Button"),
			want:      false,
		},
		// Text content search
		{
			name: "text content match",
			node: mknode("1", "Label", "TEXT", "Hello World"),
			opts: &SearchOptions{Query: "Hello"},
			want: true,
		},
		{
			name: "text content mismatch",
			node: mknode("1", "Label", "TEXT", "Hello World"),
			opts: &SearchOptions{Query: "Goodbye"},
			want: false,
		},
		{
			name: "non-TEXT node with query",
			node: mknode("1", "Frame", "FRAME", ""),
			opts: &SearchOptions{Query: "Hello"},
			want: false,
		},
		// Combined filters
		{
			name:      "type + name match",
			node:      mknode("1", "SubmitButton", "FRAME", ""),
			opts:      &SearchOptions{TypeFilter: "FRAME", NamePattern: ".*Button"},
			nameRegex: regexp.MustCompile(".*Button"),
			want:      true,
		},
		{
			name:      "type match + name mismatch",
			node:      mknode("1", "SubmitButton", "FRAME", ""),
			opts:      &SearchOptions{TypeFilter: "FRAME", NamePattern: "^Button"},
			nameRegex: regexp.MustCompile("^Button"),
			want:      false,
		},
		// Type "text" special case
		{
			name: "type text matches TEXT node",
			node: mknode("1", "Label", "TEXT", "Hello"),
			opts: &SearchOptions{TypeFilter: "text"},
			want: true,
		},
		{
			name: "type text does not match non-TEXT",
			node: mknode("1", "Frame", "FRAME", ""),
			opts: &SearchOptions{TypeFilter: "text"},
			want: false,
		},
		// No filters -> false
		{
			name: "no filters",
			node: mknode("1", "Anything", "RECTANGLE", ""),
			opts: &SearchOptions{},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesSearch(tt.node, tt.opts, tt.nameRegex)
			if got != tt.want {
				t.Errorf("matchesSearch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchNodes(t *testing.T) {
	// Build a simple tree
	root := &api.Node{
		ID:   "0",
		Name: "Root",
		Type: "FRAME",
		Children: []api.Node{
			{
				ID:   "1",
				Name: "FrameA",
				Type: "FRAME",
				Children: []api.Node{
					{
						ID:         "2",
						Name:       "Text1",
						Type:       "TEXT",
						Characters: "Hello World",
					},
					{
						ID:   "3",
						Name: "FrameB",
						Type: "FRAME",
					},
				},
			},
			{
				ID:         "4",
				Name:       "Text2",
				Type:       "TEXT",
				Characters: "Goodbye World",
			},
		},
	}

	tests := []struct {
		name    string
		opts    *SearchOptions
		wantIDs []string
	}{
		{
			name:    "search text content",
			opts:    &SearchOptions{Query: "World"},
			wantIDs: []string{"2", "4"},
		},
		{
			name:    "search text exact word",
			opts:    &SearchOptions{Query: "World", Exact: true},
			wantIDs: []string{"2", "4"},
		},
		{
			name:    "search text case-sensitive",
			opts:    &SearchOptions{Query: "world", CaseSensitive: true},
			wantIDs: []string{},
		},
		{
			name:    "filter by type FRAME",
			opts:    &SearchOptions{TypeFilter: "FRAME"},
			wantIDs: []string{"0", "1", "3"},
		},
		{
			name:    "filter by type TEXT",
			opts:    &SearchOptions{TypeFilter: "TEXT"},
			wantIDs: []string{"2", "4"},
		},
		{
			name:    "name pattern",
			opts:    &SearchOptions{NamePattern: "Frame.*"},
			wantIDs: []string{"1", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := searchNodes(root, "testkey", tt.opts)
			var gotIDs []string
			for _, r := range results {
				gotIDs = append(gotIDs, r.ID)
			}
			if len(gotIDs) != len(tt.wantIDs) {
				t.Errorf("searchNodes() returned %d results, want %d", len(gotIDs), len(tt.wantIDs))
				t.Logf("got IDs: %v", gotIDs)
				t.Logf("want IDs: %v", tt.wantIDs)
				return
			}
			for i, id := range tt.wantIDs {
				if gotIDs[i] != id {
					t.Errorf("result %d: got ID %s, want %s", i, gotIDs[i], id)
				}
			}
		})
	}
}
