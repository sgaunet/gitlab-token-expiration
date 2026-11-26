package cmd

import (
	"context"
	"fmt"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/api/client-go"
)

// projectCmd represents the project command to list expirable tokens of a project.
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "List expirable tokens of a project, or of all your projects",
	Long: `List expirable tokens (access tokens and deploy tokens) of a project.

With --id, list the tokens of that project.

Without --id, list the tokens of every project where you have at least the
Maintainer role (the role GitLab requires to list project tokens).`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		cmd.SilenceUsage = true
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
		if gitlabID == 0 {
			tokens, err = tokensOfAllProjects(ctx, a)
		} else {
			tokens, err = tokensOfProject(ctx, a)
		}
		if err != nil {
			return err
		}
		return out.render(a, tokens)
	},
}

// tokensOfAllProjects lists the tokens of every project the user maintains.
func tokensOfAllProjects(ctx context.Context, a *app.App) ([]dto.Token, error) {
	spinner := startProgress("Retrieve all projects where you are at least Maintainer")
	projects, err := a.GetAllProjects(ctx)
	if err != nil {
		spinner.Fail("Error while retrieving projects")
		return nil, fmt.Errorf("retrieving projects: %w", err)
	}
	spinner.Success(fmt.Sprintf("%d projects retrieved", len(projects)))

	spinner = startProgress("Retrieve tokens of all projects")
	tokens, err := a.GetTokensOfProjects(ctx, projects)
	if err != nil {
		spinner.Fail("Error while retrieving tokens of projects")
		return nil, fmt.Errorf("retrieving tokens of projects: %w", err)
	}
	spinner.Success("Tokens retrieved")
	return tokens, nil
}

// tokensOfProject lists the tokens of the project given by --id.
func tokensOfProject(ctx context.Context, a *app.App) ([]dto.Token, error) {
	project, err := a.GetProject(ctx, gitlabID)
	if err != nil {
		return nil, fmt.Errorf("retrieving project: %w", err)
	}
	tokens, err := a.GetTokensOfProjects(ctx, []*gitlab.Project{project})
	if err != nil {
		return nil, fmt.Errorf("retrieving tokens of project: %w", err)
	}
	return tokens, nil
}
