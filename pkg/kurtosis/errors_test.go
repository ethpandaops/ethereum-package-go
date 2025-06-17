package kurtosis

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrEnclaveNotFound", ErrEnclaveNotFound, "enclave not found"},
		{"ErrServiceNotFound", ErrServiceNotFound, "service not found"},
		{"ErrInvalidConfiguration", ErrInvalidConfiguration, "invalid configuration"},
		{"ErrTimeout", ErrTimeout, "operation timed out"},
		{"ErrKurtosisNotRunning", ErrKurtosisNotRunning, "kurtosis engine is not running"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Equal(t, tt.msg, tt.err.Error())
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	errs := []error{
		ErrEnclaveNotFound,
		ErrServiceNotFound,
		ErrInvalidConfiguration,
		ErrTimeout,
		ErrKurtosisNotRunning,
	}

	// Verify all errors are distinct
	for i := 0; i < len(errs); i++ {
		for j := i + 1; j < len(errs); j++ {
			assert.False(t, errors.Is(errs[i], errs[j]), "errors %v and %v should not be equal", errs[i], errs[j])
		}
	}
}