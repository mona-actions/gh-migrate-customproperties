package sync

import (
	"fmt"
	"log"
	"mona-actions/gh-migrate-customproperties/internal/api"
	"mona-actions/gh-migrate-customproperties/internal/file"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

// Package-level client
var ghAPI *api.GitHubAPI

func initializeAPI() {
	ghAPI = api.GetAPI()
}

// SyncStats tracks statistics about the sync operation
type SyncStats struct {
	FetchFailures    []string
	CreateFailures   []string
	TotalProcessed   int
	SuccessfulFetch  int
	SuccessfulCreate int
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

func SyncRepositoryProperties() {
	initializeAPI()

	spinner, _ := pterm.DefaultSpinner.Start("Syncing repository properties")
	spinner.UpdateText("Retrieving source custom properties from repositories")

	// Initialize sync stats
	stats := &SyncStats{}

	// Initialize and fetch properties
	repositories, err := file.ParseRepositoryFile(viper.GetString("REPOSITORY_LIST"))
	if err != nil {
		spinner.Fail(err.Error())
	}

	stats.TotalProcessed = len(repositories)
	repoProps := NewRepositoryProperties()

	if err := fetchProperties(repoProps, repositories, stats); err != nil {
		spinner.WarningPrinter.Println("Error during fetch phase... continuing")
	}

	spinner.UpdateText("Creating properties in target repositories")
	targetOwner := viper.GetString("TARGET_ORGANIZATION")

	// Create properties in target
	if err := createProperties(repoProps, targetOwner, stats); err != nil {
		log.Printf("Error during create phase: %v", err)
	}

	if len(stats.CreateFailures) > 0 && stats.SuccessfulCreate > 0 {
		spinner.Warning("Some repository properties failed to sync")
	} else if len(stats.CreateFailures) > 0 {
		spinner.Fail("All repositories failed to sync properties")
	} else {
		spinner.Success("All repository properties synced successfully")
	}
	printSyncSummary(stats)
}

// fetchProperties fetches properties for all repositories and tracks stats
func fetchProperties(rp *RepositoryProperties, repositories []string, stats *SyncStats) error {
	for _, fullRepo := range repositories {
		parts := strings.Split(fullRepo, "/")
		owner, repoName := parts[0], parts[1]

		props, err := ghAPI.GetRepositoryProperties(owner, repoName)
		if err != nil {
			log.Printf("Error fetching repository properties for %s: %v", fullRepo, err)
			stats.FetchFailures = append(stats.FetchFailures, fullRepo)
			continue
		}
		if props == nil {
			log.Printf("No repository properties found for %s", fullRepo)
			continue
		}

		rp.Repositories[repoName] = props
		stats.SuccessfulFetch++
	}

	return nil
}

// createProperties creates all stored properties in target repositories and tracks stats
func createProperties(rp *RepositoryProperties, targetOwner string, stats *SyncStats) error {
	for repoName, props := range rp.Repositories {
		err := ghAPI.CreateRepositoryProperties(targetOwner, repoName, props)
		if err != nil {
			log.Printf("Failed to create properties for repo %s: %v", repoName, err)
			stats.CreateFailures = append(stats.CreateFailures, repoName)
			continue
		}
		stats.SuccessfulCreate++
	}
	return nil
}

func printSyncSummary(stats *SyncStats) {
	fmt.Printf("\n=== Sync Operation Summary ===\n")
	fmt.Printf("📊 Total repositories processed: %d\n", stats.TotalProcessed)
	fmt.Printf("✅ Successfully fetched: %d\n", stats.SuccessfulFetch)
	fmt.Printf("✅ Successfully created: %d\n", stats.SuccessfulCreate)

	if len(stats.FetchFailures) > 0 {
		fmt.Printf("\n❌ Repositories that failed during fetch (%d):\n", len(stats.FetchFailures))
		for _, repo := range stats.FetchFailures {
			fmt.Printf("  - %s\n", repo)
		}
	}

	if len(stats.CreateFailures) > 0 {
		fmt.Printf("\n❌ Repositories that failed during create (%d):\n", len(stats.CreateFailures))
		for _, repo := range stats.CreateFailures {
			fmt.Printf("  - %s\n", repo)
		}
	}
}
