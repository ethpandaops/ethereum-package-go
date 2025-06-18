package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConsensusClient_FetchPeerID_HTTP tests the HTTP-based peer ID fetching
func TestConsensusClient_FetchPeerID_HTTP(t *testing.T) {
	tests := []struct {
		name           string
		setupServer    func() *httptest.Server
		client         func(serverURL string) *ConsensusClientImpl
		expectedPeerID string
		expectError    bool
		errorContains  string
	}{
		{
			name: "successful peer ID fetch",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/eth/v1/node/identity", r.URL.Path)
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, "application/json", r.Header.Get("Accept"))

					response := NodeIdentityResponse{
						Data: struct {
							PeerID             string   `json:"peer_id"`
							ENR                string   `json:"enr"`
							P2PAddresses       []string `json:"p2p_addresses"`
							DiscoveryAddresses []string `json:"discovery_addresses"`
							Metadata           struct {
								SeqNumber         string `json:"seq_number"`
								Attnets           string `json:"attnets"`
								SyncCommitteeNets string `json:"syncnets,omitempty"`
							} `json:"metadata"`
						}{
							PeerID: "16Uiu2HAkuVKJJuNnFVhfVjrw1nXJt6c2d1NcmAZqYLbA4Km7KLRZ",
							ENR:    "enr:-MS4QBU9k...",
							P2PAddresses: []string{
								"/ip4/192.168.1.100/tcp/9000",
								"/ip6/::1/tcp/9000",
							},
							DiscoveryAddresses: []string{
								"/ip4/192.168.1.100/udp/9000",
							},
						},
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(response)
				}))
			},
			client: func(serverURL string) *ConsensusClientImpl {
				return NewConsensusClient(
					Lighthouse,
					"lighthouse-1",
					"v1.0.0",
					serverURL, // beacon API URL
					"http://localhost:8080",
					"enr:test",
					"stored-peer-id",
					"lighthouse-service",
					"container-123",
					9000,
				)
			},
			expectedPeerID: "16Uiu2HAkuVKJJuNnFVhfVjrw1nXJt6c2d1NcmAZqYLbA4Km7KLRZ",
			expectError:    false,
		},
		{
			name: "server returns 404",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Not found"))
				}))
			},
			client: func(serverURL string) *ConsensusClientImpl {
				return NewConsensusClient(
					Lighthouse,
					"lighthouse-1",
					"v1.0.0",
					serverURL,
					"http://localhost:8080",
					"enr:test",
					"stored-peer-id",
					"lighthouse-service",
					"container-123",
					9000,
				)
			},
			expectError:   true,
			errorContains: "beacon API returned status 404",
		},
		{
			name: "server returns invalid JSON",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("invalid json"))
				}))
			},
			client: func(serverURL string) *ConsensusClientImpl {
				return NewConsensusClient(
					Lighthouse,
					"lighthouse-1",
					"v1.0.0",
					serverURL,
					"http://localhost:8080",
					"enr:test",
					"stored-peer-id",
					"lighthouse-service",
					"container-123",
					9000,
				)
			},
			expectError:   true,
			errorContains: "failed to decode response",
		},
		{
			name: "empty peer ID in response",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := NodeIdentityResponse{
						Data: struct {
							PeerID             string   `json:"peer_id"`
							ENR                string   `json:"enr"`
							P2PAddresses       []string `json:"p2p_addresses"`
							DiscoveryAddresses []string `json:"discovery_addresses"`
							Metadata           struct {
								SeqNumber         string `json:"seq_number"`
								Attnets           string `json:"attnets"`
								SyncCommitteeNets string `json:"syncnets,omitempty"`
							} `json:"metadata"`
						}{
							PeerID: "", // empty peer ID
							ENR:    "enr:-MS4QBU9k...",
						},
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(response)
				}))
			},
			client: func(serverURL string) *ConsensusClientImpl {
				return NewConsensusClient(
					Lighthouse,
					"lighthouse-1",
					"v1.0.0",
					serverURL,
					"http://localhost:8080",
					"enr:test",
					"stored-peer-id",
					"lighthouse-service",
					"container-123",
					9000,
				)
			},
			expectError:   true,
			errorContains: "peer_id is empty in response",
		},
		{
			name: "empty beacon URL",
			setupServer: func() *httptest.Server {
				return nil // No server needed for this test
			},
			client: func(serverURL string) *ConsensusClientImpl {
				return NewConsensusClient(
					Lighthouse,
					"lighthouse-1",
					"v1.0.0",
					"", // empty beacon URL
					"http://localhost:8080",
					"enr:test",
					"stored-peer-id",
					"lighthouse-service",
					"container-123",
					9000,
				)
			},
			expectError:   true,
			errorContains: "beacon API URL is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			var client *ConsensusClientImpl
			var server *httptest.Server

			if tt.setupServer != nil {
				server = tt.setupServer()
				if server != nil {
					defer server.Close()
					client = tt.client(server.URL)
				} else {
					client = tt.client("")
				}
			} else {
				client = tt.client("")
			}

			peerID, err := client.FetchPeerID(ctx)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPeerID, peerID)
			}
		})
	}
}

// TestConsensusClients_PeerIDs_HTTP tests the collection PeerIDs functionality with HTTP
func TestConsensusClients_PeerIDs_HTTP(t *testing.T) {
	// Create test servers for multiple clients
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := NodeIdentityResponse{
			Data: struct {
				PeerID             string   `json:"peer_id"`
				ENR                string   `json:"enr"`
				P2PAddresses       []string `json:"p2p_addresses"`
				DiscoveryAddresses []string `json:"discovery_addresses"`
				Metadata           struct {
					SeqNumber         string `json:"seq_number"`
					Attnets           string `json:"attnets"`
					SyncCommitteeNets string `json:"syncnets,omitempty"`
				} `json:"metadata"`
			}{
				PeerID: "16Uiu2HAkLighthouse1PeerIDExample",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := NodeIdentityResponse{
			Data: struct {
				PeerID             string   `json:"peer_id"`
				ENR                string   `json:"enr"`
				P2PAddresses       []string `json:"p2p_addresses"`
				DiscoveryAddresses []string `json:"discovery_addresses"`
				Metadata           struct {
					SeqNumber         string `json:"seq_number"`
					Attnets           string `json:"attnets"`
					SyncCommitteeNets string `json:"syncnets,omitempty"`
				} `json:"metadata"`
			}{
				PeerID: "16Uiu2HAkTeku1PeerIDExample",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server2.Close()

	// Create consensus clients collection
	clients := NewConsensusClients()

	client1 := NewConsensusClient(
		Lighthouse,
		"lighthouse-1",
		"v1.0.0",
		server1.URL,
		"http://localhost:8080",
		"enr:test1",
		"peer-id-1",
		"lighthouse-service-1",
		"container-1",
		9000,
	)

	client2 := NewConsensusClient(
		Teku,
		"teku-1",
		"v1.0.0",
		server2.URL,
		"http://localhost:8081",
		"enr:test2",
		"peer-id-2",
		"teku-service-1",
		"container-2",
		9001,
	)

	clients.Add(client1)
	clients.Add(client2)

	ctx := context.Background()

	// Test PeerIDs for all clients
	peerIDs, err := clients.PeerIDs(ctx)
	require.NoError(t, err)
	require.Len(t, peerIDs, 2)
	
	assert.Equal(t, "16Uiu2HAkLighthouse1PeerIDExample", peerIDs["lighthouse-1"])
	assert.Equal(t, "16Uiu2HAkTeku1PeerIDExample", peerIDs["teku-1"])

	// Test PeerIDsByType for Lighthouse clients
	lighthousePeerIDs, err := clients.PeerIDsByType(ctx, Lighthouse)
	require.NoError(t, err)
	require.Len(t, lighthousePeerIDs, 1)
	
	assert.Equal(t, "16Uiu2HAkLighthouse1PeerIDExample", lighthousePeerIDs["lighthouse-1"])

	// Test PeerIDsByType for Teku clients
	tekuPeerIDs, err := clients.PeerIDsByType(ctx, Teku)
	require.NoError(t, err)
	require.Len(t, tekuPeerIDs, 1)
	
	assert.Equal(t, "16Uiu2HAkTeku1PeerIDExample", tekuPeerIDs["teku-1"])
}

// TestConsensusClients_PeerIDs_Error tests error handling in PeerIDs
func TestConsensusClients_PeerIDs_Error(t *testing.T) {
	// Create a server that returns an error
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer errorServer.Close()

	clients := NewConsensusClients()
	client := NewConsensusClient(
		Lighthouse,
		"lighthouse-error",
		"v1.0.0",
		errorServer.URL,
		"http://localhost:8080",
		"enr:test",
		"peer-id",
		"lighthouse-service",
		"container",
		9000,
	)
	clients.Add(client)

	ctx := context.Background()

	// Test that error is properly propagated
	_, err := clients.PeerIDs(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch peer ID for client lighthouse-error")
}

// TestNodeIdentityResponse_JSON tests the JSON marshaling/unmarshaling
func TestNodeIdentityResponse_JSON(t *testing.T) {
	responseJSON := `{
		"data": {
			"peer_id": "16Uiu2HAkuVKJJuNnFVhfVjrw1nXJt6c2d1NcmAZqYLbA4Km7KLRZ",
			"enr": "enr:-MS4QBU9k_cMlyFm7Tlj4bpMRdiq6bOvl3KGfJUrm3JWy5hUr8l_M1S3TI2AXKQ5z_wQbr_jzb_LIGEr7vDRWKwv4_MBgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQKKFf5UZkP2-_CK9bM34hRzW5zR-ZPh1O6_bFf9NbCl1YN0Y3CCdl-DdWRwgnZf",
			"p2p_addresses": [
				"/ip4/127.0.0.1/tcp/9000",
				"/ip6/::1/tcp/9000"
			],
			"discovery_addresses": [
				"/ip4/127.0.0.1/udp/9000",
				"/ip6/::1/udp/9000"
			],
			"metadata": {
				"seq_number": "1",
				"attnets": "0xffffffffffffffff",
				"syncnets": "0xf"
			}
		}
	}`

	var response NodeIdentityResponse
	err := json.Unmarshal([]byte(responseJSON), &response)
	require.NoError(t, err)

	assert.Equal(t, "16Uiu2HAkuVKJJuNnFVhfVjrw1nXJt6c2d1NcmAZqYLbA4Km7KLRZ", response.Data.PeerID)
	assert.Contains(t, response.Data.ENR, "enr:-MS4QBU9k_")
	assert.Len(t, response.Data.P2PAddresses, 2)
	assert.Len(t, response.Data.DiscoveryAddresses, 2)
	assert.Equal(t, "1", response.Data.Metadata.SeqNumber)
	assert.Equal(t, "0xffffffffffffffff", response.Data.Metadata.Attnets)
}

// BenchmarkConsensusClient_FetchPeerID benchmarks the peer ID fetching
func BenchmarkConsensusClient_FetchPeerID(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := NodeIdentityResponse{
			Data: struct {
				PeerID             string   `json:"peer_id"`
				ENR                string   `json:"enr"`
				P2PAddresses       []string `json:"p2p_addresses"`
				DiscoveryAddresses []string `json:"discovery_addresses"`
				Metadata           struct {
					SeqNumber         string `json:"seq_number"`
					Attnets           string `json:"attnets"`
					SyncCommitteeNets string `json:"syncnets,omitempty"`
				} `json:"metadata"`
			}{
				PeerID: "16Uiu2HAkBenchmarkPeerID",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewConsensusClient(
		Lighthouse,
		"benchmark-client",
		"v1.0.0",
		server.URL,
		"http://localhost:8080",
		"enr:benchmark",
		"peer-benchmark",
		"service-benchmark",
		"container-benchmark",
		9000,
	)

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.FetchPeerID(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}