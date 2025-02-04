package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var gitlabID int // Gitlab project or group ID

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitlab-expiration-token",
	Short: "Tool to list tokens with expiration date to check if they are expired or not.",
	Long:  `Tool to list tokens with expiration date to check if they are expired or not.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	groupCmd.Flags().IntVar(&gitlabID, "id", 0, "Gitlab Group ID")
	groupCmd.Flags().BoolVar(&noRecursiveOption, "no-recursive", false, "Do not list tokens of subgroups and projects")
	rootCmd.AddCommand(groupCmd)

	projectCmd.Flags().IntVar(&gitlabID, "id", 0, "Gitlab Project ID")
	rootCmd.AddCommand(projectCmd)

	rootCmd.AddCommand(patCmd)
}
