package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ApacheConfigClient provides access to the Apache config server that hosts network configuration files
type ApacheConfigClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewApacheConfigClient creates a new Apache config client
func NewApacheConfigClient(baseURL string) *ApacheConfigClient {
	return &ApacheConfigClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// URL returns the base URL of the Apache config server
func (a *ApacheConfigClient) URL() string {
	return a.baseURL
}

// GenesisSSZURL returns the URL for downloading the genesis.ssz file
func (a *ApacheConfigClient) GenesisSSZURL() string {
	return a.baseURL + "/network-configs/genesis.ssz"
}

// ConfigYAMLURL returns the URL for downloading the config.yaml file
func (a *ApacheConfigClient) ConfigYAMLURL() string {
	return a.baseURL + "/network-configs/config.yaml"
}

// BootnodesYAMLURL returns the URL for downloading the boot_enr.yaml file
func (a *ApacheConfigClient) BootnodesYAMLURL() string {
	return a.baseURL + "/network-configs/boot_enr.yaml"
}

// DepositContractBlockURL returns the URL for downloading the deposit_contract_block.txt file
func (a *ApacheConfigClient) DepositContractBlockURL() string {
	return a.baseURL + "/network-configs/deposit_contract_block.txt"
}

// DownloadGenesisSSZ downloads the genesis.ssz file
func (a *ApacheConfigClient) DownloadGenesisSSZ(ctx context.Context) ([]byte, error) {
	return a.downloadFile(ctx, "/network-configs/genesis.ssz")
}

// DownloadConfigYAML downloads the config.yaml file
func (a *ApacheConfigClient) DownloadConfigYAML(ctx context.Context) ([]byte, error) {
	return a.downloadFile(ctx, "/network-configs/config.yaml")
}

// DownloadBootnodesYAML downloads the boot_enr.yaml file
func (a *ApacheConfigClient) DownloadBootnodesYAML(ctx context.Context) ([]byte, error) {
	return a.downloadFile(ctx, "/network-configs/boot_enr.yaml")
}

// DownloadDepositContractBlock downloads the deposit_contract_block.txt file
func (a *ApacheConfigClient) DownloadDepositContractBlock(ctx context.Context) ([]byte, error) {
	return a.downloadFile(ctx, "/network-configs/deposit_contract_block.txt")
}

// GetGenesisSSZAsString downloads and returns the genesis.ssz file as a base64 string
func (a *ApacheConfigClient) GetGenesisSSZAsString(ctx context.Context) (string, error) {
	data, err := a.DownloadGenesisSSZ(ctx)
	if err != nil {
		return "", err
	}
	// For binary files like SSZ, we might want to return as base64 or hex
	// For now, return as string (though this might not be ideal for binary data)
	return string(data), nil
}

// GetConfigYAMLAsString downloads and returns the config.yaml file as a string
func (a *ApacheConfigClient) GetConfigYAMLAsString(ctx context.Context) (string, error) {
	data, err := a.DownloadConfigYAML(ctx)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetBootnodesYAMLAsString downloads and returns the boot_enr.yaml file as a string
func (a *ApacheConfigClient) GetBootnodesYAMLAsString(ctx context.Context) (string, error) {
	data, err := a.DownloadBootnodesYAML(ctx)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetDepositContractBlockAsString downloads and returns the deposit_contract_block.txt file as a string
func (a *ApacheConfigClient) GetDepositContractBlockAsString(ctx context.Context) (string, error) {
	data, err := a.DownloadDepositContractBlock(ctx)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListAvailableFiles lists all available files in the network-configs directory
func (a *ApacheConfigClient) ListAvailableFiles(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/network-configs/", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Simple HTML parsing to extract file names
	// This is a basic implementation - in a real scenario you might want to use a proper HTML parser
	html := string(body)
	var files []string
	
	// Look for href patterns in the HTML
	lines := strings.Split(html, "\n")
	for _, line := range lines {
		if strings.Contains(line, "href=") && !strings.Contains(line, "../") {
			// Extract filename from href
			start := strings.Index(line, "href=\"")
			if start != -1 {
				start += 6
				end := strings.Index(line[start:], "\"")
				if end != -1 {
					filename := line[start : start+end]
					if filename != "" && !strings.HasSuffix(filename, "/") {
						files = append(files, filename)
					}
				}
			}
		}
	}

	return files, nil
}

// IsHealthy checks if the Apache config server is healthy and reachable
func (a *ApacheConfigClient) IsHealthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", a.baseURL+"/network-configs/", nil)
	if err != nil {
		return false
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// CheckFileExists checks if a specific file exists on the server
func (a *ApacheConfigClient) CheckFileExists(ctx context.Context, filename string) (bool, error) {
	url := fmt.Sprintf("%s/network-configs/%s", a.baseURL, filename)
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

// downloadFile downloads a file from the config server
func (a *ApacheConfigClient) downloadFile(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download file, status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// GetFileInfo returns information about a file (size, last modified, etc.)
func (a *ApacheConfigClient) GetFileInfo(ctx context.Context, filename string) (*FileInfo, error) {
	url := fmt.Sprintf("%s/network-configs/%s", a.baseURL, filename)
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file not found or request failed, status: %d", resp.StatusCode)
	}

	info := &FileInfo{
		Name:         filename,
		URL:          url,
		ContentType:  resp.Header.Get("Content-Type"),
		Size:         resp.ContentLength,
		LastModified: resp.Header.Get("Last-Modified"),
		ETag:         resp.Header.Get("ETag"),
	}

	return info, nil
}

// FileInfo represents information about a file on the config server
type FileInfo struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	ETag         string `json:"etag"`
}