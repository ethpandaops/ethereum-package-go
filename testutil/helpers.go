// Package testutil provides common test utilities for the ethereum-package-go project.
package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertErrorContains checks that an error occurs and contains the expected substring
func AssertErrorContains(t *testing.T, err error, contains string) {
	t.Helper()
	require.Error(t, err)
	assert.Contains(t, err.Error(), contains)
}

// AssertNoErrorOrFail checks that no error occurred, failing the test if one did
func AssertNoErrorOrFail(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// TableTest represents a generic table test case
type TableTest[T any] struct {
	Name    string
	Input   T
	Want    interface{}
	WantErr string
}

// RunTableTests runs a set of table-driven tests
func RunTableTests[T any](t *testing.T, tests []TableTest[T], testFunc func(t *testing.T, tt TableTest[T])) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testFunc(t, tt)
		})
	}
}

// AssertSliceEqual checks that two slices are equal in length and content
func AssertSliceEqual[T comparable](t *testing.T, expected, actual []T) {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "slice lengths differ")
	for i := range expected {
		assert.Equal(t, expected[i], actual[i], "element at index %d differs", i)
	}
}

// AssertMapEqual checks that two maps are equal
func AssertMapEqual[K comparable, V comparable](t *testing.T, expected, actual map[K]V) {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "map lengths differ")
	for k, v := range expected {
		actualV, ok := actual[k]
		require.True(t, ok, "key %v not found in actual map", k)
		assert.Equal(t, v, actualV, "value for key %v differs", k)
	}
}
