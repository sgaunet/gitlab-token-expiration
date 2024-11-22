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
	// encCmd.Flags().StringVar(&inputFile, "i", "", "file to encrypt")
	// encCmd.Flags().StringVar(&outputFile, "o", "", "output file")
	// encCmd.Flags().StringVar(&keyFile, "k", "", "file containing the key to encrypt (or set gitlab-expiration-token_KEY env variable)")
	// encCmd.Flags().BoolVar(&rmOption, "del", false, "delete source file after encryption")
	// rootCmd.AddCommand(encCmd)

	// decCmd.Flags().StringVar(&inputFile, "i", "", "file to decrypt")
	// decCmd.Flags().StringVar(&outputFile, "o", "", "output file")
	// decCmd.Flags().StringVar(&keyFile, "k", "", "file containing the key to decrypt (or set gitlab-expiration-token_KEY env variable)")
	groupCmd.Flags().IntVar(&gitlabID, "id", 0, "Gitlab Group ID")
	rootCmd.AddCommand(groupCmd)

	projectCmd.Flags().IntVar(&gitlabID, "id", 0, "Gitlab Project ID")
	rootCmd.AddCommand(projectCmd)

	rootCmd.AddCommand(patCmd)
}
