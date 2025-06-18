package kurtosis

import "errors"

var (
	// ErrEnclaveNotFound is returned when an enclave cannot be found
	ErrEnclaveNotFound = errors.New("enclave not found")

	// ErrServiceNotFound is returned when a service cannot be found
	ErrServiceNotFound = errors.New("service not found")

	// ErrInvalidConfiguration is returned when configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")

	// ErrKurtosisNotRunning is returned when Kurtosis engine is not running
	ErrKurtosisNotRunning = errors.New("kurtosis engine is not running")
)
