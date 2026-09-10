package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/inspect"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeInspectService captures the request the command sends and returns a
// pre-configured result. It records the calls so tests can assert that
// the command mapped flags/arguments correctly without depending on the
// real Figma adapter.
type fakeInspectService struct {
	gotRequest inspect.Request
	result     *inspect.Result
	err        error
	calls      int
}

func (f *fakeInspectService) Inspect(req inspect.Request) (*inspect.Result, error) {
	f.gotRequest = req
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

// depsWithFakeInspect builds a Deps whose InspectService field is wired
// to the supplied fake. LoadClient is also supplied so flag-validation
// errors that fire before the service is consulted can use it (it
// returns an error if invoked from the test paths we care about).
func depsWithFakeInspect(fake *fakeInspectService) Deps {
	deps := Deps{
		LoadClient: func() (*figma.Client, error) {
			return nil, errors.New("must not load client in inspect command tests")
		},
		ResolveExec: func() (string, error) { return "/tmp/figma", nil },
		GetEnv:      func(string) string { return "" },
		InspectService: func() (InspectService, error) {
			return fake, nil
		},
	}
	return deps
}

func TestInspectCommandEmitsStableScopedContract(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeSingle,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Single: &extract.InspectOutput{
			ID: "42:1", Name: "Button", Type: "COMPONENT",
		},
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"result":{"id":"42:1","name":"Button","type":"COMPONENT","bounds":{},"layout":{},"typography":{}}}`, result.Stdout)
	require.Equal(t, 1, fake.calls)
	assert.Equal(t, "abc", fake.gotRequest.FileID)
	assert.Equal(t, "42:1", fake.gotRequest.NodeID)
	assert.False(t, fake.gotRequest.Recursive)
	assert.False(t, fake.gotRequest.Handoff)
	assert.False(t, fake.gotRequest.IncludeVectorPaths)
}

func TestInspectCommandIncludesVectorPathsOnExplicitRequest(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeSingle,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Single: &extract.InspectOutput{
			ID: "42:1", Name: "Wave", Type: "VECTOR",
		},
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--include-vector-paths")

	require.NoError(t, result.Err)
	require.True(t, fake.gotRequest.IncludeVectorPaths, "single + vector paths must propagate to service")
	// Single-node geometry depth still defaults to 1; the transport
	// layer applies it because the service itself handles the
	// single-node depth conversion when vector paths are requested.
}

func TestInspectCommandRequiresDepthForRecursiveVectorPaths(t *testing.T) {
	result := executeCommand(newInspectCommand(depsWithFakeInspect(&fakeInspectService{})),
		"abc", "--id", "1:1", "--recursive", "--include-vector-paths")
	assert.EqualError(t, result.Err, "--include-vector-paths with --recursive requires explicit --depth")
}

func TestInspectCommandRecursivelyEmitsImplementationSpecs(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeRecursive,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Nodes: []extract.InspectOutput{
			{ID: "42:1", Name: "Button", Type: "COMPONENT"},
			{ID: "42:2", Name: "Label", Type: "TEXT", Text: "Save"},
		},
		Total: 2,
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--recursive")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"total":2,"results":[{"id":"42:1","name":"Button","type":"COMPONENT","bounds":{},"layout":{},"typography":{}},{"id":"42:2","name":"Label","type":"TEXT","text":"Save","bounds":{},"layout":{},"typography":{}}]}`, result.Stdout)
	assert.True(t, fake.gotRequest.Recursive)
}

func TestInspectCommandBoundsRecursiveTraversalWithDepth(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeRecursive,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Nodes: []extract.InspectOutput{
			{ID: "42:1", Name: "Root", Type: "FRAME"},
			{ID: "42:2", Name: "Child", Type: "FRAME"},
		},
		Total: 2,
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--recursive", "--depth", "1")

	require.NoError(t, result.Err)
	assert.Equal(t, 1, fake.gotRequest.Depth)
}

func TestInspectCommandRecursiveEmitsComputedSiblingSpacing(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeRecursive,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Nodes: []extract.InspectOutput{
			{ID: "42:1", Name: "Stack", Type: "FRAME"},
			{ID: "42:2", Name: "Copy", Type: "TEXT"},
			{ID: "42:3", Name: "Link", Type: "TEXT"},
		},
		Total: 3,
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--recursive")

	require.NoError(t, result.Err)
	assert.True(t, fake.gotRequest.Recursive)
}

func TestInspectCommandAcceptsNodeAliasForExplicitID(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:   inspect.ModeSingle,
		Scope:  output.Scope{FileKey: "exampleFileKey123", NodeIDs: []string{"0:147"}},
		Single: &extract.InspectOutput{ID: "0:147", Name: "Target", Type: "FRAME"},
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"exampleFileKey123", "--node", "0:147")

	require.NoError(t, result.Err)
	assert.Equal(t, "0:147", fake.gotRequest.NodeID)
}

func TestInspectCommandRejectsConflictingIDAndNodeAlias(t *testing.T) {
	result := executeCommand(newInspectCommand(depsWithFakeInspect(&fakeInspectService{})),
		"abc", "--id", "0:147", "--node", "0:148")

	assert.EqualError(t, result.Err, "--id and --node must match when both are provided")
}

func TestInspectCommandRecursiveExplicitIDEmitsRelativeBounds(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:   inspect.ModeSingle,
		Scope:  output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Single: &extract.InspectOutput{ID: "42:1", Name: "Button", Type: "COMPONENT"},
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"abc", "--id", "42:1", "--recursive")

	require.NoError(t, result.Err)
	assert.True(t, fake.gotRequest.Recursive)
}

func TestInspectCommandEmitsBoundedHandoff(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeHandoff,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Nodes: []extract.InspectOutput{
			{ID: "42:1", Name: "Checkout", Type: "FRAME"},
		},
		Handoff: &extract.HandoffOutput{},
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--handoff", "--depth", "1")

	require.NoError(t, result.Err)
	assert.True(t, fake.gotRequest.Handoff)
	assert.Equal(t, 1, fake.gotRequest.Depth)
}

func TestInspectCommandRendersSelectedTextFields(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeRecursive,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Nodes: []extract.InspectOutput{
			{ID: "42:1", Name: "Card", Type: "FRAME"},
		},
		Total: 1,
		Text:  "Card (FRAME) x:0 y:0 w:320 h:800 layout:VERTICAL gap:8 fills:#FFFFFF\n  Label (TEXT) x:16 y:12 w:84 h:20\n",
	}}
	result := executeDefaultCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1",
		"--recursive", "--format", "text", "--fields", "name,type,relativeBounds,layout.mode,layout.gap,fills",
	)

	require.NoError(t, result.Err)
	assert.Equal(t, inspect.Format("text"), fake.gotRequest.Format)
	assert.Equal(t, []string{"name", "type", "relativeBounds", "layout.mode", "layout.gap", "fills"}, fake.gotRequest.Fields)
	assert.Contains(t, result.Stdout, "Card")
}

func TestInspectCommandValidatesTextProjectionBeforeLoadingService(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		expected string
	}{
		{name: "unknown format", flags: []string{"--recursive", "--format", "yaml"}, expected: `unknown inspect format "yaml" (want json or text)`},
		{name: "text requires recursive", flags: []string{"--format", "text"}, expected: "--format text requires --recursive"},
		{name: "fields require text", flags: []string{"--recursive", "--fields", "name"}, expected: "--fields requires --format text"},
		{name: "unknown field", flags: []string{"--recursive", "--format", "text", "--fields", "layout.padding"}, expected: `unknown inspect field "layout.padding"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeInspectService{}
			args := append([]string{"abc", "--id", "1:1"}, test.flags...)
			result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)), args...)

			assert.EqualError(t, result.Err, test.expected)
			assert.Equal(t, 0, fake.calls, "service must not be called for flag validation failures")
		})
	}
}

func TestInspectCommandRejectsHandoffWithRecursive(t *testing.T) {
	result := executeCommand(newInspectCommand(depsWithFakeInspect(&fakeInspectService{})),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--handoff", "--recursive")

	assert.EqualError(t, result.Err, "--handoff and --recursive cannot be used together")
}

func TestInspectCommandForwardsServiceErrors(t *testing.T) {
	sentinel := errors.New("upstream timed out")
	fake := &fakeInspectService{err: sentinel}

	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1")

	require.Error(t, result.Err)
	assert.True(t, errors.Is(result.Err, sentinel), "command must surface service errors unchanged")
}

func TestInspectCommandRejectsMissingScopeWithoutLoadingService(t *testing.T) {
	fake := &fakeInspectService{}
	command := newInspectCommand(depsWithFakeInspect(fake))
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs([]string{"abc"})

	err := command.Execute()

	assert.EqualError(t, err, "inspect requires a Figma URL with node-id or --id")
	assert.Equal(t, 0, fake.calls)
}

func TestInspectCommandRecursiveWithoutDepthPassesUnboundedSentinel(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeRecursive,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Nodes: []extract.InspectOutput{},
		Total: 0,
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--recursive")

	require.NoError(t, result.Err)
	assert.True(t, fake.gotRequest.Recursive)
	assert.Equal(t, -1, fake.gotRequest.Depth, "unspecified --depth must translate to the unbounded sentinel")
}

func TestInspectCommandRecursiveWithExplicitDepthPassesIt(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:  inspect.ModeRecursive,
		Scope: output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Nodes: []extract.InspectOutput{},
		Total: 0,
	}}
	result := executeCommand(newInspectCommand(depsWithFakeInspect(fake)),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--recursive", "--depth", "2")

	require.NoError(t, result.Err)
	assert.Equal(t, 2, fake.gotRequest.Depth, "explicit --depth must be forwarded to the service")
}

func TestInspectCommandContextPropagatesToService(t *testing.T) {
	fake := &fakeInspectService{result: &inspect.Result{
		Mode:   inspect.ModeSingle,
		Scope:  output.Scope{FileKey: "abc", NodeIDs: []string{"42:1"}},
		Single: &extract.InspectOutput{ID: "42:1", Name: "Button", Type: "COMPONENT"},
	}}
	deps := depsWithFakeInspect(fake)
	command := newInspectCommand(deps)
	command.SilenceErrors = true
	command.SilenceUsage = true
	type assertion struct{}
	want := context.WithValue(context.Background(), assertion{}, "marker")
	command.SetContext(want)
	command.SetArgs([]string{"https://www.figma.com/design/abc/Name?node-id=42-1"})

	require.NoError(t, command.Execute())
	assert.NotNil(t, fake.gotRequest.Context)
	got := fake.gotRequest.Context.Value(assertion{})
	assert.Equal(t, "marker", got)
}

func TestInspectNodeIDDefaultsToURLNode(t *testing.T) {
	input, err := figma.ParseInput("https://www.figma.com/design/abc/Name?node-id=42-1")
	require.NoError(t, err)

	nodeID, err := inspectNodeID(input, "")

	require.NoError(t, err)
	assert.Equal(t, "42:1", nodeID)
}

func TestInspectNodeIDRequiresScopeForBareFile(t *testing.T) {
	input, err := figma.ParseInput("abc")
	require.NoError(t, err)

	_, err = inspectNodeID(input, "")

	assert.EqualError(t, err, "inspect requires a Figma URL with node-id or --id")
}

func TestInspectNodeIDRejectsMultipleNodes(t *testing.T) {
	input, err := figma.ParseInput("https://www.figma.com/design/abc/Name?node-id=42-1,42-2")
	require.NoError(t, err)

	_, err = inspectNodeID(input, "")

	assert.EqualError(t, err, "inspect requires exactly one node ID")
}

// Ensure the inspect package does not pull Cobra in (compile-time check).
var _ = cobra.Command{}
