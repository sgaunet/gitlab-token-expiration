package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/savioxavier/termlink"
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
)

type App struct {
	gitlabService *gitlab.GitlabService
	log           Logger
	view          views.Printer
}

type Logger interface {
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

// NewApp returns a new App struct
func NewApp(v views.Printer) *App {
	app := &App{
		gitlabService: gitlab.NewGitlabService(),
		view:          v,
		log:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	gitlab.SetLogger(app.log)
	return app
}

// SetLogger sets the logger
func (a *App) SetLogger(l Logger) {
	a.log = l
	gitlab.SetLogger(l)
}

// SetGitlabEndpoint sets the gitlab endpoint
func (a *App) SetGitlabEndpoint(gitlabApiEndpoint string) {
	a.gitlabService.SetGitlabEndpoint(gitlabApiEndpoint)
}

// SetToken sets the gitlab token
func (a *App) SetToken(token string) {
	a.gitlabService.SetToken(token)
}

// SetHttpClient sets the http client
func (a *App) SetHttpClient(httpClient *http.Client) {
	a.gitlabService.SetHttpClient(httpClient)
}

func (a *App) ListTokensOfGroup(ctx context.Context, group gitlab.GitlabGroup) error {
	// Get group
	// group, err := a.gitlabService.GetGroup(gitlabGroupID)
	// if err != nil {
	// 	return err
	// }
	// Get access tokens of the group
	groupAccessTokens, err := a.gitlabService.GetGroupAccessTokens(group.Id)
	if err != nil {
		return err
	}
	if len(groupAccessTokens) != 0 {
		fmt.Printf("\n%s\n\n",
			termlink.Link(fmt.Sprintf("# Group Access Tokens - %s", group.Path),
				fmt.Sprintf("%s/-/settings/access_tokens", group.WebUrl)))
		_ = a.view.PrintGitlabGroupAccessToken(groupAccessTokens)
	}
	// Get deploy tokens of the group
	groupDeployTokens, err := a.gitlabService.GetGroupDeployTokens(group.Id)
	if err != nil {
		return err
	}
	if len(groupDeployTokens) != 0 {
		fmt.Printf("\n# Group Deploy Tokens: %s\n\n", group.Name)
		fmt.Printf("\n%s\n\n", termlink.Link(fmt.Sprintf("# Group Deploy Tokens - %s", group.Path),
			fmt.Sprintf("%s/-/settings/repository", group.WebUrl)))
		_ = a.view.PrintGitlabGroupDeployToken(groupDeployTokens)
	}

	// Get projects of the group
	projects, err := a.gitlabService.GetProjectsOfGroup(group.Id)
	if err != nil {
		return err
	}
	for project := range projects {
		if !projects[project].Archived {
			err = a.ListTokensOfProject(ctx, projects[project])
			if err != nil {
				a.log.Error("error occured during backup", "project name", projects[project].Name, "error", err.Error())
				return err
			}
		} else {
			a.log.Info("project is archived, skip", "project name", projects[project].Name)
		}
	}
	return nil
}

func (a *App) ListTokensOfProject(ctx context.Context, project gitlab.GitlabProject) error {
	project, err := a.gitlabService.GetProject(project.Id)
	if err != nil {
		return err
	}

	tokens, err := a.gitlabService.GetProjectAccessTokens(project.Id)
	if err != nil {
		return err
	}
	if len(tokens) != 0 {
		fmt.Printf("\n%s\n\n",
			termlink.Link(fmt.Sprintf("# Project Access Tokens - %s", project.Name),
				fmt.Sprintf("%s/-/settings/access_tokens", project.WebUrl)))
		a.view.PrintGitlabProjectAccessToken(tokens)
	}
	return nil
}

func (a *App) ListPersonalAccessTokens(ctx context.Context) error {
	tokens, err := a.gitlabService.GetPersonalAccessTokens()
	if err != nil {
		return err
	}
	if len(tokens) != 0 {
		fmt.Printf("\n%s\n\n", termlink.Link("# Personal Access Tokens", "https://gitlab.com/-/user_settings/personal_access_tokens"))
		a.view.PrintGitlabPersonalAccessToken(tokens)
	}
	return nil
}

func (a *App) GetProject(projectID int) (gitlab.GitlabProject, error) {
	project, err := a.gitlabService.GetProject(projectID)
	if err != nil {
		return gitlab.GitlabProject{}, err
	}
	return project, nil
}

func (a *App) GetGroup(groupID int) (gitlab.GitlabGroup, error) {
	group, err := a.gitlabService.GetGroup(groupID)
	if err != nil {
		return gitlab.GitlabGroup{}, err
	}
	return group, nil
}
