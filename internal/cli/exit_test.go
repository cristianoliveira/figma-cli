package cli

import (
	"errors"
	"testing"

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
