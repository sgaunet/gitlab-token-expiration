package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
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
	Use:   "gitlab-token-expiration",
	Short: "Tool to list tokens with expiration date to check if they are expired or not.",
	Long: `Tool to list tokens with expiration date to check if they are expired or not.

Environment Variables:
  GITLAB_TOKEN  (required) GitLab personal access token with API access
  GITLAB_URI    (optional) GitLab instance URL (defaults to https://gitlab.com)

Token lists that GitLab does not let you read (401/403), such as the access
tokens of a group where you are only Maintainer, are skipped with a warning on
stderr. The command fails when no token list at all could be read.

Examples:
  export GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx
  export GITLAB_URI=https://gitlab.example.com
  gitlab-token-expiration group -i 12345   # tokens of group 12345, its subgroups and projects
  gitlab-token-expiration group            # tokens of all groups where you are Owner
  gitlab-token-expiration project          # tokens of all projects where you are at least Maintainer`,
	// Execute prints the errors, see below.
	SilenceErrors: true,
}

var (
	// errIgnoreCaseWithoutPattern is returned when --ignore-case is set without any filter pattern.
	errIgnoreCaseWithoutPattern = errors.New("--ignore-case requires a --filter-* or --exclude-* pattern")
	errInvalidParallel          = errors.New("--parallel must be at least 1")
	errNoRecursiveWithoutID     = errors.New("--no-recursive requires --id")
	errNoGitlabToken            = errors.New(
		"GITLAB_TOKEN is not set: export a GitLab token with at least the read_api scope")
	errNothingRead = errors.New("no token could be read: GitLab denied access to every token list " +
		"(check GITLAB_TOKEN and your role)")
)

// output is how a command shows the tokens it retrieved.
type output struct {
	filter   filter.TokenFilter
	renderer views.Renderer
}

// prepare checks the command-line flags and the environment, and builds the
// output. It runs before any GitLab API call, so that a mistake is reported
// right away.
func prepare() (output, error) {
	if parallel < 1 {
		return output{}, fmt.Errorf("%w, got %d", errInvalidParallel, parallel)
	}
	tokenFilter, err := newTokenFilter()
	if err != nil {
		return output{}, err
	}
	renderer, err := newRenderer()
	if err != nil {
		return output{}, err
	}
	if os.Getenv("GITLAB_TOKEN") == "" {
		return output{}, errNoGitlabToken
	}
	return output{filter: tokenFilter, renderer: renderer}, nil
}

// newApp returns the App used by the commands, configured from the flags.
func newApp() (*app.App, error) {
	a, err := app.NewApp(app.WithParallel(parallel))
	if err != nil {
		return nil, fmt.Errorf("creating the GitLab client: %w", err)
	}
	return a, nil
}

// render warns on stderr about the token lists skipped because access was
// denied, then renders the tokens kept by the filter. When GitLab denied
// access to every token list, nothing is rendered and errNothingRead is
// returned, so that an empty result is not mistaken for "no token".
func (o output) render(a *app.App, tokens []dto.Token) error {
	printSkipped(a)
	if len(a.Skipped()) > 0 && a.Listed() == 0 {
		return errNothingRead
	}
	if err := o.renderer.Render(o.filter.Apply(tokens)); err != nil {
		return fmt.Errorf("rendering tokens: %w", err)
	}
	return nil
}

// printSkipped warns on stderr about the token lists skipped because GitLab
// denied access to them.
func printSkipped(a *app.App) {
	skipped := a.Skipped()
	for _, s := range skipped {
		fmt.Fprintf(os.Stderr, "Warning: %s of %s skipped: access denied (%v)\n", s.Kind, s.Path, s.Err)
	}
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: %d token list(s) skipped because access was denied\n", len(skipped))
	}
}

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

// newTokenFilter builds the token filter from the command-line flags:
// the --filter-*/--exclude-* patterns, and revoked tokens unless -r is set.
func newTokenFilter() (filter.TokenFilter, error) {
	f := filter.TokenFilter{HideRevoked: !printRevoked}
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
		"Output format: "+strings.Join(views.Formats(), ", "))
}

// newRenderer returns the renderer selected by --output (case-insensitive),
// configured from the command-line flags.
func newRenderer() (views.Renderer, error) { //nolint:ireturn // picks the renderer from --output
	format := strings.ToLower(outputFormat)
	switch format {
	case views.FormatTable:
		// fatih/color sets NoColor when stdout is not a terminal or NO_COLOR is set.
		return views.NewTableOutput(views.WithColorOption(!printNoColor && !color.NoColor),
			views.WithHeaderOption(!printNoHeader),
			views.WithNbDaysBeforeExp(nbDaysBeforeExp),
		), nil
	default:
		r, err := views.NewStructuredOutput(format, os.Stdout)
		if err != nil {
			return nil, fmt.Errorf("--output: %w (expected one of %s)", err, strings.Join(views.Formats(), ", "))
		}
		return r, nil
	}
}

// exitInterrupted is the exit status after Ctrl-C (128 + SIGINT), as in shells.
const exitInterrupted = 130

// Execute runs the command given on the command line. Ctrl-C (or SIGTERM)
// cancels the GitLab API calls in flight; a second Ctrl-C kills the process.
// On error, it prints the error on stderr and exits with status 1.
func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		stop() // restore the default behaviour for a second signal
	}()
	err := rootCmd.ExecuteContext(ctx)
	if err == nil {
		return
	}
	if ctx.Err() != nil && errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, "Interrupted")
		os.Exit(exitInterrupted)
	}
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}

func init() {
	// -C also disables the colors of the progress spinners on stderr, which
	// run before the table is rendered.
	rootCmd.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		if printNoColor {
			pterm.DisableColor()
		}
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	groupCmd.Flags().Int64VarP(&gitlabID, "id", "i", 0,
		"Gitlab Group ID (if omitted, all groups where you are Owner)")
	groupCmd.Flags().BoolVarP(&noRecursiveOption, "no-recursive", "n", false,
		"Do not list tokens of subgroups and projects (requires --id)")
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
