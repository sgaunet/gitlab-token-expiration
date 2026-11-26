// Package app contains the application logic for the GitLab token expiration tool.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/logger"
	"gitlab.com/gitlab-org/api/client-go"
)

// App represents the application with GitLab client and configuration.
type App struct {
	gitlabClient *gitlab.Client
	parallel     int
	skippedMu    sync.Mutex
	skipped      []SkippedResource
	listed       atomic.Int64 // number of token lists read successfully
	log          logger.Logger
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

// WithParallel sets the maximum number of concurrent GitLab API calls
// used to retrieve tokens. Values lower than 1 are treated as 1.
func WithParallel(parallel int) Option {
	return func(a *App) {
		a.parallel = max(parallel, 1)
	}
}

// NewApp returns a new App whose GitLab client uses the GITLAB_TOKEN and
// GITLAB_URI (optional, defaults to gitlab.com) environment variables.
// It fails when GITLAB_URI is not a valid URL.
func NewApp(opts ...Option) (*App, error) {
	token := os.Getenv("GITLAB_TOKEN")
	gitlabURI := os.Getenv("GITLAB_URI")

	var clientOpts []gitlab.ClientOptionFunc
	if gitlabURI != "" {
		clientOpts = append(clientOpts, gitlab.WithBaseURL(gitlabURI))
	}
	client, err := gitlab.NewClient(token, clientOpts...)
	if err != nil {
		if gitlabURI != "" {
			return nil, fmt.Errorf("invalid GITLAB_URI %q: %w", gitlabURI, err)
		}
		return nil, fmt.Errorf("failed to create the GitLab client: %w", err)
	}

	app := &App{
		gitlabClient: client,
		parallel:     DefaultParallel,
		log:          slog.New(slog.DiscardHandler),
	}
	for _, opt := range opts {
		opt(app)
	}
	return app, nil
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
func (a *App) GetProject(ctx context.Context, projectID int64) (*gitlab.Project, error) {
	project, _, err := a.gitlabClient.Projects.GetProject(projectID, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d: %w", projectID, err)
	}
	return project, nil
}

// GetGroup returns the group that matches the given ID.
func (a *App) GetGroup(ctx context.Context, groupID int64) (*gitlab.Group, error) {
	group, _, err := a.gitlabClient.Groups.GetGroup(groupID, nil, gitlab.WithContext(ctx))
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
// Projects of other namespaces shared with these groups are not included.
func (a *App) GetRecursiveProjectsOfGroup(ctx context.Context, groupID int64) ([]*gitlab.Project, error) {
	opts := &gitlab.ListGroupProjectsOptions{
		ListOptions:      listOptions(),
		IncludeSubGroups: new(true),
		WithShared:       new(false),
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

// GetPersonalAccessTokens returns the personal access tokens of the current
// user. Only the user's own tokens are listed, even with an administrator
// token, for which GitLab would otherwise return the tokens of every user.
func (a *App) GetPersonalAccessTokens(ctx context.Context) ([]dto.Token, error) {
	user, _, err := a.gitlabClient.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get the current user: %w", err)
	}
	opts := &gitlab.ListPersonalAccessTokensOptions{ListOptions: listOptions(), UserID: &user.ID}
	tokens, err := gitlab.ScanAndCollect(
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.PersonalAccessToken, *gitlab.Response, error) {
			return a.gitlabClient.PersonalAccessTokens.ListPersonalAccessTokens(opts, p, gitlab.WithContext(ctx))
		})
	if err != nil {
		return nil, fmt.Errorf("failed to list personal access tokens: %w", err)
	}
	return ConvertPersonalGitlabTokenToDTOTokens(tokens), nil
}

// collectTokens reads every page of one list of tokens (kind) of the group or
// project at path, and converts them with path as their source. When GitLab
// denies access to the list (401 or 403), the list is recorded as skipped and
// no token is returned: the other lists of the same group or project are
// still read.
func collectTokens[T any](
	a *App,
	kind, path string,
	list func(gitlab.PaginationOptionFunc) ([]T, *gitlab.Response, error),
	convert func([]T) []dto.Token,
) ([]dto.Token, error) {
	items, err := gitlab.ScanAndCollect(list)
	if err != nil {
		if a.skipIfDenied(kind, path, err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list %s of %s: %w", kind, path, err)
	}
	a.listed.Add(1)
	tokens := convert(items)
	for i := range tokens {
		tokens[i].Source = path
	}
	return tokens, nil
}

// getTokensOfProject reads the access tokens, then the deploy tokens, of the project.
//
//nolint:dupl // mirrors getTokensOfGroup with the project services of the client
func (a *App) getTokensOfProject(ctx context.Context, project *gitlab.Project) ([]dto.Token, error) {
	accessOpts := &gitlab.ListProjectAccessTokensOptions{ListOptions: listOptions()}
	accessTokens, err := collectTokens(a, KindProjectAccessTokens, project.PathWithNamespace,
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.ProjectAccessToken, *gitlab.Response, error) {
			return a.gitlabClient.ProjectAccessTokens.ListProjectAccessTokens(project.ID, accessOpts, p, gitlab.WithContext(ctx))
		}, ConvertProjectAccessTokenToDTOTokens)
	if err != nil {
		return nil, err
	}

	deployOpts := &gitlab.ListProjectDeployTokensOptions{ListOptions: listOptions()}
	deployTokens, err := collectTokens(a, KindProjectDeployTokens, project.PathWithNamespace,
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
			return a.gitlabClient.DeployTokens.ListProjectDeployTokens(project.ID, deployOpts, p, gitlab.WithContext(ctx))
		}, ConvertProjectDeployTokenToDTOTokens)
	if err != nil {
		return nil, err
	}
	return append(accessTokens, deployTokens...), nil
}

// getTokensOfGroup reads the access tokens, then the deploy tokens, of the group.
//
//nolint:dupl // mirrors getTokensOfProject with the group services of the client
func (a *App) getTokensOfGroup(ctx context.Context, group *gitlab.Group) ([]dto.Token, error) {
	accessOpts := &gitlab.ListGroupAccessTokensOptions{ListOptions: listOptions()}
	accessTokens, err := collectTokens(a, KindGroupAccessTokens, group.FullPath,
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.GroupAccessToken, *gitlab.Response, error) {
			return a.gitlabClient.GroupAccessTokens.ListGroupAccessTokens(group.ID, accessOpts, p, gitlab.WithContext(ctx))
		}, ConvertGroupAccessTokenToDTOTokens)
	if err != nil {
		return nil, err
	}

	deployOpts := &gitlab.ListGroupDeployTokensOptions{ListOptions: listOptions()}
	deployTokens, err := collectTokens(a, KindGroupDeployTokens, group.FullPath,
		func(p gitlab.PaginationOptionFunc) ([]*gitlab.DeployToken, *gitlab.Response, error) {
			return a.gitlabClient.DeployTokens.ListGroupDeployTokens(group.ID, deployOpts, p, gitlab.WithContext(ctx))
		}, ConvertGroupDeployTokenToDTOTokens)
	if err != nil {
		return nil, err
	}
	return append(accessTokens, deployTokens...), nil
}
