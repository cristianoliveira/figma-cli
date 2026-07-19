package cli

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExitCodeErrorCarriesCode(t *testing.T) {
	err := &ExitCodeError{Code: 1}

	assert.Equal(t, 1, err.Code)
	assert.Empty(t, err.Error(), "silent: no diagnostic message")
}

func TestExitCodeErrorRecognizedByErrorsAs(t *testing.T) {
	err := error(&ExitCodeError{Code: 1})

	var target *ExitCodeError
	assert.True(t, errors.As(err, &target))
	assert.Equal(t, 1, target.Code)
}

func TestExitCodeClassifiesUsageOperationalAndExplicitErrors(t *testing.T) {
	assert.Equal(t, 2, ExitCode(NewUsageError(errors.New("bad arguments"))))
	assert.Equal(t, 1, ExitCode(errors.New("network unavailable")))
	assert.Equal(t, 2, ExitCode(errors.New(`unknown command "reed" for "tool"`)))
	assert.Equal(t, 7, ExitCode(&ExitCodeError{Code: 7}))
}

func TestMarkUsageErrorsClassifiesCobraArgumentAndFlagErrors(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return NewUsageError(err) })
	child := &cobra.Command{Use: "read <path>", Args: cobra.ExactArgs(1), RunE: func(*cobra.Command, []string) error { return nil }}
	root.AddCommand(child)
	MarkUsageErrors(root)

	root.SetArgs([]string{"read"})
	assert.Equal(t, 2, ExitCode(root.Execute()))

	root.SetArgs([]string{"read", "file", "--unknown"})
	assert.Equal(t, 2, ExitCode(root.Execute()))
}
