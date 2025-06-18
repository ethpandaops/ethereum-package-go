package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// WaitStrategy defines how to wait for a service to be ready
type WaitStrategy interface {
	WaitUntilReady(ctx context.Context, target interface{}) error
}

// HTTPWaitStrategy waits for an HTTP endpoint to respond successfully
type HTTPWaitStrategy struct {
	Port       int
	Path       string
	Method     string
	StatusCode int
	Timeout    time.Duration
	Interval   time.Duration
}

// NewHTTPWaitStrategy creates a new HTTP wait strategy with defaults
func NewHTTPWaitStrategy(port int) *HTTPWaitStrategy {
	return &HTTPWaitStrategy{
		Port:       port,
		Path:       "/",
		Method:     "GET",
		StatusCode: 200,
		Timeout:    5 * time.Minute,
		Interval:   5 * time.Second,
	}
}

// WithPath sets the path to check
func (h *HTTPWaitStrategy) WithPath(path string) *HTTPWaitStrategy {
	h.Path = path
	return h
}

// WithMethod sets the HTTP method
func (h *HTTPWaitStrategy) WithMethod(method string) *HTTPWaitStrategy {
	h.Method = method
	return h
}

// WithStatusCode sets the expected status code
func (h *HTTPWaitStrategy) WithStatusCode(code int) *HTTPWaitStrategy {
	h.StatusCode = code
	return h
}

// WithTimeout sets the overall timeout
func (h *HTTPWaitStrategy) WithTimeout(timeout time.Duration) *HTTPWaitStrategy {
	h.Timeout = timeout
	return h
}

// WithInterval sets the check interval
func (h *HTTPWaitStrategy) WithInterval(interval time.Duration) *HTTPWaitStrategy {
	h.Interval = interval
	return h
}

// WaitUntilReady waits for the HTTP endpoint to be ready
func (h *HTTPWaitStrategy) WaitUntilReady(ctx context.Context, target interface{}) error {
	var url string

	switch t := target.(type) {
	case ExecutionClient:
		url = t.RPCURL()
	case ConsensusClient:
		url = t.BeaconAPIURL()
	case string:
		url = t
	default:
		return fmt.Errorf("unsupported target type for HTTP wait strategy")
	}

	if url == "" {
		return fmt.Errorf("no URL available for waiting")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	timeout := time.After(h.Timeout)
	ticker := time.NewTicker(h.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for %s to be ready", url)
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, h.Method, url+h.Path, nil)
			if err != nil {
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == h.StatusCode {
				return nil
			}
		}
	}
}

// SyncWaitStrategy waits for a client to finish syncing
type SyncWaitStrategy struct {
	Timeout  time.Duration
	Interval time.Duration
}

// NewSyncWaitStrategy creates a new sync wait strategy
func NewSyncWaitStrategy() *SyncWaitStrategy {
	return &SyncWaitStrategy{
		Timeout:  10 * time.Minute,
		Interval: 10 * time.Second,
	}
}

// WithTimeout sets the timeout for sync waiting
func (s *SyncWaitStrategy) WithTimeout(timeout time.Duration) *SyncWaitStrategy {
	s.Timeout = timeout
	return s
}

// WithInterval sets the check interval for sync status
func (s *SyncWaitStrategy) WithInterval(interval time.Duration) *SyncWaitStrategy {
	s.Interval = interval
	return s
}

// WaitUntilReady waits for the client to finish syncing
func (s *SyncWaitStrategy) WaitUntilReady(ctx context.Context, target interface{}) error {
	timeout := time.After(s.Timeout)
	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for sync to complete")
		case <-ticker.C:
			switch client := target.(type) {
			case interface{ WaitForSync(context.Context) error }:
				// Try to sync - if it returns immediately, we're synced
				syncCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
				err := client.WaitForSync(syncCtx)
				cancel()

				if err == nil {
					return nil // Synced
				}
				if err == context.DeadlineExceeded {
					continue // Still syncing
				}
				return err // Other error
			default:
				return fmt.Errorf("target does not support sync waiting")
			}
		}
	}
}

// HealthWaitStrategy waits for a client to report healthy status
type HealthWaitStrategy struct {
	Timeout  time.Duration
	Interval time.Duration
}

// NewHealthWaitStrategy creates a new health wait strategy
func NewHealthWaitStrategy() *HealthWaitStrategy {
	return &HealthWaitStrategy{
		Timeout:  5 * time.Minute,
		Interval: 5 * time.Second,
	}
}

// WithTimeout sets the timeout for health checking
func (h *HealthWaitStrategy) WithTimeout(timeout time.Duration) *HealthWaitStrategy {
	h.Timeout = timeout
	return h
}

// WithInterval sets the check interval for health status
func (h *HealthWaitStrategy) WithInterval(interval time.Duration) *HealthWaitStrategy {
	h.Interval = interval
	return h
}

// WaitUntilReady waits for the client to report healthy status
func (h *HealthWaitStrategy) WaitUntilReady(ctx context.Context, target interface{}) error {
	timeout := time.After(h.Timeout)
	ticker := time.NewTicker(h.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for healthy status")
		case <-ticker.C:
			switch client := target.(type) {
			case interface{ IsHealthy(context.Context) bool }:
				if client.IsHealthy(ctx) {
					return nil
				}
			default:
				return fmt.Errorf("target does not support health checking")
			}
		}
	}
}

// CombinedWaitStrategy combines multiple wait strategies
type CombinedWaitStrategy struct {
	strategies []WaitStrategy
	parallel   bool
}

// NewCombinedWaitStrategy creates a new combined wait strategy
func NewCombinedWaitStrategy(strategies ...WaitStrategy) *CombinedWaitStrategy {
	return &CombinedWaitStrategy{
		strategies: strategies,
		parallel:   false,
	}
}

// WithParallel sets whether to run strategies in parallel
func (c *CombinedWaitStrategy) WithParallel(parallel bool) *CombinedWaitStrategy {
	c.parallel = parallel
	return c
}

// WaitUntilReady executes all wait strategies
func (c *CombinedWaitStrategy) WaitUntilReady(ctx context.Context, target interface{}) error {
	if c.parallel {
		return c.waitParallel(ctx, target)
	}
	return c.waitSequential(ctx, target)
}

// waitSequential executes strategies one after another
func (c *CombinedWaitStrategy) waitSequential(ctx context.Context, target interface{}) error {
	for i, strategy := range c.strategies {
		if err := strategy.WaitUntilReady(ctx, target); err != nil {
			return fmt.Errorf("wait strategy %d failed: %w", i, err)
		}
	}
	return nil
}

// waitParallel executes strategies in parallel
func (c *CombinedWaitStrategy) waitParallel(ctx context.Context, target interface{}) error {
	errChan := make(chan error, len(c.strategies))

	for i, strategy := range c.strategies {
		go func(idx int, s WaitStrategy) {
			if err := s.WaitUntilReady(ctx, target); err != nil {
				errChan <- fmt.Errorf("wait strategy %d failed: %w", idx, err)
			} else {
				errChan <- nil
			}
		}(i, strategy)
	}

	// Wait for all strategies to complete
	for i := 0; i < len(c.strategies); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	return nil
}

// DefaultExecutionClientWait returns a default wait strategy for execution clients
func DefaultExecutionClientWait() WaitStrategy {
	return NewCombinedWaitStrategy(
		NewHTTPWaitStrategy(8545).WithPath("/").WithStatusCode(405), // RPC endpoint returns 405 for GET
		NewHealthWaitStrategy(),
	)
}

// DefaultConsensusClientWait returns a default wait strategy for consensus clients
func DefaultConsensusClientWait() WaitStrategy {
	return NewCombinedWaitStrategy(
		NewHTTPWaitStrategy(5052).WithPath("/eth/v1/node/health"),
		NewHealthWaitStrategy(),
	)
}
