package app_test

import (
	"net/http"
	"testing"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	t.Setenv("GITLAB_URI", "")

	tests := []struct {
		name string
		opts []app.Option
	}{
		{name: "basic app creation", opts: nil},
		{name: "app with parallel option", opts: []app.Option{app.WithParallel(3)}},
		{
			name: "app with gitlab endpoint option",
			opts: []app.Option{app.WithGitlabEndpoint("https://gitlab.example.com")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			application, err := app.NewApp(tt.opts...)
			require.NoError(t, err)
			assert.NotNil(t, application)
		})
	}
}

func TestNewApp_InvalidGitlabURI(t *testing.T) {
	t.Setenv("GITLAB_URI", "http://[::1")

	application, err := app.NewApp()

	require.Error(t, err)
	assert.Nil(t, application)
	assert.Contains(t, err.Error(), `invalid GITLAB_URI "http://[::1"`)
}

func newApp(t *testing.T) *app.App {
	t.Helper()
	t.Setenv("GITLAB_URI", "")
	application, err := app.NewApp()
	require.NoError(t, err)
	return application
}

func TestApp_SetToken(t *testing.T) {
	application := newApp(t)

	// Test setting a token
	application.SetToken("test-token-123")
	// Since we can't directly test the internal client, we just ensure no panic
	assert.NotNil(t, application)
}

func TestApp_SetGitlabEndpoint(t *testing.T) {
	application := newApp(t)

	// Test setting an endpoint
	application.SetGitlabEndpoint("https://gitlab.example.com")
	// Since we can't directly test the internal client, we just ensure no panic
	assert.NotNil(t, application)
}

func TestApp_SetHTTPClient(t *testing.T) {
	application := newApp(t)

	// Test setting an HTTP client
	httpClient := &http.Client{}
	application.SetHTTPClient(httpClient)
	// Since we can't directly test the internal client, we just ensure no panic
	assert.NotNil(t, application)
}
