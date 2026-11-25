// Package app contains the application logic for the GitLab token expiration tool.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/logger"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"gitlab.com/gitlab-org/api/client-go"
)

// App represents the application with GitLab client and configuration.
type App struct {
	gitlabClient *gitlab.Client
	printRevoked bool
	parallel     int
	skipDenied   bool
	skippedMu    sync.Mutex
	skipped      []SkippedResource
	log          logger.Logger
	view         views.Renderer
}

// Option is a function that configures the App.
type Option func(*App)

// WithGitlabEndpoint sets the gitlab endpoint.
func WithGitlabEndpoint(gitlabAPIEndpoint string) Option {
	return func(a *App) {
		// Create new client with custom base URL
		token := os.Getenv("GITLAB_TOKEN")
		client, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabAPIEndpoint))
		if err == nil {
			a.gitlabClient = client
		}
	}
}

// WithRevokedToken sets the printRevoked flag.
func WithRevokedToken(printRevoked bool) Option {
	return func(a *App) {
		a.printRevoked = printRevoked
	}
}

// WithParallel sets the maximum number of concurrent GitLab API calls
// used to retrieve tokens. Values lower than 1 are treated as 1.
func WithParallel(parallel int) Option {
	return func(a *App) {
		a.parallel = max(parallel, 1)
	}
}

// NewApp returns a new App struct.
func NewApp(v views.Renderer, opts ...Option) *App {
	token := os.Getenv("GITLAB_TOKEN")
	gitlabURI := os.Getenv("GITLAB_URI")

	var client *gitlab.Client
	var err error

	// If GITLAB_URI is set, use it as the base URL
	if gitlabURI != "" {
		client, err = gitlab.NewClient(token, gitlab.WithBaseURL(gitlabURI))
	} else {
		client, err = gitlab.NewClient(token)
	}

	if err != nil {
		// Handle error or use a fallback
		client = nil
	}

	app := &App{
		gitlabClient: client,
		view:         v,
		parallel:     DefaultParallel,
		log:          slog.New(slog.DiscardHandler),
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// SetLogger sets the logger.
func (a *App) SetLogger(l logger.Logger) {
	a.log = l
}

// SetGitlabEndpoint sets the gitlab endpoint.
func (a *App) SetGitlabEndpoint(gitlabAPIEndpoint string) {
	// Create new client with custom base URL
	token := os.Getenv("GITLAB_TOKEN")
	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(gitlabAPIEndpoint))
	if err == nil {
		a.gitlabClient = client
	}
}

// SetToken sets the gitlab token.
func (a *App) SetToken(token string) {
	// Create new client with the provided token
	client, err := gitlab.NewClient(token)
	if err == nil {
		a.gitlabClient = client
	}
}

// SetHTTPClient sets the http client.
func (a *App) SetHTTPClient(httpClient *http.Client) {
	// Create new client with custom HTTP client
	token := os.Getenv("GITLAB_TOKEN")
	client, err := gitlab.NewClient(token, gitlab.WithHTTPClient(httpClient))
	if err == nil {
		a.gitlabClient = client
	}
}

// perPage is the page size used for paginated GitLab API calls (the maximum allowed).
const perPage = 100

// listOptions returns the pagination options used for every list call.
func listOptions() gitlab.ListOptions {
	return gitlab.ListOptions{PerPage: perPage}
}

// GetTokensOfProjects returns the access tokens of the given projects.
// Projects are processed concurrently, see WithParallel.
func (a *App) GetTokensOfProjects(ctx context.Context, projects []*gitlab.Project) ([]dto.Token, error) {
	parts, err := parallelMap(ctx, a.parallel, projects, a.getTokensOfProject)
	if err != nil {
		return nil, err
	}
	return flatten(parts), nil
}

// GetTokensOfGroups returns the access tokens and deploy tokens of the given groups.
// Groups are processed concurrently, see WithParallel.
func (a *App) GetTokensOfGroups(ctx context.Context, groups []*gitlab.Group) ([]dto.Token, error) {
	parts, err := parallelMap(ctx, a.parallel, groups, a.getTokensOfGroup)
	if err != nil {
		return nil, err
	}
	return flatten(parts), nil
}

// GetProject returns the project that matches the given ID.
func (a *App) GetProject(projectID int64) (*gitlab.Project, error) {
	project, _, err := a.gitlabClient.Projects.GetProject(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d: %w", projectID, err)
	}
	return project, nil
}

// GetGroup returns the group that matches the given ID.
func (a *App) GetGroup(groupID int64) (*gitlab.Group, error) {
	group, _, err := a.gitlabClient.Groups.GetGroup(groupID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group %d: %w", groupID, err)
	}
	return group, nil
}

// GetSubGroups returns all the descendant groups (subgroups at any depth)
// of the group that matches the given ID.
func (a *App) GetSubGroups(ctx context.Context, groupID int64) ([]*gitlab.Group, error) {
	opts := &gitlab.ListDescendantGroupsOptions{ListOptions: listOptions()}
	groups, err := gitlab.ScanAndCollect(
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.Group, *gitlab.Response, error) {
			return a.gitlabClient.Groups.ListDescendantGroups(groupID, opts, p, gitlab.WithContext(ctx))
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list subgroups for group %d: %w", groupID, err)
	}
	return groups, nil
}

// GetRecursiveProjectsOfGroup returns the projects of the group that matches
// the given ID, including the projects of its subgroups at any depth.
func (a *App) GetRecursiveProjectsOfGroup(ctx context.Context, groupID int64) ([]*gitlab.Project, error) {
	opts := &gitlab.ListGroupProjectsOptions{
		ListOptions:      listOptions(),
		IncludeSubGroups: new(true),
	}
	projects, err := gitlab.ScanAndCollect(
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.Project, *gitlab.Response, error) {
			return a.gitlabClient.Groups.ListGroupProjects(groupID, opts, p, gitlab.WithContext(ctx))
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects for group %d: %w", groupID, err)
	}
	return projects, nil
}

// GetAllGroups returns every group where the current user has the Owner role,
// which is the role required to list group access tokens and deploy tokens.
func (a *App) GetAllGroups(ctx context.Context) ([]*gitlab.Group, error) {
	opts := &gitlab.ListGroupsOptions{
		ListOptions:    listOptions(),
		MinAccessLevel: gitlab.Ptr(gitlab.OwnerPermissions),
	}
	groups, err := gitlab.ScanAndCollect(
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.Group, *gitlab.Response, error) {
			return a.gitlabClient.Groups.ListGroups(opts, p, gitlab.WithContext(ctx))
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	return groups, nil
}

// GetAllProjects returns every project where the current user has at least the
// Maintainer role, which is the role required to list project access tokens.
func (a *App) GetAllProjects(ctx context.Context) ([]*gitlab.Project, error) {
	opts := &gitlab.ListProjectsOptions{
		ListOptions:    listOptions(),
		MinAccessLevel: gitlab.Ptr(gitlab.MaintainerPermissions),
		Simple:         new(true),
	}
	projects, err := gitlab.ScanAndCollect(
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.Project, *gitlab.Response, error) {
			return a.gitlabClient.Projects.ListProjects(opts, p, gitlab.WithContext(ctx))
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

// GetPersonalAccessTokens returns the personal access tokens.
func (a *App) GetPersonalAccessTokens(_ context.Context) ([]dto.Token, error) {
	tokens, _, err := a.gitlabClient.PersonalAccessTokens.ListPersonalAccessTokens(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list personal access tokens: %w", err)
	}

	res := ConvertPersonalGitlabTokenToDTOTokens(tokens)
	return res, nil
}

func (a *App) getTokensOfProject(ctx context.Context, project *gitlab.Project) ([]dto.Token, error) {
	opts := &gitlab.ListProjectAccessTokensOptions{ListOptions: listOptions()}
	projectAccessTokens, err := gitlab.ScanAndCollect(
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.ProjectAccessToken, *gitlab.Response, error) {
			return a.gitlabClient.ProjectAccessTokens.ListProjectAccessTokens(project.ID, opts, p, gitlab.WithContext(ctx))
		})
	if err != nil {
		if a.skipIfDenied("project", project.PathWithNamespace, err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list project access tokens for project %d: %w", project.ID, err)
	}
	tokens := ConvertProjectAccessTokenToDTOTokens(projectAccessTokens)
	for i := range tokens {
		tokens[i].Source = project.PathWithNamespace
	}
	return tokens, nil
}

func (a *App) getTokensOfGroup(ctx context.Context, group *gitlab.Group) ([]dto.Token, error) {
	gatOpts := &gitlab.ListGroupAccessTokensOptions{ListOptions: listOptions()}
	groupAccessTokens, err := gitlab.ScanAndCollect(
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.GroupAccessToken, *gitlab.Response, error) {
			return a.gitlabClient.GroupAccessTokens.ListGroupAccessTokens(group.ID, gatOpts, p, gitlab.WithContext(ctx))
		})
	if err != nil {
		if a.skipIfDenied("group", group.FullPath, err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list group access tokens for group %d: %w", group.ID, err)
	}

	deployOpts := &gitlab.ListGroupDeployTokensOptions{ListOptions: listOptions()}
	groupDeployTokens, err := gitlab.ScanAndCollect(
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
			return a.gitlabClient.DeployTokens.ListGroupDeployTokens(group.ID, deployOpts, p, gitlab.WithContext(ctx))
		})
	if err != nil {
		if a.skipIfDenied("group", group.FullPath, err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list group deploy tokens for group %d: %w", group.ID, err)
	}

	tokens := ConvertGroupAccessTokenToDTOTokens(groupAccessTokens)
	tokens = append(tokens, ConvertGroupDeployTokenToDTOTokens(groupDeployTokens)...)
	for i := range tokens {
		tokens[i].Source = group.FullPath
	}
	return tokens, nil
}
