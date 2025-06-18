package kurtosis

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// errorDefinition represents an error and its expected message
type errorDefinition struct {
	name string
	err  error
	msg  string
}

// getAllErrorDefinitions returns all error definitions for testing
func getAllErrorDefinitions() []errorDefinition {
	return []errorDefinition{
		{"ErrEnclaveNotFound", ErrEnclaveNotFound, "enclave not found"},
		{"ErrServiceNotFound", ErrServiceNotFound, "service not found"},
		{"ErrInvalidConfiguration", ErrInvalidConfiguration, "invalid configuration"},
		{"ErrTimeout", ErrTimeout, "operation timed out"},
		{"ErrKurtosisNotRunning", ErrKurtosisNotRunning, "kurtosis engine is not running"},
	}
}

func TestErrors(t *testing.T) {
	for _, tt := range getAllErrorDefinitions() {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Equal(t, tt.msg, tt.err.Error())
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	defs := getAllErrorDefinitions()
	errs := make([]error, len(defs))
	for i, def := range defs {
		errs[i] = def.err
	}

	// Verify all errors are distinct
	for i := range len(errs) {
		for j := i + 1; j < len(errs); j++ {
			assert.False(t, errors.Is(errs[i], errs[j]), "errors %v and %v should not be equal", errs[i], errs[j])
		}
	}
}