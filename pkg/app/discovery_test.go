package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeGitLab starts a fake GitLab API serving the given JSON responses,
// keyed by "path?page" ("" page means the first one). The "X-Next-Page"
// header is set when a response for the next page exists.
func fakeGitLab(t *testing.T, routes map[string]any, check func(r *http.Request)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}
		page := r.URL.Query().Get("page")
		key := r.URL.Path
		if page != "" && page != "1" {
			key += "?" + page
		}
		body, ok := routes[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if status, isStatus := body.(int); isStatus {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"message":"error"}`))
			return
		}
		next := "2"
		if n, err := strconv.Atoi(page); err == nil {
			next = strconv.Itoa(n + 1)
		}
		if _, hasNext := routes[r.URL.Path+"?"+next]; hasNext {
			w.Header().Set("X-Next-Page", next)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestApp(t *testing.T, srv *httptest.Server, parallel int, opts ...app.Option) *app.App {
	t.Helper()
	t.Setenv("GITLAB_TOKEN", "test-token")
	t.Setenv("GITLAB_URI", "")
	opts = append([]app.Option{app.WithGitlabEndpoint(srv.URL), app.WithParallel(parallel)}, opts...)
	a, err := app.NewApp(opts...)
	require.NoError(t, err)
	return a
}

// hitCounter counts the requests per path, safely for concurrent requests.
type hitCounter struct {
	mu   sync.Mutex
	hits map[string]int
}

func (h *hitCounter) record(r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.hits == nil {
		h.hits = map[string]int{}
	}
	h.hits[r.URL.Path]++
}

func (h *hitCounter) count(path string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.hits[path]
}

type obj = map[string]any

func TestApp_GetAllGroups_PaginatesAndFiltersOnOwner(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/groups":   []obj{{"id": 1, "full_path": "a"}},
		"/api/v4/groups?2": []obj{{"id": 2, "full_path": "a/b"}},
	}, func(r *http.Request) {
		assert.Equal(t, "50", r.URL.Query().Get("min_access_level"))
		assert.Equal(t, "100", r.URL.Query().Get("per_page"))
	})

	groups, err := newTestApp(t, srv, 2).GetAllGroups(context.Background())

	require.NoError(t, err)
	require.Len(t, groups, 2)
	assert.Equal(t, "a", groups[0].FullPath)
	assert.Equal(t, "a/b", groups[1].FullPath)
}

func TestApp_GetAllProjects_FiltersOnMaintainer(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/projects": []obj{{"id": 7, "path_with_namespace": "a/p"}},
	}, func(r *http.Request) {
		assert.Equal(t, "40", r.URL.Query().Get("min_access_level"))
		assert.Equal(t, "true", r.URL.Query().Get("simple"))
	})

	projects, err := newTestApp(t, srv, 2).GetAllProjects(context.Background())

	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "a/p", projects[0].PathWithNamespace)
}

func TestApp_GetTokensOfGroups_AllGroupsInOrder(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/groups":                   []obj{{"id": 1, "full_path": "a"}, {"id": 2, "full_path": "a/b"}},
		"/api/v4/groups/1/access_tokens":   []obj{{"id": 11, "name": "gat-a", "expires_at": "2027-01-01"}},
		"/api/v4/groups/1/deploy_tokens":   []obj{},
		"/api/v4/groups/2/access_tokens":   []obj{{"id": 21, "name": "gat-b"}},
		"/api/v4/groups/2/access_tokens?2": []obj{{"id": 23, "name": "gat-b-page2"}},
		"/api/v4/groups/2/deploy_tokens":   []obj{{"id": 22, "name": "deploy-b", "expires_at": "2027-02-01T00:00:00Z"}},
	}, nil)
	a := newTestApp(t, srv, 4)

	groups, err := a.GetAllGroups(context.Background())
	require.NoError(t, err)
	tokens, err := a.GetTokensOfGroups(context.Background(), groups)

	require.NoError(t, err)
	assert.Equal(t, []dto.Token{
		{Source: "a", Type: dto.TypeGroupAccessToken, ID: 11, Name: "gat-a", ExpiresAt: "2027-01-01"},
		{Source: "a/b", Type: dto.TypeGroupAccessToken, ID: 21, Name: "gat-b"},
		{Source: "a/b", Type: dto.TypeGroupAccessToken, ID: 23, Name: "gat-b-page2"},
		{Source: "a/b", Type: dto.TypeGroupDeployToken, ID: 22, Name: "deploy-b", ExpiresAt: "2027-02-01"},
	}, tokens)
}

func TestApp_GetTokensOfProjects_AccessAndDeployTokens(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/projects":                 []obj{{"id": 7, "path_with_namespace": "a/p"}, {"id": 8, "path_with_namespace": "a/q"}},
		"/api/v4/projects/7/access_tokens": []obj{{"id": 71, "name": "pat-p", "expires_at": "2027-01-01"}},
		"/api/v4/projects/7/deploy_tokens": []obj{{"id": 72, "name": "deploy-p", "expires_at": "2027-02-01T00:00:00Z"}},
		"/api/v4/projects/8/access_tokens": []obj{},
		"/api/v4/projects/8/deploy_tokens": []obj{{"id": 82, "name": "deploy-q"}},
	}, nil)
	a := newTestApp(t, srv, 2)

	projects, err := a.GetAllProjects(context.Background())
	require.NoError(t, err)
	tokens, err := a.GetTokensOfProjects(context.Background(), projects)

	require.NoError(t, err)
	assert.Equal(t, []dto.Token{
		{Source: "a/p", Type: dto.TypeProjectAccessToken, ID: 71, Name: "pat-p", ExpiresAt: "2027-01-01"},
		{Source: "a/p", Type: dto.TypeProjectDeployToken, ID: 72, Name: "deploy-p", ExpiresAt: "2027-02-01"},
		{Source: "a/q", Type: dto.TypeProjectDeployToken, ID: 82, Name: "deploy-q"},
	}, tokens)
	assert.Equal(t, 4, a.Listed())
	assert.Empty(t, a.Skipped())
}

func TestApp_GetTokensOfProjects_OtherErrorsAreReturned(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/projects":                 []obj{{"id": 7, "path_with_namespace": "a/p"}, {"id": 8, "path_with_namespace": "a/q"}},
		"/api/v4/projects/7/access_tokens": []obj{{"id": 71, "name": "pat-p"}},
		"/api/v4/projects/7/deploy_tokens": []obj{},
		"/api/v4/projects/8/access_tokens": http.StatusNotFound,
	}, nil)
	a := newTestApp(t, srv, 2)

	projects, err := a.GetAllProjects(context.Background())
	require.NoError(t, err)
	tokens, err := a.GetTokensOfProjects(context.Background(), projects)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "project access tokens of a/q")
	assert.Nil(t, tokens)
	assert.Empty(t, a.Skipped(), "only 401 and 403 are skipped")
}

// A denied token list must not hide the other token list of the same group:
// group access tokens need the Owner role, group deploy tokens only Maintainer.
func TestApp_DeniedTokenListIsSkippedAlone(t *testing.T) {
	var hits hitCounter
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/groups": []obj{
			{"id": 1, "full_path": "a"}, {"id": 2, "full_path": "b"}, {"id": 3, "full_path": "c"},
		},
		"/api/v4/groups/1/access_tokens": []obj{{"id": 11, "name": "gat-a"}},
		"/api/v4/groups/1/deploy_tokens": []obj{},
		"/api/v4/groups/2/access_tokens": http.StatusUnauthorized,
		"/api/v4/groups/2/deploy_tokens": []obj{{"id": 22, "name": "deploy-b"}},
		"/api/v4/groups/3/access_tokens": []obj{{"id": 31, "name": "gat-c"}},
		"/api/v4/groups/3/deploy_tokens": http.StatusForbidden,
	}, hits.record)
	a := newTestApp(t, srv, 3)

	groups, err := a.GetAllGroups(context.Background())
	require.NoError(t, err)
	tokens, err := a.GetTokensOfGroups(context.Background(), groups)

	require.NoError(t, err)
	assert.Equal(t, 1, hits.count("/api/v4/groups/2/deploy_tokens"), "deploy tokens must be requested after a denied access token list")
	assert.Equal(t, []dto.Token{
		{Source: "a", Type: dto.TypeGroupAccessToken, ID: 11, Name: "gat-a"},
		{Source: "b", Type: dto.TypeGroupDeployToken, ID: 22, Name: "deploy-b"},
		{Source: "c", Type: dto.TypeGroupAccessToken, ID: 31, Name: "gat-c"},
	}, tokens)
	skipped := a.Skipped()
	require.Len(t, skipped, 2)
	assert.Equal(t, app.KindGroupAccessTokens, skipped[0].Kind)
	assert.Equal(t, "b", skipped[0].Path)
	assert.Equal(t, app.KindGroupDeployTokens, skipped[1].Kind)
	assert.Equal(t, "c", skipped[1].Path)
	assert.Error(t, skipped[0].Err)
	assert.Equal(t, 4, a.Listed())
}

func TestApp_AllTokenListsDenied_NothingListed(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/projects":                 []obj{{"id": 7, "path_with_namespace": "a/p"}},
		"/api/v4/projects/7/access_tokens": http.StatusForbidden,
		"/api/v4/projects/7/deploy_tokens": http.StatusForbidden,
	}, nil)
	a := newTestApp(t, srv, 2)

	projects, err := a.GetAllProjects(context.Background())
	require.NoError(t, err)
	tokens, err := a.GetTokensOfProjects(context.Background(), projects)

	require.NoError(t, err)
	assert.Empty(t, tokens)
	assert.Equal(t, 0, a.Listed())
	assert.Len(t, a.Skipped(), 2)
}

// Personal access tokens are paginated and restricted to the current user:
// with an administrator token, GitLab would return the tokens of every user.
func TestApp_GetPersonalAccessTokens_CurrentUserAndPaginated(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/user":                     obj{"id": 7, "username": "me"},
		"/api/v4/personal_access_tokens":   []obj{{"id": 1, "name": "p1"}},
		"/api/v4/personal_access_tokens?2": []obj{{"id": 2, "name": "p2", "expires_at": "2027-01-01"}},
	}, func(r *http.Request) {
		if r.URL.Path == "/api/v4/personal_access_tokens" {
			assert.Equal(t, "7", r.URL.Query().Get("user_id"))
			assert.Equal(t, "100", r.URL.Query().Get("per_page"))
		}
	})

	tokens, err := newTestApp(t, srv, 2).GetPersonalAccessTokens(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []dto.Token{
		{Type: dto.TypePersonalAccessToken, ID: 1, Name: "p1"},
		{Type: dto.TypePersonalAccessToken, ID: 2, Name: "p2", ExpiresAt: "2027-01-01"},
	}, tokens)
}

func TestApp_GetRecursiveProjectsOfGroup_PaginatesWithoutSharedProjects(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/groups/5/projects":   []obj{{"id": 1, "path_with_namespace": "g/p1"}},
		"/api/v4/groups/5/projects?2": []obj{{"id": 2, "path_with_namespace": "g/sub/p2"}},
	}, func(r *http.Request) {
		assert.Equal(t, "true", r.URL.Query().Get("include_subgroups"))
		assert.Equal(t, "false", r.URL.Query().Get("with_shared"))
	})

	projects, err := newTestApp(t, srv, 2).GetRecursiveProjectsOfGroup(context.Background(), 5)

	require.NoError(t, err)
	require.Len(t, projects, 2)
	assert.Equal(t, "g/sub/p2", projects[1].PathWithNamespace)
}

func TestApp_GetSubGroups_Paginates(t *testing.T) {
	srv := fakeGitLab(t, map[string]any{
		"/api/v4/groups/5/descendant_groups":   []obj{{"id": 6, "full_path": "g/a"}},
		"/api/v4/groups/5/descendant_groups?2": []obj{{"id": 7, "full_path": "g/a/b"}},
	}, nil)

	groups, err := newTestApp(t, srv, 2).GetSubGroups(context.Background(), 5)

	require.NoError(t, err)
	require.Len(t, groups, 2)
	assert.Equal(t, "g/a/b", groups[1].FullPath)
}
