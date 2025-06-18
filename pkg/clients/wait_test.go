package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPWaitStrategy_WaitUntilReady(t *testing.T) {
	tests := []struct {
		name          string
		strategy      *HTTPWaitStrategy
		serverHandler func(*atomic.Int32) http.HandlerFunc
		expectedCalls int
		expectError   bool
	}{
		{
			name: "immediate success",
			strategy: NewHTTPWaitStrategy(8080).
				WithPath("/health").
				WithStatusCode(200).
				WithInterval(10 * time.Millisecond),
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					assert.Equal(t, "/health", r.URL.Path)
					w.WriteHeader(http.StatusOK)
				}
			},
			expectedCalls: 1,
			expectError:   false,
		},
		{
			name: "success after retries",
			strategy: NewHTTPWaitStrategy(8080).
				WithPath("/ready").
				WithStatusCode(200).
				WithInterval(10 * time.Millisecond),
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					count := calls.Add(1)
					if count < 3 {
						w.WriteHeader(http.StatusServiceUnavailable)
					} else {
						w.WriteHeader(http.StatusOK)
					}
				}
			},
			expectedCalls: 3,
			expectError:   false,
		},
		{
			name: "timeout",
			strategy: NewHTTPWaitStrategy(8080).
				WithPath("/never-ready").
				WithTimeout(50 * time.Millisecond).
				WithInterval(10 * time.Millisecond),
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					w.WriteHeader(http.StatusServiceUnavailable)
				}
			},
			expectedCalls: 3, // Approximate calls before timeout
			expectError:   true,
		},
		{
			name: "custom status code",
			strategy: NewHTTPWaitStrategy(8080).
				WithMethod("POST").
				WithStatusCode(201).
				WithInterval(10 * time.Millisecond),
			serverHandler: func(calls *atomic.Int32) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					assert.Equal(t, "POST", r.Method)
					w.WriteHeader(http.StatusCreated)
				}
			},
			expectedCalls: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(tt.serverHandler(&calls))
			defer server.Close()

			ctx := context.Background()
			err := tt.strategy.WaitUntilReady(ctx, server.URL)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check approximate number of calls
			actualCalls := calls.Load()
			assert.GreaterOrEqual(t, actualCalls, int32(tt.expectedCalls))
		})
	}
}

func TestHTTPWaitStrategy_ParseURL(t *testing.T) {
	strategy := NewHTTPWaitStrategy(8080)

	// Test with invalid target type
	err := strategy.WaitUntilReady(context.Background(), 123)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported target type")

	// Test with empty URL
	err = strategy.WaitUntilReady(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no URL available")
}

func TestSyncWaitStrategy_WaitUntilReady(t *testing.T) {
	// Mock client that implements WaitForSync
	mockClient := &mockSyncClient{syncAfter: 3}

	strategy := NewSyncWaitStrategy().
		WithTimeout(100 * time.Millisecond).
		WithInterval(10 * time.Millisecond)

	ctx := context.Background()
	err := strategy.WaitUntilReady(ctx, mockClient)
	assert.NoError(t, err)
	// The strategy calls WaitForSync with a 1-second timeout.
	// Our mock returns nil after 5ms, so it should succeed on the first call
	assert.GreaterOrEqual(t, mockClient.syncCalls, 1)
}

func TestSyncWaitStrategy_UnsupportedTarget(t *testing.T) {
	strategy := NewSyncWaitStrategy()

	// Test with unsupported target
	err := strategy.WaitUntilReady(context.Background(), "not-a-client")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target does not support sync waiting")
}

func TestHealthWaitStrategy_WaitUntilReady(t *testing.T) {
	// Mock client that implements IsHealthy
	mockClient := &mockHealthClient{healthyAfter: 3}

	strategy := NewHealthWaitStrategy().
		WithTimeout(100 * time.Millisecond).
		WithInterval(10 * time.Millisecond)

	ctx := context.Background()
	err := strategy.WaitUntilReady(ctx, mockClient)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, mockClient.healthCalls, 3)
}

func TestHealthWaitStrategy_Timeout(t *testing.T) {
	// Mock client that is never healthy
	mockClient := &mockUnhealthyClient{}

	strategy := NewHealthWaitStrategy().
		WithTimeout(50 * time.Millisecond).
		WithInterval(10 * time.Millisecond)

	ctx := context.Background()
	err := strategy.WaitUntilReady(ctx, mockClient)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for healthy status")
}

func TestCombinedWaitStrategy_Sequential(t *testing.T) {
	var calls []string

	// Create strategies that track order
	strategy1 := &mockWaitStrategy{
		name: "strategy1",
		waitFunc: func(ctx context.Context, target interface{}) error {
			calls = append(calls, "strategy1")
			return nil
		},
	}

	strategy2 := &mockWaitStrategy{
		name: "strategy2",
		waitFunc: func(ctx context.Context, target interface{}) error {
			calls = append(calls, "strategy2")
			return nil
		},
	}

	combined := NewCombinedWaitStrategy(strategy1, strategy2)

	err := combined.WaitUntilReady(context.Background(), "target")
	assert.NoError(t, err)
	assert.Equal(t, []string{"strategy1", "strategy2"}, calls)
}

func TestCombinedWaitStrategy_Parallel(t *testing.T) {
	// Use channels to track parallel execution
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})

	strategy1 := &mockWaitStrategy{
		name: "strategy1",
		waitFunc: func(ctx context.Context, target interface{}) error {
			close(ch1)
			<-ch2 // Wait for strategy2
			return nil
		},
	}

	strategy2 := &mockWaitStrategy{
		name: "strategy2",
		waitFunc: func(ctx context.Context, target interface{}) error {
			close(ch2)
			<-ch1 // Wait for strategy1
			return nil
		},
	}

	combined := NewCombinedWaitStrategy(strategy1, strategy2).WithParallel(true)

	err := combined.WaitUntilReady(context.Background(), "target")
	assert.NoError(t, err)
}

func TestCombinedWaitStrategy_Error(t *testing.T) {
	strategy1 := &mockWaitStrategy{
		name: "strategy1",
		waitFunc: func(ctx context.Context, target interface{}) error {
			return nil
		},
	}

	strategy2 := &mockWaitStrategy{
		name: "strategy2",
		waitFunc: func(ctx context.Context, target interface{}) error {
			return assert.AnError
		},
	}

	// Test sequential error
	combined := NewCombinedWaitStrategy(strategy1, strategy2)
	err := combined.WaitUntilReady(context.Background(), "target")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wait strategy 1 failed")

	// Test parallel error
	combinedParallel := NewCombinedWaitStrategy(strategy1, strategy2).WithParallel(true)
	err = combinedParallel.WaitUntilReady(context.Background(), "target")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wait strategy 1 failed")
}

func TestDefaultWaitStrategies(t *testing.T) {
	// Test DefaultExecutionClientWait
	execWait := DefaultExecutionClientWait()
	assert.NotNil(t, execWait)
	combined, ok := execWait.(*CombinedWaitStrategy)
	require.True(t, ok)
	assert.Len(t, combined.strategies, 2)

	// Test DefaultConsensusClientWait
	consWait := DefaultConsensusClientWait()
	assert.NotNil(t, consWait)
	combined, ok = consWait.(*CombinedWaitStrategy)
	require.True(t, ok)
	assert.Len(t, combined.strategies, 2)
}

// mockWaitStrategy is a test helper
type mockWaitStrategy struct {
	name     string
	waitFunc func(context.Context, interface{}) error
}

func (m *mockWaitStrategy) WaitUntilReady(ctx context.Context, target interface{}) error {
	if m.waitFunc != nil {
		return m.waitFunc(ctx, target)
	}
	return nil
}

// mockSyncClient for testing sync wait strategy
type mockSyncClient struct {
	syncCalls int
	syncAfter int
}

func (m *mockSyncClient) WaitForSync(ctx context.Context) error {
	m.syncCalls++
	if m.syncCalls >= m.syncAfter {
		return nil // Synced
	}
	// Simulate still syncing by blocking
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Millisecond):
		return nil
	}
}

// mockHealthClient for testing health wait strategy
type mockHealthClient struct {
	healthCalls  int
	healthyAfter int
}

func (m *mockHealthClient) IsHealthy(ctx context.Context) bool {
	m.healthCalls++
	return m.healthCalls >= m.healthyAfter
}

// mockUnhealthyClient for testing timeout scenarios
type mockUnhealthyClient struct{}

func (m *mockUnhealthyClient) IsHealthy(ctx context.Context) bool {
	return false
}