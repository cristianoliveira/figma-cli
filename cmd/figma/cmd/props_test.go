package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
)

func TestNodeToPropertyMap(t *testing.T) {
	node := &api.Node{
		ID:      "1:2",
		Name:    "Example Frame",
		Type:    "FRAME",
		Visible: true,
		AbsoluteBoundingBox: &api.Rectangle{
			X:      100.0,
			Y:      200.0,
			Width:  300.0,
			Height: 400.0,
		},
		RelativeTransform: api.Transform{{1.0, 0.0, 100.0}, {0.0, 1.0, 200.0}},
		Rotation:          45.0,
		IsFixed:           false,
		Constraints: &api.LayoutConstraint{
			Horizontal: "LEFT",
			Vertical:   "TOP",
		},
		LayoutMode:            "HORIZONTAL",
		PrimaryAxisAlignItems: "CENTER",
		CounterAxisAlignItems: "CENTER",
		ItemSpacing:           10.0,
		PaddingLeft:           20.0,
		PaddingRight:          20.0,
		PaddingTop:            30.0,
		PaddingBottom:         30.0,
		Opacity:               0.8,
		BlendMode:             "PASS_THROUGH",
		IsMask:                false,
		Locked:                false,
		StrokeWeight:          2.0,
		CornerRadius:          8.0,
	}

	props := nodeToPropertyMap(node)

	// Check basic properties
	if props["id"] != "1:2" {
		t.Errorf("expected id '1:2', got %v", props["id"])
	}
	if props["name"] != "Example Frame" {
		t.Errorf("expected name 'Example Frame', got %v", props["name"])
	}
	if props["type"] != "FRAME" {
		t.Errorf("expected type 'FRAME', got %v", props["type"])
	}
	if props["visible"] != true {
		t.Errorf("expected visible true, got %v", props["visible"])
	}

	// Geometry
	if props["x"] != 100.0 {
		t.Errorf("expected x 100.0, got %v", props["x"])
	}
	if props["y"] != 200.0 {
		t.Errorf("expected y 200.0, got %v", props["y"])
	}
	if props["width"] != 300.0 {
		t.Errorf("expected width 300.0, got %v", props["width"])
	}
	if props["height"] != 400.0 {
		t.Errorf("expected height 400.0, got %v", props["height"])
	}
	if props["rotation"] != 45.0 {
		t.Errorf("expected rotation 45.0, got %v", props["rotation"])
	}

	// Transform
	transform, ok := props["relativeTransform"].(api.Transform)
	if !ok || len(transform) != 2 || len(transform[0]) != 3 || transform[0][0] != 1.0 {
		t.Errorf("unexpected relativeTransform: %v", props["relativeTransform"])
	}

	// Constraints
	constraints, ok := props["constraints"].(map[string]interface{})
	if !ok || constraints["horizontal"] != "LEFT" || constraints["vertical"] != "TOP" {
		t.Errorf("unexpected constraints: %v", props["constraints"])
	}
	if props["isFixed"] != false {
		t.Errorf("expected isFixed false, got %v", props["isFixed"])
	}

	// Layout
	layout, ok := props["layout"].(map[string]interface{})
	if !ok || layout["layoutMode"] != "HORIZONTAL" {
		t.Errorf("unexpected layout: %v", props["layout"])
	}

	// Additional properties
	if props["opacity"] != 0.8 {
		t.Errorf("expected opacity 0.8, got %v", props["opacity"])
	}
	if props["blendMode"] != "PASS_THROUGH" {
		t.Errorf("expected blendMode PASS_THROUGH, got %v", props["blendMode"])
	}
	if props["strokeWeight"] != 2.0 {
		t.Errorf("expected strokeWeight 2.0, got %v", props["strokeWeight"])
	}
	if props["cornerRadius"] != 8.0 {
		t.Errorf("expected cornerRadius 8.0, got %v", props["cornerRadius"])
	}
}

func TestNodeToPropertyMapMissingFields(t *testing.T) {
	node := &api.Node{
		ID:   "1:1",
		Name: "Simple",
		Type: "RECTANGLE",
	}
	props := nodeToPropertyMap(node)
	// Should not have geometry fields
	if _, ok := props["x"]; ok {
		t.Error("x should not be present when AbsoluteBoundingBox is nil")
	}
	if _, ok := props["rotation"]; ok {
		t.Error("rotation should not be present when zero")
	}
	if _, ok := props["layout"]; ok {
		t.Error("layout should not be present when LayoutMode empty")
	}
}

func TestFilterProperties(t *testing.T) {
	props := map[string]interface{}{
		"id":       "1:2",
		"name":     "Frame",
		"type":     "FRAME",
		"x":        100.0,
		"y":        200.0,
		"width":    300.0,
		"height":   400.0,
		"rotation": 45.0,
		"isFixed":  false,
		"opacity":  0.8,
	}

	// Filter by group
	filtered := filterProperties(props, "geometry")
	if len(filtered) != 5 {
		t.Errorf("expected 5 geometry fields, got %v", filtered)
	}
	if _, ok := filtered["x"]; !ok {
		t.Error("geometry should include x")
	}
	if _, ok := filtered["id"]; ok {
		t.Error("geometry should not include id")
	}

	// Filter by multiple groups
	filtered = filterProperties(props, "geometry,basic")
	if len(filtered) != 8 { // id,name,type + geometry fields (x,y,width,height,rotation)
		t.Errorf("expected 8 fields, got %v", filtered)
	}
	if _, ok := filtered["isFixed"]; ok {
		t.Error("should not include isFixed")
	}

	// Filter by specific field
	filtered = filterProperties(props, "opacity")
	if len(filtered) != 1 || filtered["opacity"] != 0.8 {
		t.Errorf("expected only opacity, got %v", filtered)
	}

	// Empty filter (should return empty)
	filtered = filterProperties(props, "")
	if len(filtered) != 0 {
		t.Errorf("empty filter should return empty map, got %v", filtered)
	}
}

func TestRunProps(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "node_props.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("Could not read fixture: %v", err)
	}

	// Create mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Request: %s %s", r.Method, r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/v1/files/") && strings.Contains(r.URL.Path, "/nodes") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
		// Fallback to file endpoint (not needed for this test)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"err": "Not Found"}`))
	}))
	defer ts.Close()

	// Create config with test server URL
	cfg := config.DefaultConfig()
	cfg.API.BaseURL = ts.URL + "/v1"
	cfg.API.Timeout = 30 * time.Second
	cfg.API.Tier = 1
	cfg.API.SeatType = "dev_full"
	cfg.Token = "testtoken"
	cfg.TokenType = "pat"
	cfg.OutputFormat = config.OutputFormatText

	// Create a no-op logger
	logger := logging.NewNopLogger()

	// Create command context
	ctx := context.Background()
	ctx = context.WithValue(ctx, configKey{}, &cfg)
	ctx = context.WithValue(ctx, loggerKey{}, logger)

	// Create props command and set context
	cmd := propsCmd
	cmd.SetContext(ctx)

	// Set flags (defaults)
	cmd.Flags().Set("fields", "")
	cmd.Flags().Set("output-format", "")

	// Execute with a file key and node ID that matches the fixture
	err = runProps(cmd, []string{"https://www.figma.com/file/ABC123/Design?node-id=1:2"})
	if err != nil {
		t.Errorf("runProps failed: %v", err)
	}
}

func TestRunPropsWithFieldsFilter(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "node_props.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("Could not read fixture: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/files/") && strings.Contains(r.URL.Path, "/nodes") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	cfg := config.DefaultConfig()
	cfg.API.BaseURL = ts.URL + "/v1"
	cfg.API.Timeout = 30 * time.Second
	cfg.Token = "testtoken"
	cfg.TokenType = "pat"
	cfg.OutputFormat = config.OutputFormatText

	logger := logging.NewNopLogger()

	ctx := context.Background()
	ctx = context.WithValue(ctx, configKey{}, &cfg)
	ctx = context.WithValue(ctx, loggerKey{}, logger)

	cmd := propsCmd
	cmd.SetContext(ctx)

	// Set fields filter
	cmd.Flags().Set("fields", "geometry,constraints")
	cmd.Flags().Set("output-format", "")

	err = runProps(cmd, []string{"https://www.figma.com/file/ABC123/Design?node-id=1:2"})
	if err != nil {
		t.Errorf("runProps with fields filter failed: %v", err)
	}
}

func TestRunPropsOutputFormats(t *testing.T) {
	// Load fixture
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "node_props.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Skipf("Could not read fixture: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/v1/files/") && strings.Contains(r.URL.Path, "/nodes") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	cfg := config.DefaultConfig()
	cfg.API.BaseURL = ts.URL + "/v1"
	cfg.API.Timeout = 30 * time.Second
	cfg.Token = "testtoken"
	cfg.TokenType = "pat"

	logger := logging.NewNopLogger()

	ctx := context.Background()
	ctx = context.WithValue(ctx, configKey{}, &cfg)
	ctx = context.WithValue(ctx, loggerKey{}, logger)

	// Test each output format
	formats := []string{config.OutputFormatJSON, outputFormatPretty, config.OutputFormatYAML, config.OutputFormatText}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cmd := propsCmd
			cmd.SetContext(ctx)
			cmd.Flags().Set("fields", "")
			cmd.Flags().Set("output-format", format)
			err = runProps(cmd, []string{"https://www.figma.com/file/ABC123/Design?node-id=1:2"})
			if err != nil {
				t.Errorf("runProps with output format %s failed: %v", format, err)
			}
		})
	}
}
