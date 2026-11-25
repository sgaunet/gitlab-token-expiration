// Package cmd contains the command-line interface for the GitLab token expiration tool.
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/api/client-go"
)

var noRecursiveOption bool

// groupCmd represents the group command to list expirable tokens of a group.
var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "List expirable tokens of a group, or of all your groups",
	Long: `List expirable tokens (access tokens and deploy tokens) of a group.

With --id, list the tokens of that group and, unless --no-recursive is set,
of all its subgroups and projects.

Without --id, list the tokens of every group where you have the Owner role
(the role GitLab requires to list group tokens).`,
	Run: func(_ *cobra.Command, _ []string) {
		tokenFilter, err := newTokenFilter()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		v := newTableOutput()
		// When several groups are processed, skip those whose tokens cannot be read
		// instead of failing the whole run.
		multipleGroups := gitlabID == 0 || !noRecursiveOption
		a := app.NewApp(v,
			app.WithRevokedToken(printRevoked),
			app.WithParallel(parallel),
			app.WithSkipDenied(multipleGroups),
		)
		ctx := context.Background()

		var tokens []dto.Token
		switch {
		case gitlabID == 0:
			tokens, err = tokensOfAllGroups(ctx, a)
		case noRecursiveOption:
			tokens, err = tokensOfGroup(ctx, a)
		default:
			tokens, err = tokensOfGroupRecursive(ctx, a)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		printSkipped(a)
		tokens = tokenFilter.Apply(tokens)
		if err := v.Render(tokens); err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering tokens: %v\n", err)
			os.Exit(1)
		}
	},
}

// tokensOfAllGroups lists the tokens of every group the user owns.
func tokensOfAllGroups(ctx context.Context, a *app.App) ([]dto.Token, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Retrieve all groups where you are Owner")
	groups, err := a.GetAllGroups(ctx)
	if err != nil {
		spinner.Fail("Error while retrieving groups")
		return nil, fmt.Errorf("retrieving groups: %w", err)
	}
	spinner.Success(fmt.Sprintf("%d groups retrieved", len(groups)))

	spinner, _ = pterm.DefaultSpinner.Start("Retrieve tokens of all groups")
	tokens, err := a.GetTokensOfGroups(ctx, groups)
	if err != nil {
		spinner.Fail("Error while retrieving tokens of groups")
		return nil, fmt.Errorf("retrieving tokens of groups: %w", err)
	}
	spinner.Success("Tokens retrieved")
	return tokens, nil
}

// tokensOfGroup lists the tokens of the group given by --id only.
func tokensOfGroup(ctx context.Context, a *app.App) ([]dto.Token, error) {
	group, err := a.GetGroup(gitlabID)
	if err != nil {
		return nil, fmt.Errorf("retrieving group: %w", err)
	}
	tokens, err := a.GetTokensOfGroups(ctx, []*gitlab.Group{group})
	if err != nil {
		return nil, fmt.Errorf("retrieving tokens of group: %w", err)
	}
	return tokens, nil
}

// tokensOfGroupRecursive lists the tokens of the group given by --id,
// of all its subgroups and of all their projects.
func tokensOfGroupRecursive(ctx context.Context, a *app.App) ([]dto.Token, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Retrieve informations of all subgroups and projects")

	actualGroup, err := a.GetGroup(gitlabID)
	if err != nil {
		spinner.Fail("Error while retrieving group informations")
		return nil, fmt.Errorf("retrieving group: %w", err)
	}
	groups, err := a.GetSubGroups(ctx, gitlabID)
	if err != nil {
		spinner.Fail("Error while retrieving subgroups")
		return nil, fmt.Errorf("retrieving subgroups: %w", err)
	}
	groups = append([]*gitlab.Group{actualGroup}, groups...)

	projects, err := a.GetRecursiveProjectsOfGroup(ctx, gitlabID)
	if err != nil {
		spinner.Fail("Error while retrieving projects")
		return nil, fmt.Errorf("retrieving projects: %w", err)
	}
	spinner.Success("Groups and projects retrieved")

	spinner, _ = pterm.DefaultSpinner.Start("Retrieve tokens of all subgroups and projects")
	tokens, err := a.GetTokensOfGroups(ctx, groups)
	if err != nil {
		spinner.Fail("Error while retrieving tokens of groups")
		return nil, fmt.Errorf("retrieving tokens of groups: %w", err)
	}
	tokensOfProjects, err := a.GetTokensOfProjects(ctx, projects)
	if err != nil {
		spinner.Fail("Error while retrieving tokens of projects")
		return nil, fmt.Errorf("retrieving tokens of projects: %w", err)
	}
	spinner.Success("Tokens retrieved")

	return append(tokens, tokensOfProjects...), nil
}
