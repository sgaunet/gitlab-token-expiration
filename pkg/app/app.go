package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
	"github.com/sgaunet/gitlab-token-expiration/pkg/logger"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
)

type App struct {
	gitlabService *gitlab.GitlabService
	printRevoked  bool
	log           logger.Logger
	view          views.Renderer
}

// option pattern to set the gitlab endpoint
type Option func(*App)

// WithGitlabEndpoint sets the gitlab endpoint
func WithGitlabEndpoint(gitlabApiEndpoint string) Option {
	return func(a *App) {
		a.gitlabService.SetGitlabEndpoint(gitlabApiEndpoint)
	}
}

// WithRevokedToken sets the printRevoked flag
func WithRevokedToken(printRevoked bool) Option {
	return func(a *App) {
		a.printRevoked = printRevoked
	}
}

// NewApp returns a new App struct
func NewApp(v views.Renderer, opts ...Option) *App {
	app := &App{
		gitlabService: gitlab.NewGitlabService(),
		view:          v,
		log:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// SetLogger sets the logger
func (a *App) SetLogger(l logger.Logger) {
	a.log = l
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
	}
	return tokens, nil
}

// GetProject returns the project that matches the given ID
func (a *App) GetProject(projectID int) (gitlab.GitlabProject, error) {
	project, err := a.gitlabService.GetProject(projectID)
	if err != nil {
		return gitlab.GitlabProject{}, err
	}
	return project, nil
}

// GetGroup returns the group that matches the given ID
func (a *App) GetGroup(groupID int) (gitlab.GitlabGroup, error) {
	group, err := a.gitlabService.GetGroup(groupID)
	if err != nil {
		return gitlab.GitlabGroup{}, err
	}
	return group, nil
}

// GetSubGroups returns the subgroups of the group that matches the given ID
func (a *App) GetSubGroups(groupID int) ([]gitlab.GitlabGroup, error) {
	groups, err := a.gitlabService.GetSubgroups(groupID)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GetRecursiveProjectsOfGroup returns the projects of the group that matches the given ID
func (a *App) GetRecursiveProjectsOfGroup(groupID int) ([]gitlab.GitlabProject, error) {
	projects, err := a.gitlabService.GetProjectsOfGroup(groupID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetPersonalAccessTokens returns the personal access tokens
func (a *App) GetPersonalAccessTokens(ctx context.Context) ([]dto.Token, error) {
	tokens, err := a.gitlabService.GetPersonalAccessTokens()
	if err != nil {
		return nil, err
	}

	res := convertPersonalGitlabTokenToDTOTokens(tokens)
	return res, nil
}
