package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const DefaultNbDaysBeforeExp = 60

var gitlabID int         // Gitlab project or group ID
var nbDaysBeforeExp uint // Number of days before expiration date to display it in yellow
var printRevoked bool
var printNoHeader bool
var printNoColor bool

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

	groupCmd.Flags().IntVarP(&gitlabID, "id", "i", 0, "Gitlab Group ID")
	groupCmd.Flags().BoolVarP(&noRecursiveOption, "no-recursive", "n", false, "Do not list tokens of subgroups and projects")
	groupCmd.Flags().BoolVarP(&printRevoked, "revoked", "r", false, "Print revoked tokens")
	groupCmd.Flags().BoolVarP(&printNoHeader, "no-header", "H", false, "Do not print header")
	groupCmd.Flags().BoolVarP(&printNoColor, "no-color", "C", false, "Do not print color")
	groupCmd.Flags().UintVarP(&nbDaysBeforeExp, "days-before-expiration", "d", DefaultNbDaysBeforeExp, "Number of days before expiration date to display it in yellow")
	rootCmd.AddCommand(groupCmd)

	projectCmd.Flags().IntVarP(&gitlabID, "id", "i", 0, "Gitlab Project ID")
	projectCmd.Flags().BoolVarP(&printRevoked, "revoked", "r", false, "Print revoked tokens")
	projectCmd.Flags().BoolVarP(&printNoHeader, "no-header", "H", false, "Do not print header")
	projectCmd.Flags().BoolVarP(&printNoColor, "no-color", "C", false, "Do not print color")
	projectCmd.Flags().UintVarP(&nbDaysBeforeExp, "days-before-expiration", "d", DefaultNbDaysBeforeExp, "Number of days before expiration date to display it in yellow")
	rootCmd.AddCommand(projectCmd)

	patCmd.Flags().BoolVarP(&printRevoked, "revoked", "r", false, "Print revoked tokens")
	patCmd.Flags().BoolVarP(&printNoHeader, "no-header", "H", false, "Do not print header")
	patCmd.Flags().BoolVarP(&printNoColor, "no-color", "C", false, "Do not print color")
	patCmd.Flags().UintVarP(&nbDaysBeforeExp, "days-before-expiration", "d", DefaultNbDaysBeforeExp, "Number of days before expiration date to display it in yellow")
	rootCmd.AddCommand(patCmd)
}
