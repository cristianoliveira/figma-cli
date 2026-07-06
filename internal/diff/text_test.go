package diff

import "testing"

func TestText(t *testing.T) {
	from := []TextNode{
		{ID: "1:1", Name: "Title", Text: "Old title"},
		{ID: "1:2", Name: "Removed", Text: "Gone"},
		{ID: "1:3", Name: "Same", Text: "Keep"},
	}
	to := []TextNode{
		{ID: "1:1", Name: "Title", Text: "New title"},
		{ID: "1:3", Name: "Same", Text: "Keep"},
		{ID: "1:4", Name: "Added", Text: "New"},
	}

	got := Text(from, to)

	if len(got.Changed) != 1 || got.Changed[0].From != "Old title" || got.Changed[0].To != "New title" {
		t.Fatalf("changed diff = %#v", got.Changed)
	}
	if len(got.Removed) != 1 || got.Removed[0].Text != "Gone" {
		t.Fatalf("removed diff = %#v", got.Removed)
	}
	if len(got.Added) != 1 || got.Added[0].Text != "New" {
		t.Fatalf("added diff = %#v", got.Added)
	}
}
