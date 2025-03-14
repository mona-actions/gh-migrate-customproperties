package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"github.com/google/go-github/v66/github"
	"github.com/jferrl/go-githubauth"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// ClientConfig holds all possible configuration options for creating a GitHub client
type ClientConfig struct {
	Token          string
	Hostname       string
	AppID          string
	PrivateKey     []byte
	InstallationID int64
}

// GitHubAPI holds the clients for interacting with GitHub
type GitHubAPI struct {
	sourceClient      *github.Client
	targetClient      *github.Client
	sourceGraphClient *RateLimitAwareGraphQLClient
	targetGraphClient *RateLimitAwareGraphQLClient
}

// Package-level instance of GitHubAPI
var defaultAPI *GitHubAPI

// GetAPI returns the default GitHubAPI instance, initializing it if necessary
func GetAPI() *GitHubAPI {
	if defaultAPI == nil {
		defaultAPI = NewGitHubAPI()
	}
	return defaultAPI
}

// For testing purposes - allows resetting the default API
func resetAPI() {
	defaultAPI = nil
}

// newGitHubAPI creates a new GitHubAPI instance with configured clients
// Now private since we want to control initialization through GetAPI()
func NewGitHubAPI() *GitHubAPI {
	sourceConfig := ClientConfig{
		Token:          viper.GetString("SOURCE_TOKEN"),
		Hostname:       viper.GetString("SOURCE_HOSTNAME"),
		AppID:          viper.GetString("SOURCE_APP_ID"),
		PrivateKey:     []byte(viper.GetString("SOURCE_PRIVATE_KEY")),
		InstallationID: viper.GetInt64("SOURCE_INSTALLATION_ID"),
	}

	targetConfig := ClientConfig{
		Token:          viper.GetString("TARGET_TOKEN"),
		Hostname:       viper.GetString("TARGET_HOSTNAME"),
		AppID:          viper.GetString("TARGET_APP_ID"),
		PrivateKey:     []byte(viper.GetString("TARGET_PRIVATE_KEY")),
		InstallationID: viper.GetInt64("TARGET_INSTALLATION_ID"),
	}

	sourceClient := newGitHubClient(sourceConfig)
	targetClient := newGitHubClient(targetConfig)
	sourceGraphClient := newGitHubGraphQLClient(sourceConfig)
	targetGraphClient := newGitHubGraphQLClient(targetConfig)

	return &GitHubAPI{
		sourceClient:      sourceClient,
		targetClient:      targetClient,
		sourceGraphClient: sourceGraphClient,
		targetGraphClient: targetGraphClient,
	}
}

// createAuthenticatedClient creates an HTTP client with proper authentication and rate limiting
func createAuthenticatedClient(config ClientConfig) (*http.Client, error) {
	var httpClient *http.Client

	if config.AppID != "" && len(config.PrivateKey) != 0 && config.InstallationID != 0 {
		// GitHub App authentication
		appIDInt, err := strconv.ParseInt(config.AppID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error converting app ID to int64: %v", err)
		}

		appToken, err := githubauth.NewApplicationTokenSource(appIDInt, config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("error creating app token: %v", err)
		}

		installationToken := githubauth.NewInstallationTokenSource(config.InstallationID, appToken)
		httpClient = oauth2.NewClient(context.Background(), installationToken)
	} else if config.Token != "" {
		// Personal access token authentication
		src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
		httpClient = oauth2.NewClient(context.Background(), src)
	} else {
		return nil, fmt.Errorf("please provide either a token or GitHub App credentials")
	}

	rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(httpClient.Transport)
	if err != nil {
		return nil, err
	}

	return rateLimiter, nil
}

// newGitHubClient creates a new GitHub REST client based on the provided configuration
func newGitHubClient(config ClientConfig) *github.Client {
	httpClient, err := createAuthenticatedClient(config)
	if err != nil {
		log.Fatalf("Failed to create authenticated client: %v", err)
	}

	client := github.NewClient(httpClient)

	// Configure enterprise URL if hostname is provided
	if config.Hostname != "" {
		hostname := strings.TrimSuffix(config.Hostname, "/")
		if !strings.HasPrefix(hostname, "https://") {
			hostname = "https://" + hostname
		}
		baseURL := fmt.Sprintf("%s/api/v3/", hostname)
		client, err = client.WithEnterpriseURLs(baseURL, baseURL)
		if err != nil {
			log.Fatalf("Failed to configure enterprise URLs: %v", err)
		}
	}

	return client
}

type RateLimitAwareGraphQLClient struct {
	client *githubv4.Client
}

// newGitHubGraphQLClient creates a new GitHub GraphQL client based on the provided configuration
func newGitHubGraphQLClient(config ClientConfig) *RateLimitAwareGraphQLClient {
	httpClient, err := createAuthenticatedClient(config)
	if err != nil {
		log.Fatalf("Failed to create authenticated client: %v", err)
	}

	var baseClient *githubv4.Client

	// If hostname is provided, create enterprise client
	if config.Hostname != "" {
		hostname := strings.TrimSuffix(config.Hostname, "/")
		if !strings.HasPrefix(hostname, "https://") {
			hostname = "https://" + hostname
		}
		baseClient = githubv4.NewEnterpriseClient(hostname+"/api/graphql", httpClient)
	} else {
		baseClient = githubv4.NewClient(httpClient)
	}

	return &RateLimitAwareGraphQLClient{
		client: baseClient,
	}
}

func (c *RateLimitAwareGraphQLClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	var rateLimitQuery struct {
		RateLimit struct {
			Remaining int
			ResetAt   githubv4.DateTime
		}
	}

	for {
		// Check the current rate limit
		if err := c.client.Query(ctx, &rateLimitQuery, nil); err != nil {
			return err
		}

		//log.Println("Rate limit remaining:", rateLimitQuery.RateLimit.Remaining)

		if rateLimitQuery.RateLimit.Remaining > 0 {
			// Proceed with the actual query
			err := c.client.Query(ctx, q, variables)
			if err != nil {
				return err
			}
			return nil
		} else {
			// Sleep until rate limit resets
			log.Println("Rate limit exceeded, sleeping until reset at:", rateLimitQuery.RateLimit.ResetAt.Time)
			time.Sleep(time.Until(rateLimitQuery.RateLimit.ResetAt.Time))
		}
	}
}

func (api *GitHubAPI) GetRepositoryProperties(owner, repo string) ([]*github.CustomPropertyValue, error) {
	ctx := context.Background()

	repoInfo, _, err := api.sourceClient.Repositories.GetAllCustomPropertyValues(ctx, owner, repo)
	if err != nil {
		if strings.Contains(err.Error(), "403 Resource not accessible by integration") {
			return nil, err
		}
		return nil, err
	}

	return repoInfo, nil
}

func (api *GitHubAPI) CreateRepositoryProperties(owner, repo string, properties []*github.CustomPropertyValue) error {
	ctx := context.Background()

	_, err := api.targetClient.Repositories.CreateOrUpdateCustomProperties(ctx, owner, repo, properties)
	if err != nil {
		if strings.Contains(err.Error(), "403 Resource not accessible by integration") {
			return err
		}
		return err
	}

	return nil
}

func (api *GitHubAPI) GetSourceAuthenticatedUser() (*github.User, error) {
	ctx := context.Background()

	user, _, err := api.sourceClient.Users.Get(ctx, "")
	if err != nil {
		if strings.Contains(err.Error(), "403 Resource not accessible by integration") {
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

func (api *GitHubAPI) GetTargetAuthenticatedUser() (*github.User, error) {
	ctx := context.Background()

	user, _, err := api.targetClient.Users.Get(ctx, "")
	if err != nil {
		if strings.Contains(err.Error(), "403 Resource not accessible by integration") {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (api *GitHubAPI) GetSourceGraphQLAuthenticatedUser() (*github.User, error) {
	ctx := context.Background()

	var query struct {
		Viewer struct {
			Login string
			Email string
		}
	}

	err := api.sourceGraphClient.Query(ctx, &query, nil)
	if err != nil {
		return nil, err
	}

	user := &github.User{
		Login: github.String(query.Viewer.Login),
		Email: github.String(query.Viewer.Email),
	}

	return user, nil
}

func (api *GitHubAPI) GetTargetGraphQLAuthenticatedUser() (*github.User, error) {
	ctx := context.Background()

	var query struct {
		Viewer struct {
			Login string
			Email string
		}
	}

	err := api.targetGraphClient.Query(ctx, &query, nil)
	if err != nil {
		return nil, err
	}

	user := &github.User{
		Login: github.String(query.Viewer.Login),
		Email: github.String(query.Viewer.Email),
	}

	return user, nil
}
