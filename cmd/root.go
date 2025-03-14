/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"mona-actions/gh-migrate-customproperties/pkg/sync"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-migrate-customproperties",
	Short: "help migrate repo custom properties",
	Long: `This is a migration CLI extension that provides additional capabilities to migrate
	repositories with custom properties from one organization to another.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get parameters
		targetOrganization := cmd.Flag("target-organization").Value.String()
		sourceToken := cmd.Flag("source-token").Value.String()
		targetToken := cmd.Flag("target-token").Value.String()
		ghHostname := cmd.Flag("source-hostname").Value.String()
		repositoryList := cmd.Flag("repository-list").Value.String()
		convertProps, _ := cmd.Flags().GetBool("convert-props")

		// Set ENV variables
		os.Setenv("GHMC_TARGET_ORGANIZATION", targetOrganization)
		os.Setenv("GHMC_SOURCE_TOKEN", sourceToken)
		os.Setenv("GHMC_TARGET_TOKEN", targetToken)
		os.Setenv("GHMC_SOURCE_HOSTNAME", ghHostname)
		os.Setenv("GHMC_REPOSITORY_LIST", repositoryList)
		os.Setenv("GHMC_CONVERT_PROPS", strconv.FormatBool(convertProps))

		// Bind ENV variables in Viper
		viper.BindEnv("TARGET_ORGANIZATION")
		viper.BindEnv("SOURCE_TOKEN")
		viper.BindEnv("TARGET_TOKEN")
		viper.BindEnv("SOURCE_HOSTNAME")
		viper.BindEnv("REPOSITORY_LIST")
		viper.BindEnv("CONVERT_PROPS")

		viper.BindEnv("SOURCE_PRIVATE_KEY")
		viper.BindEnv("SOURCE_APP_ID")
		viper.BindEnv("SOURCE_INSTALLATION_ID")
		viper.BindEnv("TARGET_PRIVATE_KEY")
		viper.BindEnv("TARGET_APP_ID")
		viper.BindEnv("TARGET_INSTALLATION_ID")

		sync.SyncRepositoryProperties()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("target-organization", "t", "", "Target Organization to sync properties to")
	rootCmd.MarkFlagRequired("target-organization")

	rootCmd.Flags().StringP("source-token", "a", "", "Source Organization GitHub token. Scopes: read:org, read:user, user:email")
	rootCmd.MarkFlagRequired("source-token")

	rootCmd.Flags().StringP("target-token", "b", "", "Target Organization GitHub token. Scopes: admin:org")
	rootCmd.MarkFlagRequired("target-token")

	rootCmd.Flags().StringP("source-hostname", "u", "", "GitHub Enterprise source hostname url (optional) Ex. https://github.example.com")

	rootCmd.Flags().StringP("repository-list", "r", "", "File containing list of repositories to sync properties from. One repository per line. Must be in owner/repo format.")
	rootCmd.MarkFlagRequired("repository-list")

	rootCmd.Flags().BoolP("convert-props", "c", false, "Convert custom properties to target format. Default: false; Currently only supports single-select to multi-select conversion")

	viper.SetEnvPrefix("GHMC") // GHMigrateCustomProperties

	// Read in environment variables that match
	viper.AutomaticEnv()
}
