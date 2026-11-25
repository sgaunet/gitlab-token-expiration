package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/filter"
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
var parallel int          // Maximum number of concurrent GitLab API calls
var filterName string     // Keep the tokens whose name matches this regular expression
var excludeName string    // Hide the tokens whose name matches this regular expression
var filterSource string   // Keep the tokens whose source matches this regular expression
var excludeSource string  // Hide the tokens whose source matches this regular expression
var filterIgnoreCase bool // Case-insensitive match for every filter pattern
var outputFormat string   // table, json or yaml

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

// errIgnoreCaseWithoutPattern is returned when --ignore-case is set without any filter pattern.
var errIgnoreCaseWithoutPattern = errors.New("--ignore-case requires a --filter-* or --exclude-* pattern")

// addFilterFlags adds the token filter flags to cmd. The source flags are only
// added when withSource is set, since personal access tokens have no source.
func addFilterFlags(cmd *cobra.Command, withSource bool) {
	cmd.Flags().StringVar(&filterName, "filter-name", "",
		"Only show tokens whose name matches this regular expression")
	cmd.Flags().StringVar(&excludeName, "exclude-name", "",
		"Hide tokens whose name matches this regular expression")
	if withSource {
		cmd.Flags().StringVar(&filterSource, "filter-source", "",
			"Only show tokens whose source (group or project path) matches this regular expression")
		cmd.Flags().StringVar(&excludeSource, "exclude-source", "",
			"Hide tokens whose source (group or project path) matches this regular expression")
	}
	cmd.Flags().BoolVar(&filterIgnoreCase, "ignore-case", false,
		"Case-insensitive match for every --filter-* and --exclude-* pattern")
}

// newTokenFilter builds the token filter from the command-line flags.
func newTokenFilter() (filter.TokenFilter, error) {
	var f filter.TokenFilter
	patterns := []struct {
		flag    string
		pattern string
		dst     **regexp.Regexp
	}{
		{"--filter-name", filterName, &f.IncludeName},
		{"--exclude-name", excludeName, &f.ExcludeName},
		{"--filter-source", filterSource, &f.IncludeSource},
		{"--exclude-source", excludeSource, &f.ExcludeSource},
	}
	hasPattern := false
	for _, p := range patterns {
		if p.pattern == "" {
			continue
		}
		re, err := filter.CompilePattern(p.pattern, filterIgnoreCase)
		if err != nil {
			return filter.TokenFilter{}, fmt.Errorf("%s: %w", p.flag, err)
		}
		*p.dst = re
		hasPattern = true
	}
	if filterIgnoreCase && !hasPattern {
		return filter.TokenFilter{}, errIgnoreCaseWithoutPattern
	}
	return f, nil
}

// addOutputFlag adds the output format flag to cmd.
func addOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&outputFormat, "output", "o", views.FormatTable,
		"Output format: "+strings.Join(views.Formats, ", "))
}

// newRenderer returns the renderer selected by --output, configured from the
// command-line flags.
func newRenderer() (views.Renderer, error) { //nolint:ireturn // picks the renderer from --output
	switch outputFormat {
	case views.FormatTable:
		return views.NewTableOutput(views.WithColorOption(!printNoColor),
			views.WithHeaderOption(!printNoHeader),
			views.WithPrintRevokedOption(printRevoked),
			views.WithNbDaysBeforeExp(nbDaysBeforeExp),
		), nil
	default:
		r, err := views.NewStructuredOutput(outputFormat, os.Stdout, printRevoked)
		if err != nil {
			return nil, fmt.Errorf("--output: %w (expected one of %s)", err, strings.Join(views.Formats, ", "))
		}
		return r, nil
	}
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
	addFilterFlags(groupCmd, true)
	addOutputFlag(groupCmd)
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
	addFilterFlags(projectCmd, true)
	addOutputFlag(projectCmd)
	rootCmd.AddCommand(projectCmd)

	patCmd.Flags().BoolVarP(&printRevoked, "revoked", "r", false, "Print revoked tokens")
	patCmd.Flags().BoolVarP(&printNoHeader, "no-header", "H", false, "Do not print header")
	patCmd.Flags().BoolVarP(&printNoColor, "no-color", "C", false, "Do not print color")
	patCmd.Flags().UintVarP(&nbDaysBeforeExp, "days-before-expiration", "d", DefaultNbDaysBeforeExp,
		"Number of days before expiration date to display it in yellow")
	addFilterFlags(patCmd, false)
	addOutputFlag(patCmd)
	rootCmd.AddCommand(patCmd)
}
