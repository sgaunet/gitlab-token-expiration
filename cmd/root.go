package cmd

import (
	"fmt"
	"os"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/spf13/cobra"
)

// DefaultNbDaysBeforeExp is the default number of days before expiration to display in yellow.
const DefaultNbDaysBeforeExp = 60

var gitlabID int64       // Gitlab project or group ID
var nbDaysBeforeExp uint // Number of days before expiration date to display it in yellow
var printRevoked bool
var printNoHeader bool
var printNoColor bool
var parallel int // Maximum number of concurrent GitLab API calls

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gitlab-expiration-token",
	Short: "Tool to list tokens with expiration date to check if they are expired or not.",
	Long: `Tool to list tokens with expiration date to check if they are expired or not.

Environment Variables:
  GITLAB_TOKEN  (required) GitLab personal access token with API access
  GITLAB_URI    (optional) GitLab instance URL (defaults to https://gitlab.com)

Examples:
  export GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx
  export GITLAB_URI=https://gitlab.example.com
  gitlab-token-expiration group -i 12345   # tokens of group 12345, its subgroups and projects
  gitlab-token-expiration group            # tokens of all groups where you are Owner
  gitlab-token-expiration project          # tokens of all projects where you are at least Maintainer`,
}

// printSkipped warns on stderr about the groups and projects skipped because
// GitLab denied access to their tokens.
func printSkipped(a *app.App) {
	skipped := a.Skipped()
	for _, s := range skipped {
		fmt.Fprintf(os.Stderr, "Warning: %s %s skipped: access denied (%v)\n", s.Kind, s.Path, s.Err)
	}
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d resource(s) skipped because access was denied\n", len(skipped))
	}
}

// newTableOutput returns the table renderer configured from the command-line flags.
func newTableOutput() views.TableOutput {
	return views.NewTableOutput(views.WithColorOption(!printNoColor),
		views.WithHeaderOption(!printNoHeader),
		views.WithPrintRevokedOption(printRevoked),
		views.WithNbDaysBeforeExp(nbDaysBeforeExp),
	)
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

	groupCmd.Flags().Int64VarP(&gitlabID, "id", "i", 0,
		"Gitlab Group ID (if omitted, all groups where you are Owner)")
	groupCmd.Flags().BoolVarP(&noRecursiveOption, "no-recursive", "n", false,
		"Do not list tokens of subgroups and projects")
	groupCmd.Flags().BoolVarP(&printRevoked, "revoked", "r", false, "Print revoked tokens")
	groupCmd.Flags().BoolVarP(&printNoHeader, "no-header", "H", false, "Do not print header")
	groupCmd.Flags().BoolVarP(&printNoColor, "no-color", "C", false, "Do not print color")
	groupCmd.Flags().UintVarP(&nbDaysBeforeExp, "days-before-expiration", "d", DefaultNbDaysBeforeExp,
		"Number of days before expiration date to display it in yellow")
	groupCmd.Flags().IntVarP(&parallel, "parallel", "p", app.DefaultParallel,
		"Maximum number of concurrent GitLab API calls")
	rootCmd.AddCommand(groupCmd)

	projectCmd.Flags().Int64VarP(&gitlabID, "id", "i", 0,
		"Gitlab Project ID (if omitted, all projects where you are at least Maintainer)")
	projectCmd.Flags().BoolVarP(&printRevoked, "revoked", "r", false, "Print revoked tokens")
	projectCmd.Flags().BoolVarP(&printNoHeader, "no-header", "H", false, "Do not print header")
	projectCmd.Flags().BoolVarP(&printNoColor, "no-color", "C", false, "Do not print color")
	projectCmd.Flags().UintVarP(&nbDaysBeforeExp, "days-before-expiration", "d", DefaultNbDaysBeforeExp,
		"Number of days before expiration date to display it in yellow")
	projectCmd.Flags().IntVarP(&parallel, "parallel", "p", app.DefaultParallel,
		"Maximum number of concurrent GitLab API calls")
	rootCmd.AddCommand(projectCmd)

	patCmd.Flags().BoolVarP(&printRevoked, "revoked", "r", false, "Print revoked tokens")
	patCmd.Flags().BoolVarP(&printNoHeader, "no-header", "H", false, "Do not print header")
	patCmd.Flags().BoolVarP(&printNoColor, "no-color", "C", false, "Do not print color")
	patCmd.Flags().UintVarP(&nbDaysBeforeExp, "days-before-expiration", "d", DefaultNbDaysBeforeExp,
		"Number of days before expiration date to display it in yellow")
	rootCmd.AddCommand(patCmd)
}
