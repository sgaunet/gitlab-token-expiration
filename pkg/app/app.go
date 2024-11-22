package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
)

type App struct {
	gitlabService *gitlab.GitlabService
	log           Logger
	view          views.Renderer
}

type Logger interface {
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Info(msg string, args ...any)
}

// NewApp returns a new App struct
func NewApp(v views.Renderer) *App {
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

// GetTokensOfProject returns the tokens of a project
func (a *App) GetTokensOfProjects(ctx context.Context, projects []gitlab.GitlabProject) ([]dto.Token, error) {
	var tokens []dto.Token

	for _, project := range projects {
		projectAccessTokens, err := a.gitlabService.GetProjectAccessTokens(project.Id)
		if err != nil {
			return nil, err
		}
		dtoTokens := ConvertProjectAccessTokenToDTOTokens(projectAccessTokens)

		// Add the source
		for i := range dtoTokens {
			dtoTokens[i].Source = project.PathWithNamespace
		}
		tokens = append(tokens, dtoTokens...)
	}
	return tokens, nil
}

// GetTokensOfGroups returns the tokens of all groups
func (a *App) GetTokensOfGroups(ctx context.Context, groups []gitlab.GitlabGroup) ([]dto.Token, error) {
	var tokens []dto.Token

	for _, group := range groups {
		// Get access tokens of the group
		groupAccessTokens, err := a.gitlabService.GetGroupAccessTokens(group.Id)
		if err != nil {
			return nil, err
		}
		dtoTokens := ConvertGroupAccessTokenToDTOTokens(groupAccessTokens)
		// Add the source
		for i := range dtoTokens {
			dtoTokens[i].Source = group.Path
		}
		tokens = append(tokens, dtoTokens...)

		// if len(groupAccessTokens) != 0 {
		// 	fmt.Printf("\n%s\n\n",
		// 		termlink.Link(fmt.Sprintf("# Group Access Tokens - %s", group.Path),
		// 			fmt.Sprintf("%s/-/settings/access_tokens", group.WebUrl)))
		// 	_ = a.view.PrintGitlabGroupAccessToken(groupAccessTokens)
		// }
		// Get deploy tokens of the group
		groupDeployTokens, err := a.gitlabService.GetGroupDeployTokens(group.Id)
		if err != nil {
			return nil, err
		}
		dtoTokens = ConvertGroupDeployTokenToDTOTokens(groupDeployTokens)
		// Add the source
		for i := range dtoTokens {
			dtoTokens[i].Source = group.Path
		}
		tokens = append(tokens, dtoTokens...)
		// if len(groupDeployTokens) != 0 {
		// 	fmt.Printf("\n# Group Deploy Tokens: %s\n\n", group.Name)
		// 	fmt.Printf("\n%s\n\n", termlink.Link(fmt.Sprintf("# Group Deploy Tokens - %s", group.Path),
		// 		fmt.Sprintf("%s/-/settings/repository", group.WebUrl)))
		// 	_ = a.view.PrintGitlabGroupDeployToken(groupDeployTokens)
		// }
	}
	return tokens, nil
}

// 	// Get projects of the group
// 	projects, err := a.gitlabService.GetProjectsOfGroup(group.Id)
// 	if err != nil {
// 		return err
// 	}
// 	for project := range projects {
// 		if !projects[project].Archived {
// 			err = a.ListTokensOfProject(ctx, projects[project])
// 			if err != nil {
// 				a.log.Error("error occured during backup", "project name", projects[project].Name, "error", err.Error())
// 				return err
// 			}
// 		} else {
// 			a.log.Info("project is archived, skip", "project name", projects[project].Name)
// 		}
// 	}
// 	return nil
// }

// func (a *App) RenderTokensOfProjects(ctx context.Context, projects []gitlab.GitlabProject) error {
// 	tokens, err := a.GetTokensOfProjects(ctx, projects)
// 	if err != nil {
// 		return err
// 	}
// 	if len(tokens) != 0 {
// 		// fmt.Printf("\n%s\n\n",
// 		// 	termlink.Link(fmt.Sprintf("# Project Access Tokens - %s", project.Name),
// 		// 		fmt.Sprintf("%s/-/settings/access_tokens", project.WebUrl)))
// 		a.view.Render(tokens)
// 	}
// 	return nil
// }

// func (a *App) ListPersonalAccessTokens(ctx context.Context) error {
// 	tokens, err := a.gitlabService.GetPersonalAccessTokens()
// 	if err != nil {
// 		return err
// 	}
// 	if len(tokens) != 0 {
// 		fmt.Printf("\n%s\n\n", termlink.Link("# Personal Access Tokens", "https://gitlab.com/-/user_settings/personal_access_tokens"))
// 		a.view.PrintGitlabPersonalAccessToken(tokens)
// 	}
// 	return nil
// }

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

func (a *App) GetSubGroups(groupID int) ([]gitlab.GitlabGroup, error) {
	groups, err := a.gitlabService.GetSubgroups(groupID)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (a *App) GetRecursiveProjectsOfGroup(groupID int) ([]gitlab.GitlabProject, error) {
	projects, err := a.gitlabService.GetProjectsOfGroup(groupID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (a *App) GetPersonalAccessTokens(ctx context.Context) ([]dto.Token, error) {
	tokens, err := a.gitlabService.GetPersonalAccessTokens()
	if err != nil {
		return nil, err
	}
	// if len(tokens) != 0 {
	// 	fmt.Printf("\n%s\n\n", termlink.Link("# Personal Access Tokens", "https://gitlab.com/-/user_settings/personal_access_tokens"))
	// 	a.view.PrintGitlabPersonalAccessToken(tokens)
	// }
	res := convertPersonalGitlabTokenToDTOTokens(tokens)
	if err != nil {
		return nil, err
	}
	return res, nil
}
