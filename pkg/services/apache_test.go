package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApacheConfigClient_URLs(t *testing.T) {
	baseURL := "http://localhost:8080"
	client := NewApacheConfigClient(baseURL)

	assert.Equal(t, baseURL, client.URL())
	assert.Equal(t, baseURL+"/network-configs/genesis.ssz", client.GenesisSSZURL())
	assert.Equal(t, baseURL+"/network-configs/config.yaml", client.ConfigYAMLURL())
	assert.Equal(t, baseURL+"/network-configs/boot_enr.yaml", client.BootnodesYAMLURL())
	assert.Equal(t, baseURL+"/network-configs/deposit_contract_block.txt", client.DepositContractBlockURL())
}

func TestApacheConfigClient_URLWithTrailingSlash(t *testing.T) {
	baseURL := "http://localhost:8080/"
	client := NewApacheConfigClient(baseURL)

	// Should trim trailing slash
	assert.Equal(t, "http://localhost:8080", client.URL())
	assert.Equal(t, "http://localhost:8080/network-configs/genesis.ssz", client.GenesisSSZURL())
}

func TestApacheConfigClient_Downloads(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/network-configs/genesis.ssz":
			w.Write([]byte("genesis-data"))
		case "/network-configs/config.yaml":
			w.Write([]byte("config: test"))
		case "/network-configs/boot_enr.yaml":
			w.Write([]byte("enr: test"))
		case "/network-configs/deposit_contract_block.txt":
			w.Write([]byte("12345"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewApacheConfigClient(server.URL)
	ctx := context.Background()

	// Test genesis download
	genesis, err := client.DownloadGenesisSSZ(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("genesis-data"), genesis)

	// Test config download
	config, err := client.DownloadConfigYAML(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("config: test"), config)

	// Test bootnodes download
	bootnodes, err := client.DownloadBootnodesYAML(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("enr: test"), bootnodes)

	// Test deposit contract block download
	block, err := client.DownloadDepositContractBlock(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("12345"), block)
}

func TestApacheConfigClient_DownloadAsString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/network-configs/config.yaml":
			w.Write([]byte("config: test"))
		case "/network-configs/boot_enr.yaml":
			w.Write([]byte("enr: test"))
		case "/network-configs/deposit_contract_block.txt":
			w.Write([]byte("12345"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewApacheConfigClient(server.URL)
	ctx := context.Background()

	// Test config as string
	config, err := client.GetConfigYAMLAsString(ctx)
	require.NoError(t, err)
	assert.Equal(t, "config: test", config)

	// Test bootnodes as string
	bootnodes, err := client.GetBootnodesYAMLAsString(ctx)
	require.NoError(t, err)
	assert.Equal(t, "enr: test", bootnodes)

	// Test deposit contract block as string
	block, err := client.GetDepositContractBlockAsString(ctx)
	require.NoError(t, err)
	assert.Equal(t, "12345", block)
}

func TestApacheConfigClient_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		expected bool
	}{
		{
			name: "healthy server",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "HEAD", r.Method)
				assert.Equal(t, "/network-configs/", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			expected: true,
		},
		{
			name: "unhealthy server",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewApacheConfigClient(server.URL)
			ctx := context.Background()

			healthy := client.IsHealthy(ctx)
			assert.Equal(t, tt.expected, healthy)
		})
	}
}

func TestApacheConfigClient_CheckFileExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "HEAD", r.Method)
		switch r.URL.Path {
		case "/network-configs/genesis.ssz":
			w.WriteHeader(http.StatusOK)
		case "/network-configs/missing.txt":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client := NewApacheConfigClient(server.URL)
	ctx := context.Background()

	// Test existing file
	exists, err := client.CheckFileExists(ctx, "genesis.ssz")
	require.NoError(t, err)
	assert.True(t, exists)

	// Test missing file
	exists, err = client.CheckFileExists(ctx, "missing.txt")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test server error
	_, err = client.CheckFileExists(ctx, "error.txt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}

func TestApacheConfigClient_GetFileInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "HEAD", r.Method)
		if r.URL.Path == "/network-configs/test.yaml" {
			w.Header().Set("Content-Type", "text/yaml")
			w.Header().Set("Content-Length", "1234")
			w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
			w.Header().Set("ETag", "\"abc123\"")
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewApacheConfigClient(server.URL)
	ctx := context.Background()

	// Test file info
	info, err := client.GetFileInfo(ctx, "test.yaml")
	require.NoError(t, err)
	assert.Equal(t, "test.yaml", info.Name)
	assert.Equal(t, server.URL+"/network-configs/test.yaml", info.URL)
	assert.Equal(t, "text/yaml", info.ContentType)
	assert.Equal(t, int64(1234), info.Size)
	assert.Equal(t, "Mon, 02 Jan 2006 15:04:05 GMT", info.LastModified)
	assert.Equal(t, "\"abc123\"", info.ETag)

	// Test missing file
	_, err = client.GetFileInfo(ctx, "missing.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}
