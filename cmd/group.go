// Package cmd contains the command-line interface for the GitLab token expiration tool.
package cmd

import (
	"context"
	"fmt"

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
(the role GitLab requires to list group access tokens). Their subgroups and
projects are not included: use the project command for project tokens.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cmd.SilenceUsage = true
		if noRecursiveOption && gitlabID == 0 {
			return errNoRecursiveWithoutID
		}
		out, err := prepare()
		if err != nil {
			return err
		}
		a, err := newApp()
		if err != nil {
			return err
		}
		ctx := cmd.Context()

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
			return err
		}
		return out.render(a, tokens)
	},
}

// tokensOfAllGroups lists the tokens of every group the user owns.
func tokensOfAllGroups(ctx context.Context, a *app.App) ([]dto.Token, error) {
	spinner := startProgress("Retrieve all groups where you are Owner")
	groups, err := a.GetAllGroups(ctx)
	if err != nil {
		spinner.Fail("Error while retrieving groups")
		return nil, fmt.Errorf("retrieving groups: %w", err)
	}
	spinner.Success(fmt.Sprintf("%d groups retrieved", len(groups)))

	spinner = startProgress("Retrieve tokens of all groups")
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
	group, err := a.GetGroup(ctx, gitlabID)
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
	spinner := startProgress("Retrieve informations of all subgroups and projects")

	actualGroup, err := a.GetGroup(ctx, gitlabID)
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

	spinner = startProgress("Retrieve tokens of all subgroups and projects")
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
