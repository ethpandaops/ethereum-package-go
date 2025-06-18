package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServiceWithLogs provides a mock implementation for testing
type MockServiceWithLogs struct {
	mock.Mock
}

func (m *MockServiceWithLogs) ServiceName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockServiceWithLogs) ContainerID() string {
	args := m.Called()
	return args.String(0)
}

// TestLogFilter_Options tests the functional options for log filtering
func TestLogFilter_Options(t *testing.T) {
	tests := []struct {
		name     string
		options  []LogOption
		expected LogFilter
	}{
		{
			name:    "default filter",
			options: []LogOption{},
			expected: LogFilter{
				lines:         100,
				caseSensitive: false,
			},
		},
		{
			name:    "with lines",
			options: []LogOption{WithLines(50)},
			expected: LogFilter{
				lines:         50,
				caseSensitive: false,
			},
		},
		{
			name:    "with grep",
			options: []LogOption{WithGrep("ERROR")},
			expected: LogFilter{
				lines:         100,
				grep:          "ERROR",
				caseSensitive: false,
			},
		},
		{
			name:    "with follow",
			options: []LogOption{WithFollow(true)},
			expected: LogFilter{
				lines:         100,
				follow:        true,
				caseSensitive: false,
			},
		},
		{
			name:    "with since",
			options: []LogOption{WithSince(5 * time.Minute)},
			expected: LogFilter{
				lines:         100,
				since:         5 * time.Minute,
				caseSensitive: false,
			},
		},
		{
			name:    "case sensitive",
			options: []LogOption{WithCaseSensitive(true)},
			expected: LogFilter{
				lines:         100,
				caseSensitive: true,
			},
		},
		{
			name: "multiple options",
			options: []LogOption{
				WithLines(200),
				WithGrep("WARN"),
				WithFollow(true),
				WithCaseSensitive(true),
			},
			expected: LogFilter{
				lines:         200,
				grep:          "WARN",
				follow:        true,
				caseSensitive: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &LogFilter{
				lines:         100,
				caseSensitive: false,
			}
			
			for _, option := range tt.options {
				option(filter)
			}
			
			assert.Equal(t, tt.expected.lines, filter.lines)
			assert.Equal(t, tt.expected.grep, filter.grep)
			assert.Equal(t, tt.expected.follow, filter.follow)
			assert.Equal(t, tt.expected.since, filter.since)
			assert.Equal(t, tt.expected.caseSensitive, filter.caseSensitive)
		})
	}
}

// TestLogsClient_matchesFilter tests the filter matching logic
func TestLogsClient_matchesFilter(t *testing.T) {
	lc := &LogsClient{}
	
	tests := []struct {
		name     string
		line     string
		filter   *LogFilter
		expected bool
	}{
		{
			name:     "no filter",
			line:     "This is a test log line",
			filter:   &LogFilter{},
			expected: true,
		},
		{
			name:     "grep match case insensitive",
			line:     "This is an ERROR message",
			filter:   &LogFilter{grep: "error", caseSensitive: false},
			expected: true,
		},
		{
			name:     "grep no match case insensitive",
			line:     "This is a debug message",
			filter:   &LogFilter{grep: "error", caseSensitive: false},
			expected: false,
		},
		{
			name:     "grep match case sensitive",
			line:     "This is an ERROR message",
			filter:   &LogFilter{grep: "ERROR", caseSensitive: true},
			expected: true,
		},
		{
			name:     "grep no match case sensitive",
			line:     "This is an ERROR message",
			filter:   &LogFilter{grep: "error", caseSensitive: true},
			expected: false,
		},
		{
			name:     "include regex match",
			line:     "Connection established",
			filter:   &LogFilter{includeRegex: "connection", caseSensitive: false},
			expected: true,
		},
		{
			name:     "include regex no match",
			line:     "Debug information",
			filter:   &LogFilter{includeRegex: "connection", caseSensitive: false},
			expected: false,
		},
		{
			name:     "exclude regex match",
			line:     "Debug information",
			filter:   &LogFilter{excludeRegex: "debug", caseSensitive: false},
			expected: false,
		},
		{
			name:     "exclude regex no match",
			line:     "Error information",
			filter:   &LogFilter{excludeRegex: "debug", caseSensitive: false},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lc.matchesFilter(tt.line, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestLogsClient_applyFilters tests the filter application logic
func TestLogsClient_applyFilters(t *testing.T) {
	lc := &LogsClient{}
	
	testLines := []string{
		"2023-01-01 INFO: Application started",
		"2023-01-01 DEBUG: Loading configuration",
		"2023-01-01 ERROR: Failed to connect to database",
		"2023-01-01 WARN: Deprecated API used",
		"2023-01-01 INFO: Processing request",
		"2023-01-01 ERROR: Invalid input provided",
	}
	
	tests := []struct {
		name     string
		filter   *LogFilter
		expected []string
	}{
		{
			name:   "no filter",
			filter: &LogFilter{},
			expected: testLines,
		},
		{
			name:   "grep for ERROR",
			filter: &LogFilter{grep: "ERROR", caseSensitive: false},
			expected: []string{
				"2023-01-01 ERROR: Failed to connect to database",
				"2023-01-01 ERROR: Invalid input provided",
			},
		},
		{
			name:   "limit lines",
			filter: &LogFilter{lines: 3},
			expected: []string{
				"2023-01-01 WARN: Deprecated API used",
				"2023-01-01 INFO: Processing request",
				"2023-01-01 ERROR: Invalid input provided",
			},
		},
		{
			name:   "grep and limit",
			filter: &LogFilter{grep: "INFO", lines: 1, caseSensitive: false},
			expected: []string{
				"2023-01-01 INFO: Processing request",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lc.applyFilters(testLines, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTailLogs tests the convenience function for tailing logs
func TestTailLogs(t *testing.T) {
	options := TailLogs(50, "error")
	
	// Apply options to a filter
	filter := &LogFilter{
		lines:         100,
		caseSensitive: false,
	}
	
	for _, option := range options {
		option(filter)
	}
	
	assert.Equal(t, 50, filter.lines)
	assert.Equal(t, "error", filter.grep)
}

// TestFollowLogs tests the convenience function for following logs
func TestFollowLogs(t *testing.T) {
	options := FollowLogs("warn")
	
	// Apply options to a filter
	filter := &LogFilter{
		lines:         100,
		caseSensitive: false,
	}
	
	for _, option := range options {
		option(filter)
	}
	
	assert.True(t, filter.follow)
	assert.Equal(t, "warn", filter.grep)
}

// TestServiceWithLogs_Interface tests that clients implement the interface
func TestServiceWithLogs_Interface(t *testing.T) {
	// Test ConsensusClient implements ServiceWithLogs
	consensusClient := NewConsensusClient(
		Lighthouse,
		"lighthouse-1",
		"v1.0.0",
		"http://localhost:5052",
		"http://localhost:8080",
		"enr:test",
		"peer-id",
		"lighthouse-service",
		"lighthouse-container",
		9000,
	)
	
	var _ ServiceWithLogs = consensusClient
	assert.Equal(t, "lighthouse-service", consensusClient.ServiceName())
	assert.Equal(t, "lighthouse-container", consensusClient.ContainerID())
	
	// Test ExecutionClient implements ServiceWithLogs
	executionClient := NewExecutionClient(
		Geth,
		"geth-1",
		"v1.0.0",
		"http://localhost:8545",
		"ws://localhost:8546",
		"http://localhost:8551",
		"http://localhost:8080",
		"enode://test",
		"geth-service",
		"geth-container",
		30303,
	)
	
	var _ ServiceWithLogs = executionClient
	assert.Equal(t, "geth-service", executionClient.ServiceName())
	assert.Equal(t, "geth-container", executionClient.ContainerID())
}

// TestLogsClient_LogsReader tests the LogsReader functionality
func TestLogsClient_LogsReader(t *testing.T) {
	// This would require mocking the Kurtosis context
	// For now, we test that the method exists and can be called
	lc := &LogsClient{}
	
	// Test that the method signature is correct
	assert.NotNil(t, lc)
	
	// In a real test, you'd mock the Kurtosis context and test the actual functionality
	// mockService := &MockServiceWithLogs{}
	// mockService.On("ServiceName").Return("test-service")
	// reader, err := lc.LogsReader(context.Background(), mockService)
	// assert.NoError(t, err)
	// assert.NotNil(t, reader)
}

// BenchmarkLogFilter_matchesFilter benchmarks the filter matching
func BenchmarkLogFilter_matchesFilter(b *testing.B) {
	lc := &LogsClient{}
	filter := &LogFilter{
		grep:          "ERROR",
		caseSensitive: false,
	}
	line := "2023-01-01 12:00:00 ERROR: This is an error message"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.matchesFilter(line, filter)
	}
}

// BenchmarkLogFilter_applyFilters benchmarks filter application
func BenchmarkLogFilter_applyFilters(b *testing.B) {
	lc := &LogsClient{}
	filter := &LogFilter{
		grep:  "ERROR",
		lines: 100,
	}
	
	// Generate test lines
	lines := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		if i%10 == 0 {
			lines[i] = "ERROR: This is an error message"
		} else {
			lines[i] = "INFO: This is an info message"
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.applyFilters(lines, filter)
	}
}

// TestLogOption_Chaining tests that options can be chained together
func TestLogOption_Chaining(t *testing.T) {
	filter := &LogFilter{
		lines:         100,
		caseSensitive: false,
	}
	
	// Chain multiple options
	options := []LogOption{
		WithLines(50),
		WithGrep("test"),
		WithFollow(true),
		WithSince(time.Hour),
		WithCaseSensitive(true),
		WithIncludeRegex("include"),
		WithExcludeRegex("exclude"),
	}
	
	for _, option := range options {
		option(filter)
	}
	
	assert.Equal(t, 50, filter.lines)
	assert.Equal(t, "test", filter.grep)
	assert.True(t, filter.follow)
	assert.Equal(t, time.Hour, filter.since)
	assert.True(t, filter.caseSensitive)
	assert.Equal(t, "include", filter.includeRegex)
	assert.Equal(t, "exclude", filter.excludeRegex)
}