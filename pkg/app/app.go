package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/logger"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"gitlab.com/gitlab-org/api/client-go"
)

type App struct {
	gitlabClient  *gitlab.Client
	printRevoked  bool
	log           logger.Logger
	view          views.Renderer
}

// option pattern to set the gitlab endpoint
type Option func(*App)

// WithGitlabEndpoint sets the gitlab endpoint
func WithGitlabEndpoint(gitlabApiEndpoint string) Option {
	return func(a *App) {
		// Create new client with custom base URL
		token := os.Getenv("GITLAB_TOKEN")
		client, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabApiEndpoint))
		if err == nil {
			a.gitlabClient = client
		}
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
	token := os.Getenv("GITLAB_TOKEN")
	client, err := gitlab.NewClient(token)
	if err != nil {
		// Handle error or use a fallback
		client = nil
	}

	app := &App{
		gitlabClient: client,
		view:         v,
		log:          slog.New(slog.NewTextHandler(io.Discard, nil)),
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
	// Create new client with custom base URL
	token := os.Getenv("GITLAB_TOKEN")
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabApiEndpoint))
	if err == nil {
		a.gitlabClient = client
	}
}

// SetToken sets the gitlab token
func (a *App) SetToken(token string) {
	// Create new client with the provided token
	client, err := gitlab.NewClient(token)
	if err == nil {
		a.gitlabClient = client
	}
}

// SetHttpClient sets the http client
func (a *App) SetHttpClient(httpClient *http.Client) {
	// Create new client with custom HTTP client
	token := os.Getenv("GITLAB_TOKEN")
	client, err := gitlab.NewClient(token, gitlab.WithHTTPClient(httpClient))
	if err == nil {
		a.gitlabClient = client
	}
}

// GetTokensOfProject returns the tokens of a project
func (a *App) GetTokensOfProjects(ctx context.Context, projects []*gitlab.Project) ([]dto.Token, error) {
	var tokens []dto.Token

	for _, project := range projects {
		projectAccessTokens, _, err := a.gitlabClient.ProjectAccessTokens.ListProjectAccessTokens(project.ID, nil)
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
func (a *App) GetTokensOfGroups(ctx context.Context, groups []*gitlab.Group) ([]dto.Token, error) {
	var tokens []dto.Token

	for _, group := range groups {
		// Get access tokens of the group
		groupAccessTokens, _, err := a.gitlabClient.GroupAccessTokens.ListGroupAccessTokens(group.ID, nil)
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
		groupDeployTokens, _, err := a.gitlabClient.DeployTokens.ListGroupDeployTokens(group.ID, nil)
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
func (a *App) GetProject(projectID int) (*gitlab.Project, error) {
	project, _, err := a.gitlabClient.Projects.GetProject(projectID, nil)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// GetGroup returns the group that matches the given ID
func (a *App) GetGroup(groupID int) (*gitlab.Group, error) {
	group, _, err := a.gitlabClient.Groups.GetGroup(groupID, nil)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// GetSubGroups returns the subgroups of the group that matches the given ID
func (a *App) GetSubGroups(groupID int) ([]*gitlab.Group, error) {
	groups, _, err := a.gitlabClient.Groups.ListSubGroups(groupID, nil)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GetRecursiveProjectsOfGroup returns the projects of the group that matches the given ID
func (a *App) GetRecursiveProjectsOfGroup(groupID int) ([]*gitlab.Project, error) {
	// Get projects of the group
	projects, _, err := a.gitlabClient.Groups.ListGroupProjects(groupID, nil)
	if err != nil {
		return nil, err
	}

	// Get subgroups recursively
	subgroups, err := a.GetSubGroups(groupID)
	if err != nil {
		return projects, err // Return what we have so far
	}

	// Get projects from each subgroup recursively
	for _, subgroup := range subgroups {
		subProjects, err := a.GetRecursiveProjectsOfGroup(subgroup.ID)
		if err != nil {
			continue // Skip this subgroup on error
		}
		projects = append(projects, subProjects...)
	}

	return projects, nil
}

// GetPersonalAccessTokens returns the personal access tokens
func (a *App) GetPersonalAccessTokens(ctx context.Context) ([]dto.Token, error) {
	tokens, _, err := a.gitlabClient.PersonalAccessTokens.ListPersonalAccessTokens(nil)
	if err != nil {
		return nil, err
	}

	res := convertPersonalGitlabTokenToDTOTokens(tokens)
	return res, nil
}
