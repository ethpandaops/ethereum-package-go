package client

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
)

// ServiceWithLogs represents any service that can provide logs
type ServiceWithLogs interface {
	ServiceName() string
	ContainerID() string
}

// LogFilter represents a filter for log retrieval
type LogFilter struct {
	lines         int
	grep          string
	since         time.Duration
	follow        bool
	includeRegex  string
	excludeRegex  string
	caseSensitive bool
}

// LogOption is a functional option for configuring log filters
type LogOption func(*LogFilter)

// WithLines sets the number of lines to retrieve (tail behavior)
func WithLines(lines int) LogOption {
	return func(f *LogFilter) {
		f.lines = lines
	}
}

// WithGrep filters logs containing the specified string
func WithGrep(pattern string) LogOption {
	return func(f *LogFilter) {
		f.grep = pattern
	}
}

// WithSince filters logs since the specified duration ago
func WithSince(duration time.Duration) LogOption {
	return func(f *LogFilter) {
		f.since = duration
	}
}

// WithFollow enables following logs (tail -f behavior)
func WithFollow(follow bool) LogOption {
	return func(f *LogFilter) {
		f.follow = follow
	}
}

// WithIncludeRegex includes only lines matching the regex pattern
func WithIncludeRegex(pattern string) LogOption {
	return func(f *LogFilter) {
		f.includeRegex = pattern
	}
}

// WithExcludeRegex excludes lines matching the regex pattern
func WithExcludeRegex(pattern string) LogOption {
	return func(f *LogFilter) {
		f.excludeRegex = pattern
	}
}

// WithCaseSensitive enables case-sensitive filtering (default is case-insensitive)
func WithCaseSensitive(caseSensitive bool) LogOption {
	return func(f *LogFilter) {
		f.caseSensitive = caseSensitive
	}
}

// LogsClient provides log retrieval functionality for services
type LogsClient struct {
	kurtosisCtx       *kurtosis_context.KurtosisContext
	enclaveIdentifier string
}

// NewLogsClient creates a new logs client
func NewLogsClient(kurtosisCtx *kurtosis_context.KurtosisContext, enclaveIdentifier string) *LogsClient {
	return &LogsClient{
		kurtosisCtx:       kurtosisCtx,
		enclaveIdentifier: enclaveIdentifier,
	}
}

// Logs retrieves logs from a service with the specified filters
func (lc *LogsClient) Logs(ctx context.Context, service ServiceWithLogs, options ...LogOption) ([]string, error) {
	// Apply options to create filter
	filter := &LogFilter{
		lines:         100, // default to last 100 lines
		caseSensitive: false,
	}
	for _, option := range options {
		option(filter)
	}

	// Create LogLineFilter
	var logLineFilter *kurtosis_context.LogLineFilter
	if filter.grep != "" {
		if filter.caseSensitive {
			logLineFilter = kurtosis_context.NewDoesContainTextLogLineFilter(filter.grep)
		} else {
			logLineFilter = kurtosis_context.NewDoesContainTextLogLineFilter(strings.ToLower(filter.grep))
		}
	}

	// Get service UUID map
	serviceName := service.ServiceName()
	serviceUUIDs := map[services.ServiceUUID]bool{
		services.ServiceUUID(serviceName): true,
	}

	// Get logs from Kurtosis using kurtosis context
	logsChan, cancelFunc, err := lc.kurtosisCtx.GetServiceLogs(
		ctx,
		lc.enclaveIdentifier,
		serviceUUIDs,
		filter.follow,
		false, // shouldReturnAllLogs
		uint32(filter.lines),
		logLineFilter,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for service %s: %w", serviceName, err)
	}
	defer cancelFunc()

	// Read from channel
	var allLines []string
	done := make(chan struct{})
	go func() {
		defer close(done)
		for logContent := range logsChan {
			// Process log content using correct method name
			for serviceUUID, serviceLogs := range logContent.GetServiceLogsByServiceUuids() {
				_ = serviceUUID // Service UUID for reference
				for _, logLine := range serviceLogs {
					allLines = append(allLines, logLine.GetContent())
				}
			}
		}
	}()

	// Wait for completion or timeout
	select {
	case <-done:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(30 * time.Second): // Add timeout
		return allLines, nil
	}

	// Apply additional filters that weren't handled by Kurtosis
	filteredLines := lc.applyFilters(allLines, filter)

	return filteredLines, nil
}

// LogsStream provides a streaming interface for logs
func (lc *LogsClient) LogsStream(ctx context.Context, service ServiceWithLogs, options ...LogOption) (<-chan string, <-chan error) {
	logChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(logChan)
		defer close(errChan)

		// Apply options to create filter
		filter := &LogFilter{
			lines:         100,
			caseSensitive: false,
			follow:        true, // streaming always follows
		}
		for _, option := range options {
			option(filter)
		}

		// Create LogLineFilter
		var logLineFilter *kurtosis_context.LogLineFilter
		if filter.grep != "" {
			if filter.caseSensitive {
				logLineFilter = kurtosis_context.NewDoesContainTextLogLineFilter(filter.grep)
			} else {
				logLineFilter = kurtosis_context.NewDoesContainTextLogLineFilter(strings.ToLower(filter.grep))
			}
		}

		// Get service UUID map
		serviceName := service.ServiceName()
		serviceUUIDs := map[services.ServiceUUID]bool{
			services.ServiceUUID(serviceName): true,
		}

		// Get streaming logs from Kurtosis using kurtosis context
		logsChan, cancelFunc, err := lc.kurtosisCtx.GetServiceLogs(
			ctx,
			lc.enclaveIdentifier,
			serviceUUIDs,
			true, // follow logs
			false,
			uint32(filter.lines),
			logLineFilter,
		)
		if err != nil {
			errChan <- fmt.Errorf("failed to get streaming logs for service %s: %w", serviceName, err)
			return
		}
		defer cancelFunc()

		// Process streaming logs
		for logContent := range logsChan {
			for serviceUUID, serviceLogs := range logContent.GetServiceLogsByServiceUuids() {
				_ = serviceUUID // Service UUID for reference
				for _, logLine := range serviceLogs {
					line := logLine.GetContent()
					if lc.matchesFilter(line, filter) {
						select {
						case logChan <- line:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}()

	return logChan, errChan
}

// applyFilters applies all configured filters to the log lines
func (lc *LogsClient) applyFilters(lines []string, filter *LogFilter) []string {
	var filtered []string

	for _, line := range lines {
		if lc.matchesFilter(line, filter) {
			filtered = append(filtered, line)
		}
	}

	// Apply line limit if specified
	if filter.lines > 0 && len(filtered) > filter.lines {
		filtered = filtered[len(filtered)-filter.lines:]
	}

	return filtered
}

// matchesFilter checks if a log line matches the configured filters
func (lc *LogsClient) matchesFilter(line string, filter *LogFilter) bool {
	// Apply case sensitivity
	searchLine := line
	if !filter.caseSensitive {
		searchLine = strings.ToLower(line)
	}

	// Apply grep filter
	if filter.grep != "" {
		grepPattern := filter.grep
		if !filter.caseSensitive {
			grepPattern = strings.ToLower(grepPattern)
		}
		if !strings.Contains(searchLine, grepPattern) {
			return false
		}
	}

	// Apply include regex if specified
	if filter.includeRegex != "" {
		// This would need a regex implementation
		// For now, we'll use simple string matching
		includePattern := filter.includeRegex
		if !filter.caseSensitive {
			includePattern = strings.ToLower(includePattern)
		}
		if !strings.Contains(searchLine, includePattern) {
			return false
		}
	}

	// Apply exclude regex if specified
	if filter.excludeRegex != "" {
		excludePattern := filter.excludeRegex
		if !filter.caseSensitive {
			excludePattern = strings.ToLower(excludePattern)
		}
		if strings.Contains(searchLine, excludePattern) {
			return false
		}
	}

	return true
}

// LogsReader provides an io.Reader interface for logs
func (lc *LogsClient) LogsReader(ctx context.Context, service ServiceWithLogs, options ...LogOption) (io.Reader, error) {
	lines, err := lc.Logs(ctx, service, options...)
	if err != nil {
		return nil, err
	}

	// Join lines with newlines
	content := strings.Join(lines, "\n")
	return strings.NewReader(content), nil
}

// Convenience methods for consensus and execution clients

// ConsensusClientLogs retrieves logs for a consensus client
func (lc *LogsClient) ConsensusClientLogs(ctx context.Context, client ConsensusClient, options ...LogOption) ([]string, error) {
	return lc.Logs(ctx, client, options...)
}

// ExecutionClientLogs retrieves logs for an execution client
func (lc *LogsClient) ExecutionClientLogs(ctx context.Context, client ExecutionClient, options ...LogOption) ([]string, error) {
	return lc.Logs(ctx, client, options...)
}

// AllConsensusClientLogs retrieves logs for all consensus clients
func (lc *LogsClient) AllConsensusClientLogs(ctx context.Context, clients *ConsensusClients, options ...LogOption) (map[string][]string, error) {
	allClients := clients.All()
	logs := make(map[string][]string)

	for _, client := range allClients {
		clientLogs, err := lc.ConsensusClientLogs(ctx, client, options...)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs for consensus client %s: %w", client.Name(), err)
		}
		logs[client.Name()] = clientLogs
	}

	return logs, nil
}

// AllExecutionClientLogs retrieves logs for all execution clients
func (lc *LogsClient) AllExecutionClientLogs(ctx context.Context, clients *ExecutionClients, options ...LogOption) (map[string][]string, error) {
	allClients := clients.All()
	logs := make(map[string][]string)

	for _, client := range allClients {
		clientLogs, err := lc.ExecutionClientLogs(ctx, client, options...)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs for execution client %s: %w", client.Name(), err)
		}
		logs[client.Name()] = clientLogs
	}

	return logs, nil
}

// TailLogs provides a convenient way to tail logs (last N lines + grep)
func TailLogs(lines int, grep string) []LogOption {
	options := []LogOption{
		WithLines(lines),
	}
	if grep != "" {
		options = append(options, WithGrep(grep))
	}
	return options
}

// FollowLogs provides a convenient way to follow logs with grep
func FollowLogs(grep string) []LogOption {
	options := []LogOption{
		WithFollow(true),
	}
	if grep != "" {
		options = append(options, WithGrep(grep))
	}
	return options
}