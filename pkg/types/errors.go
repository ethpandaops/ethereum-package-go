package types

import "errors"

var (
	// Config errors
	ErrInvalidPreset   = errors.New("invalid preset")
	ErrEmptyConfigPath = errors.New("config path is empty")
	ErrNilConfig       = errors.New("config is nil")

	// Network errors
	ErrNetworkNotFound = errors.New("network not found")
	ErrServiceNotFound = errors.New("service not found")

	// Client errors
	ErrClientNotFound = errors.New("client not found")
	ErrInvalidClient  = errors.New("invalid client type")
)