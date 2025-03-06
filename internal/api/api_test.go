package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v62/github"
)

// mockHTTPClient implements a mock http.RoundTripper for testing
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestCreateAuthenticatedClient(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
	}{
		{
			name: "valid token config",
			config: ClientConfig{
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "valid GitHub App config",
			config: ClientConfig{
				AppID:          "123",
				PrivateKey:     []byte("test-key"),
				InstallationID: 456,
			},
			wantErr: true,
		},
		{
			name:    "missing credentials",
			config:  ClientConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := createAuthenticatedClient(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if client == nil {
					t.Error("expected client, got nil")
				}
			}
		})
	}
}

func TestNewGitHubClient(t *testing.T) {
	tests := []struct {
		name     string
		config   ClientConfig
		wantHost string
	}{
		{
			name: "github.com client",
			config: ClientConfig{
				Token: "test-token",
			},
			wantHost: "https://api.github.com/",
		},
		{
			name: "enterprise client",
			config: ClientConfig{
				Token:    "test-token",
				Hostname: "github.example.com",
			},
			wantHost: "https://github.example.com/api/v3/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newGitHubClient(tt.config)
			if client == nil {
				t.Error("expected client, got nil")
			}
		})
	}
}

func TestRateLimitAwareGraphQLClient_Query(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"rateLimit":{"remaining":100,"resetAt":"2024-12-31T23:59:59Z"}}}`))
	}))
	defer server.Close()

	config := ClientConfig{
		Token: "test-token",
	}

	client := newGitHubGraphQLClient(config)
	if client == nil {
		t.Fatal("expected client, got nil")
	}

	ctx := context.Background()
	var query struct {
		Viewer struct {
			Login string
		}
	}

	err := client.Query(ctx, &query, nil)
	if err == nil {
		t.Error("expected error due to test environment, got nil")
	}
}

func TestGitHubAPI_GetRepositoryProperties(t *testing.T) {
	mockClient := github.NewClient(nil)
	api := &GitHubAPI{
		sourceClient: mockClient,
	}

	properties, err := api.GetRepositoryProperties("testowner", "testrepo")
	if err == nil {
		t.Error("expected error due to no actual GitHub connection, got nil")
	}
	if properties != nil {
		t.Error("expected nil properties, got properties object")
	}
}

func TestGitHubAPI_CreateRepositoryProperties(t *testing.T) {
	mockClient := github.NewClient(nil)
	api := &GitHubAPI{
		targetClient: mockClient,
	}

	testValue := "test-value"
	testProperties := []*github.CustomPropertyValue{
		{
			PropertyName: "test-property",
			Value:        &testValue,
		},
	}

	err := api.CreateRepositoryProperties("testowner", "testrepo", testProperties)
	if err == nil {
		t.Error("expected error due to no actual GitHub connection, got nil")
	}
}
