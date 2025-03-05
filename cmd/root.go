/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gh-migrate-customproperties",
	Short: "help migrate repo custom properties",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

this CLI extension is used for....
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		// Get parameters
		sourceOrganization := cmd.Flag("source-organization").Value.String()
		targetOrganization := cmd.Flag("target-organization").Value.String()
		sourceToken := cmd.Flag("source-token").Value.String()
		targetToken := cmd.Flag("target-token").Value.String()
		ghHostname := cmd.Flag("source-hostname").Value.String()

		// Set ENV variables
		os.Setenv("GHET_SOURCE_ORGANIZATION", sourceOrganization)
		os.Setenv("GHET_TARGET_ORGANIZATION", targetOrganization)
		os.Setenv("GHET_SOURCE_TOKEN", sourceToken)
		os.Setenv("GHET_TARGET_TOKEN", targetToken)
		os.Setenv("GHET_SOURCE_HOSTNAME", ghHostname)

		// Bind ENV variables in Viper
		viper.BindEnv("SOURCE_ORGANIZATION")
		viper.BindEnv("TARGET_ORGANIZATION")
		viper.BindEnv("SOURCE_TOKEN")
		viper.BindEnv("TARGET_TOKEN")
		viper.BindEnv("SOURCE_HOSTNAME")
		viper.BindEnv("SOURCE_PRIVATE_KEY")
		viper.BindEnv("SOURCE_APP_ID")
		viper.BindEnv("SOURCE_INSTALLATION_ID")
		viper.BindEnv("TARGET_PRIVATE_KEY")
		viper.BindEnv("TARGET_APP_ID")
		viper.BindEnv("TARGET_INSTALLATION_ID")

		//sync.SyncProperties()
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

	rootCmd.Flags().StringP("source-organization", "s", "", "Source Organization to sync teams from")
	rootCmd.MarkFlagRequired("source-organization")

	rootCmd.Flags().StringP("target-organization", "t", "", "Target Organization to sync teams from")
	rootCmd.MarkFlagRequired("target-organization")

	rootCmd.Flags().StringP("source-token", "a", "", "Source Organization GitHub token. Scopes: read:org, read:user, user:email")
	rootCmd.MarkFlagRequired("source-token")

	rootCmd.Flags().StringP("target-token", "b", "", "Target Organization GitHub token. Scopes: admin:org")
	rootCmd.MarkFlagRequired("target-token")

	rootCmd.Flags().StringP("source-hostname", "u", "", "GitHub Enterprise source hostname url (optional) Ex. https://github.example.com")

	viper.SetEnvPrefix("GHMC") // GHMigrateCustomProperties

	// Read in environment variables that match
	viper.AutomaticEnv()
}
