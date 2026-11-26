package cmd_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pterm/pterm"
	"github.com/sgaunet/gitlab-token-expiration/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests are not parallel: they run the root command, whose flags are
// package globals, and they redirect os.Stdout and os.Stderr.

// run executes the command line args, with stdout and stderr discarded.
func run(t *testing.T, args ...string) error {
	t.Helper()
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err)
	stdout, stderr, ptermOut := os.Stdout, os.Stderr, pterm.GetDefaultOutput()
	os.Stdout, os.Stderr = devNull, devNull
	pterm.SetDefaultOutput(io.Discard)
	t.Cleanup(func() {
		os.Stdout, os.Stderr = stdout, stderr
		pterm.SetDefaultOutput(ptermOut)
		_ = devNull.Close()
		cmd.ResetFlags()
	})

	cmd.ResetFlags()
	cmd.RootCmd.SetArgs(args)
	cmd.RootCmd.SetOut(io.Discard)
	cmd.RootCmd.SetErr(io.Discard)
	return cmd.RootCmd.ExecuteContext(context.Background())
}

// fakeGitLab serves the given JSON bodies, or HTTP status codes (int), by path.
// Any other path fails the test.
func fakeGitLab(t *testing.T, routes map[string]any) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := routes[r.URL.Path]
		if !ok {
			t.Errorf("unexpected GitLab API call: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if status, isStatus := body.(int); isStatus {
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`{"message":"denied"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITLAB_URI", srv.URL)
	t.Setenv("GITLAB_TOKEN", "test-token")
}

type obj = map[string]any

func TestInvalidUsageFailsBeforeAnyAPICall(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		token   string
		wantErr string
	}{
		{"no-recursive without id", []string{"group", "-n"}, "x", "--no-recursive requires --id"},
		{"parallel zero", []string{"project", "-p", "0"}, "x", "--parallel must be at least 1, got 0"},
		{"negative parallel", []string{"group", "--parallel=-3"}, "x", "--parallel must be at least 1, got -3"},
		{"invalid pattern", []string{"pat", "--filter-name", "a("}, "x", "--filter-name: invalid regular expression"},
		{"ignore-case without pattern", []string{"pat", "--ignore-case"}, "x", "--ignore-case requires"},
		{"unknown output format", []string{"group", "-o", "xml"}, "x", `unknown output format: "xml"`},
		{"missing GITLAB_TOKEN", []string{"pat"}, "", "GITLAB_TOKEN is not set"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeGitLab(t, nil) // any API call fails the test
			t.Setenv("GITLAB_TOKEN", tt.token)

			err := run(t, tt.args...)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestInvalidGitlabURI(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "x")
	t.Setenv("GITLAB_URI", "http://[::1")

	err := run(t, "pat")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GITLAB_URI")
}

// A group Maintainer can read the group deploy tokens but not the group
// access tokens: the readable tokens are shown and the command succeeds,
// even for a single --id.
func TestGroup_PartlyDeniedSucceeds(t *testing.T) {
	fakeGitLab(t, map[string]any{
		"/api/v4/groups/1":               obj{"id": 1, "full_path": "team-a"},
		"/api/v4/groups/1/access_tokens": http.StatusForbidden,
		"/api/v4/groups/1/deploy_tokens": []obj{{"id": 12, "name": "deploy"}},
	})

	require.NoError(t, run(t, "group", "-i", "1", "-n", "-o", "json"))
}

// When GitLab denies access to every token list (e.g. a Developer of the
// project, or an expired GITLAB_TOKEN), the command fails instead of printing
// an empty result.
func TestProject_AllDeniedFails(t *testing.T) {
	fakeGitLab(t, map[string]any{
		"/api/v4/projects/7":               obj{"id": 7, "path_with_namespace": "team-a/app"},
		"/api/v4/projects/7/access_tokens": http.StatusForbidden,
		"/api/v4/projects/7/deploy_tokens": http.StatusForbidden,
	})

	err := run(t, "project", "-i", "7")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token could be read")
}

// No resource at all is not an error: there is simply no token to show.
func TestGroup_NoGroupSucceeds(t *testing.T) {
	fakeGitLab(t, map[string]any{
		"/api/v4/groups": []obj{},
	})

	require.NoError(t, run(t, "group", "-o", "yaml"))
}

func TestPat_ListsCurrentUserTokens(t *testing.T) {
	fakeGitLab(t, map[string]any{
		"/api/v4/user":                   obj{"id": 7},
		"/api/v4/personal_access_tokens": []obj{{"id": 1, "name": "mine"}},
	})

	require.NoError(t, run(t, "pat", "-r"))
}
