package sync

import (
	"fmt"
	"log"
	"mona-actions/gh-migrate-customproperties/internal/api"
	"mona-actions/gh-migrate-customproperties/internal/file"

	"github.com/google/go-github/v62/github"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

// Package-level client
var ghAPI *api.GitHubAPI

func initializeAPI() {
	ghAPI = api.GetAPI()
}

// RepositoryProperties stores custom properties for all repositories
type RepositoryProperties struct {
	Repositories map[string][]*github.CustomPropertyValue
}

// NewRepositoryProperties initializes a new RepositoryProperties instance
func NewRepositoryProperties() *RepositoryProperties {
	return &RepositoryProperties{
		Repositories: make(map[string][]*github.CustomPropertyValue),
	}
}

func SyncRepositoryProperties() error {

	initializeAPI()

	spinner, _ := pterm.DefaultSpinner.Start("Syncing repository properties")
	spinner.UpdateText("Retrieving source custom properties from repositories")

	// Initialize and fetch properties
	repositories, err := file.ParseRepositoryFile(viper.GetString("REPOSITORY_LIST"))
	if err != nil {
		return err
	}

	repoProps := NewRepositoryProperties()
	if err := repoProps.FetchProperties(repositories); err != nil {
		return err
	}

	spinner.UpdateText("Creating properties in target repositories")
	targetOwner := viper.GetString("TARGET_ORGANIZATION")

	// Create properties in target
	if err := repoProps.CreateProperties(targetOwner); err != nil {
		return err
	}

	spinner.Success("Repository properties synced successfully")
	return nil
}

func (rp *RepositoryProperties) FetchProperties(repositories []string) error {
	owner := viper.GetString("SOURCE_ORGANIZATION")

	for _, repo := range repositories {
		props, err := ghAPI.GetRepositoryProperties(owner, repo)
		if err != nil {
			log.Printf("Error fetching repository properties for repo %s: %v", repo, err)
			continue
		}
		if props == nil {
			log.Printf("No repository properties found for repo %s", repo)
			continue
		}

		rp.Repositories[repo] = props
	}

	return nil
}

// CreateProperties creates all stored properties in target repositories
func (rp *RepositoryProperties) CreateProperties(targetOwner string) error {
	for repo, props := range rp.Repositories {
		err := ghAPI.CreateRepositoryProperties(targetOwner, repo, props)
		if err != nil {
			return fmt.Errorf("failed to create properties for repo %s: %v", repo, err)
		}
	}
	return nil
}
