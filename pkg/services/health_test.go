package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthChecker_CheckHealth(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serviceName  string
		expectedStatus ServiceStatus
		expectedMsg  string
	}{
		{
			name:        "healthy service",
			serviceName: "test-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			},
			expectedStatus: StatusHealthy,
		},
		{
			name:        "unhealthy service",
			serviceName: "test-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Service Unavailable"))
			},
			expectedStatus: StatusUnhealthy,
			expectedMsg:    "unhealthy status code: 503",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			checker := NewHealthChecker()
			checker.RegisterHTTPCheck(tt.serviceName, server.URL)
			ctx := context.Background()

			status, err := checker.CheckHealth(ctx, tt.serviceName)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, status.Status)
			if tt.expectedMsg != "" {
				assert.Contains(t, status.Message, tt.expectedMsg)
			}
		})
	}
}

func TestHealthChecker_CheckAllHealth(t *testing.T) {
	// Create test servers
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthyServer.Close()

	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthyServer.Close()

	checker := NewHealthChecker()
	checker.RegisterHTTPCheck("healthy-service", healthyServer.URL)
	checker.RegisterHTTPCheck("unhealthy-service", unhealthyServer.URL)
	ctx := context.Background()

	results := checker.CheckAllHealth(ctx)
	
	assert.Len(t, results, 2)
	assert.Equal(t, StatusHealthy, results["healthy-service"].Status)
	assert.Equal(t, StatusUnhealthy, results["unhealthy-service"].Status)
}

func TestHealthChecker_AggregateHealth(t *testing.T) {
	checker := NewHealthChecker()
	
	statuses := map[string]*ServiceHealthStatus{
		"service1": {
			Name:   "service1",
			Status: StatusHealthy,
		},
		"service2": {
			Name:   "service2",
			Status: StatusUnhealthy,
			Message: "Connection failed",
		},
		"service3": {
			Name:   "service3",
			Status: StatusDegraded,
		},
	}
	
	overall := checker.AggregateHealth(statuses)
	
	assert.Equal(t, StatusDegraded, overall.Status)
	assert.Equal(t, 1, overall.HealthyServices)
	assert.Equal(t, 1, overall.UnhealthyServices)
	assert.Equal(t, 1, overall.DegradedServices)
	assert.Equal(t, 3, overall.TotalServices)
}

func TestCreateServiceHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		serviceType  string
		serviceName  string
		url          string
		expectedURL  string
		expectedType HealthCheckType
	}{
		{
			name:         "prometheus health check",
			serviceType:  "prometheus",
			serviceName:  "prom-1",
			url:          "http://localhost:9090",
			expectedURL:  "http://localhost:9090/-/healthy",
			expectedType: HealthCheckHTTP,
		},
		{
			name:         "grafana health check",
			serviceType:  "grafana",
			serviceName:  "graf-1",
			url:          "http://localhost:3000",
			expectedURL:  "http://localhost:3000/api/health",
			expectedType: HealthCheckHTTP,
		},
		{
			name:         "consensus health check",
			serviceType:  "consensus",
			serviceName:  "cl-1",
			url:          "http://localhost:5052",
			expectedURL:  "http://localhost:5052/eth/v1/node/health",
			expectedType: HealthCheckHTTP,
		},
		{
			name:         "execution health check",
			serviceType:  "execution",
			serviceName:  "el-1",
			url:          "http://localhost:8545",
			expectedType: HealthCheckCustom,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CreateServiceHealthCheck(tt.serviceType, tt.serviceName, tt.url)
			assert.Equal(t, tt.serviceName, check.Name)
			assert.Equal(t, tt.expectedType, check.Type)
			if tt.expectedURL != "" {
				assert.Equal(t, tt.expectedURL, check.URL)
			}
		})
	}
}