package extract

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

func TestExtractNodeIDFromClientMeta_FrameOffset(t *testing.T) {
	var cm api.Comment_ClientMeta
	if err := cm.FromFrameOffset(api.FrameOffset{NodeId: "1:2", NodeOffset: api.Vector{}}); err != nil {
		t.Fatalf("FromFrameOffset: %v", err)
	}

	if got := ExtractNodeIDFromClientMeta(cm); got != "1:2" {
		t.Errorf("got %q, want 1:2", got)
	}
}

func TestExtractNodeIDFromClientMeta_VectorEmpty(t *testing.T) {
	// A plain Vector carries no node_id.
	var cm api.Comment_ClientMeta
	if err := cm.FromVector(api.Vector{}); err != nil {
		t.Fatalf("FromVector: %v", err)
	}

	if got := ExtractNodeIDFromClientMeta(cm); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}
