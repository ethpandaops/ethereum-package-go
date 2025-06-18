package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthChecker provides health checking capabilities for services
type HealthChecker struct {
	httpClient *http.Client
	checks     map[string]HealthCheck
	mu         sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		checks: make(map[string]HealthCheck),
	}
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	Name         string
	Type         HealthCheckType
	URL          string
	Method       string
	Interval     time.Duration
	Timeout      time.Duration
	SuccessCodes []int
	CheckFunc    func(ctx context.Context) error
}

// HealthCheckType represents the type of health check
type HealthCheckType string

const (
	HealthCheckHTTP   HealthCheckType = "http"
	HealthCheckTCP    HealthCheckType = "tcp"
	HealthCheckCustom HealthCheckType = "custom"
)

// ServiceHealthStatus represents the health status of a service
type ServiceHealthStatus struct {
	Name      string        `json:"name"`
	Status    ServiceStatus `json:"status"`
	Message   string        `json:"message,omitempty"`
	LastCheck time.Time     `json:"last_check"`
	Uptime    time.Duration `json:"uptime,omitempty"`
	Details   interface{}   `json:"details,omitempty"`
}

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	StatusHealthy   ServiceStatus = "healthy"
	StatusUnhealthy ServiceStatus = "unhealthy"
	StatusDegraded  ServiceStatus = "degraded"
	StatusUnknown   ServiceStatus = "unknown"
)

// RegisterCheck registers a new health check
func (h *HealthChecker) RegisterCheck(check HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Set defaults
	if check.Method == "" {
		check.Method = "GET"
	}
	if check.Timeout == 0 {
		check.Timeout = 5 * time.Second
	}
	if len(check.SuccessCodes) == 0 {
		check.SuccessCodes = []int{http.StatusOK}
	}

	h.checks[check.Name] = check
}

// RegisterHTTPCheck registers an HTTP health check
func (h *HealthChecker) RegisterHTTPCheck(name, url string) {
	h.RegisterCheck(HealthCheck{
		Name:         name,
		Type:         HealthCheckHTTP,
		URL:          url,
		Method:       "GET",
		SuccessCodes: []int{http.StatusOK},
	})
}

// CheckHealth performs a health check for a specific service
func (h *HealthChecker) CheckHealth(ctx context.Context, name string) (*ServiceHealthStatus, error) {
	h.mu.RLock()
	check, exists := h.checks[name]
	h.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("health check %s not found", name)
	}

	status := &ServiceHealthStatus{
		Name:      name,
		LastCheck: time.Now(),
		Status:    StatusUnknown,
	}

	switch check.Type {
	case HealthCheckHTTP:
		err := h.performHTTPCheck(ctx, check)
		if err != nil {
			status.Status = StatusUnhealthy
			status.Message = err.Error()
		} else {
			status.Status = StatusHealthy
		}

	case HealthCheckCustom:
		if check.CheckFunc != nil {
			err := check.CheckFunc(ctx)
			if err != nil {
				status.Status = StatusUnhealthy
				status.Message = err.Error()
			} else {
				status.Status = StatusHealthy
			}
		}

	default:
		return nil, fmt.Errorf("unsupported health check type: %s", check.Type)
	}

	return status, nil
}

// CheckAllHealth performs health checks for all registered services
func (h *HealthChecker) CheckAllHealth(ctx context.Context) map[string]*ServiceHealthStatus {
	h.mu.RLock()
	checkNames := make([]string, 0, len(h.checks))
	for name := range h.checks {
		checkNames = append(checkNames, name)
	}
	h.mu.RUnlock()

	results := make(map[string]*ServiceHealthStatus)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, name := range checkNames {
		wg.Add(1)
		go func(checkName string) {
			defer wg.Done()

			status, err := h.CheckHealth(ctx, checkName)
			if err != nil {
				status = &ServiceHealthStatus{
					Name:      checkName,
					Status:    StatusUnknown,
					Message:   err.Error(),
					LastCheck: time.Now(),
				}
			}

			mu.Lock()
			results[checkName] = status
			mu.Unlock()
		}(name)
	}

	wg.Wait()
	return results
}

// performHTTPCheck performs an HTTP health check
func (h *HealthChecker) performHTTPCheck(ctx context.Context, check HealthCheck) error {
	req, err := http.NewRequestWithContext(ctx, check.Method, check.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if status code is in success codes
	for _, code := range check.SuccessCodes {
		if resp.StatusCode == code {
			return nil
		}
	}

	return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
}

// AggregateHealth aggregates multiple health statuses into an overall status
func (h *HealthChecker) AggregateHealth(statuses map[string]*ServiceHealthStatus) *OverallHealth {
	overall := &OverallHealth{
		Status:    StatusHealthy,
		Services:  statuses,
		Timestamp: time.Now(),
	}

	healthyCount := 0
	unhealthyCount := 0
	degradedCount := 0

	for _, status := range statuses {
		switch status.Status {
		case StatusHealthy:
			healthyCount++
		case StatusUnhealthy:
			unhealthyCount++
		case StatusDegraded:
			degradedCount++
		}
	}

	overall.HealthyServices = healthyCount
	overall.UnhealthyServices = unhealthyCount
	overall.DegradedServices = degradedCount
	overall.TotalServices = len(statuses)

	// Determine overall status
	if unhealthyCount > 0 {
		overall.Status = StatusUnhealthy
		if unhealthyCount < len(statuses) {
			overall.Status = StatusDegraded
		}
	} else if degradedCount > 0 {
		overall.Status = StatusDegraded
	}

	return overall
}

// StartPeriodicChecks starts periodic health checks for all registered services
func (h *HealthChecker) StartPeriodicChecks(ctx context.Context, interval time.Duration) <-chan map[string]*ServiceHealthStatus {
	results := make(chan map[string]*ServiceHealthStatus)

	go func() {
		defer close(results)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Initial check
		select {
		case results <- h.CheckAllHealth(ctx):
		case <-ctx.Done():
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case results <- h.CheckAllHealth(ctx):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return results
}

// OverallHealth represents the overall health of all services
type OverallHealth struct {
	Status            ServiceStatus                   `json:"status"`
	Services          map[string]*ServiceHealthStatus `json:"services"`
	HealthyServices   int                             `json:"healthy_services"`
	UnhealthyServices int                             `json:"unhealthy_services"`
	DegradedServices  int                             `json:"degraded_services"`
	TotalServices     int                             `json:"total_services"`
	Timestamp         time.Time                       `json:"timestamp"`
}

// CreateServiceHealthCheck creates a health check for common service types
func CreateServiceHealthCheck(serviceType, name, url string) HealthCheck {
	switch serviceType {
	case "prometheus":
		return HealthCheck{
			Name:         name,
			Type:         HealthCheckHTTP,
			URL:          url + "/-/healthy",
			SuccessCodes: []int{http.StatusOK},
		}
	case "grafana":
		return HealthCheck{
			Name:         name,
			Type:         HealthCheckHTTP,
			URL:          url + "/api/health",
			SuccessCodes: []int{http.StatusOK},
		}
	case "execution":
		return HealthCheck{
			Name: name,
			Type: HealthCheckCustom,
			CheckFunc: func(ctx context.Context) error {
				// Custom check for execution client
				// This would typically check if the client is synced
				return nil
			},
		}
	case "consensus":
		return HealthCheck{
			Name:         name,
			Type:         HealthCheckHTTP,
			URL:          url + "/eth/v1/node/health",
			SuccessCodes: []int{http.StatusOK, http.StatusPartialContent},
		}
	default:
		return HealthCheck{
			Name:         name,
			Type:         HealthCheckHTTP,
			URL:          url,
			SuccessCodes: []int{http.StatusOK},
		}
	}
}
