package app_test

import (
	"testing"
	"time"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/api/client-go"
)

func TestConvertGroupAccessTokenToDTOToken(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		input    *gitlab.GroupAccessToken
		expected string // expected expires_at value
	}{
		{
			name: "with expiry date",
			input: &gitlab.GroupAccessToken{
				PersonalAccessToken: gitlab.PersonalAccessToken{
					ID:        123,
					Name:      "test-token",
					Revoked:   false,
					ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
				},
			},
			expected: "2024-12-31",
		},
		{
			name: "without expiry date",
			input: &gitlab.GroupAccessToken{
				PersonalAccessToken: gitlab.PersonalAccessToken{
					ID:        456,
					Name:      "no-expiry-token",
					Revoked:   true,
					ExpiresAt: nil,
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.ConvertGroupAccessTokenToDTOToken(tt.input)
			
			assert.Equal(t, tt.input.ID, result.ID)
			assert.Equal(t, tt.input.Name, result.Name)
			assert.Equal(t, tt.input.Revoked, result.Revoked)
			assert.Equal(t, tt.expected, result.ExpiresAt)
			assert.Equal(t, "group", result.Source)
			assert.Equal(t, "access_token", result.Type)
		})
	}
}

func TestConvertGroupDeployTokenToDTOToken(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		input    *gitlab.DeployToken
		expected string
	}{
		{
			name: "with expiry date",
			input: &gitlab.DeployToken{
				ID:        123,
				Name:      "deploy-token",
				Revoked:   false,
				ExpiresAt: &expiresAt,
			},
			expected: "2024-12-31",
		},
		{
			name: "without expiry date",
			input: &gitlab.DeployToken{
				ID:        456,
				Name:      "no-expiry-deploy",
				Revoked:   true,
				ExpiresAt: nil,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.ConvertGroupDeployTokenToDTOToken(tt.input)
			
			assert.Equal(t, tt.input.ID, result.ID)
			assert.Equal(t, tt.input.Name, result.Name)
			assert.Equal(t, tt.input.Revoked, result.Revoked)
			assert.Equal(t, tt.expected, result.ExpiresAt)
			assert.Equal(t, "group", result.Source)
			assert.Equal(t, "deploy_token", result.Type)
		})
	}
}

func TestConvertProjectAccessTokenToDTOToken(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		input    *gitlab.ProjectAccessToken
		expected string
	}{
		{
			name: "with expiry date",
			input: &gitlab.ProjectAccessToken{
				PersonalAccessToken: gitlab.PersonalAccessToken{
					ID:        123,
					Name:      "project-token",
					Revoked:   false,
					ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
				},
			},
			expected: "2024-12-31",
		},
		{
			name: "without expiry date",
			input: &gitlab.ProjectAccessToken{
				PersonalAccessToken: gitlab.PersonalAccessToken{
					ID:        456,
					Name:      "no-expiry-project",
					Revoked:   true,
					ExpiresAt: nil,
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.ConvertProjectAccessTokenToDTOToken(tt.input)
			
			assert.Equal(t, tt.input.ID, result.ID)
			assert.Equal(t, tt.input.Name, result.Name)
			assert.Equal(t, tt.input.Revoked, result.Revoked)
			assert.Equal(t, tt.expected, result.ExpiresAt)
			assert.Equal(t, "project", result.Source)
			assert.Equal(t, "access_token", result.Type)
		})
	}
}

func TestConvertProjectDeployTokenToDTOToken(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		input    *gitlab.DeployToken
		expected string
	}{
		{
			name: "with expiry date",
			input: &gitlab.DeployToken{
				ID:        123,
				Name:      "project-deploy",
				Revoked:   false,
				ExpiresAt: &expiresAt,
			},
			expected: "2024-12-31",
		},
		{
			name: "without expiry date",
			input: &gitlab.DeployToken{
				ID:        456,
				Name:      "no-expiry-deploy",
				Revoked:   true,
				ExpiresAt: nil,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.ConvertProjectDeployTokenToDTOToken(tt.input)
			
			assert.Equal(t, tt.input.ID, result.ID)
			assert.Equal(t, tt.input.Name, result.Name)
			assert.Equal(t, tt.input.Revoked, result.Revoked)
			assert.Equal(t, tt.expected, result.ExpiresAt)
			assert.Equal(t, "project", result.Source)
			assert.Equal(t, "deploy_token", result.Type)
		})
	}
}

func TestConvertPersonalGitlabTokenToDTOToken(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		input    *gitlab.PersonalAccessToken
		expected string
	}{
		{
			name: "with expiry date",
			input: &gitlab.PersonalAccessToken{
				ID:        123,
				Name:      "personal-token",
				Revoked:   false,
				ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
			},
			expected: "2024-12-31",
		},
		{
			name: "without expiry date",
			input: &gitlab.PersonalAccessToken{
				ID:        456,
				Name:      "no-expiry-personal",
				Revoked:   true,
				ExpiresAt: nil,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.ConvertPersonalGitlabTokenToDTOToken(tt.input)
			
			assert.Equal(t, tt.input.ID, result.ID)
			assert.Equal(t, tt.input.Name, result.Name)
			assert.Equal(t, tt.input.Revoked, result.Revoked)
			assert.Equal(t, tt.expected, result.ExpiresAt)
			assert.Equal(t, "", result.Source)
			assert.Equal(t, "personal_access_token", result.Type)
		})
	}
}

func TestConvertGroupAccessTokenToDTOTokens(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	input := []*gitlab.GroupAccessToken{
		{
			PersonalAccessToken: gitlab.PersonalAccessToken{
				ID:        1,
				Name:      "token-1",
				Revoked:   false,
				ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
			},
		},
		{
			PersonalAccessToken: gitlab.PersonalAccessToken{
				ID:        2,
				Name:      "token-2",
				Revoked:   true,
				ExpiresAt: nil,
			},
		},
	}

	result := app.ConvertGroupAccessTokenToDTOTokens(input)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "token-1", result[0].Name)
	assert.Equal(t, "2024-12-31", result[0].ExpiresAt)
	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, "token-2", result[1].Name)
	assert.Equal(t, "", result[1].ExpiresAt)
}

func TestConvertGroupDeployTokenToDTOTokens(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	input := []*gitlab.DeployToken{
		{
			ID:        1,
			Name:      "deploy-1",
			Revoked:   false,
			ExpiresAt: &expiresAt,
		},
		{
			ID:        2,
			Name:      "deploy-2",
			Revoked:   true,
			ExpiresAt: nil,
		},
	}

	result := app.ConvertGroupDeployTokenToDTOTokens(input)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "deploy-1", result[0].Name)
	assert.Equal(t, "2024-12-31", result[0].ExpiresAt)
	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, "deploy-2", result[1].Name)
	assert.Equal(t, "", result[1].ExpiresAt)
}

func TestConvertProjectAccessTokenToDTOTokens(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	input := []*gitlab.ProjectAccessToken{
		{
			PersonalAccessToken: gitlab.PersonalAccessToken{
				ID:        1,
				Name:      "project-1",
				Revoked:   false,
				ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
			},
		},
		{
			PersonalAccessToken: gitlab.PersonalAccessToken{
				ID:        2,
				Name:      "project-2",
				Revoked:   true,
				ExpiresAt: nil,
			},
		},
	}

	result := app.ConvertProjectAccessTokenToDTOTokens(input)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "project-1", result[0].Name)
	assert.Equal(t, "2024-12-31", result[0].ExpiresAt)
	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, "project-2", result[1].Name)
	assert.Equal(t, "", result[1].ExpiresAt)
}

func TestConvertProjectDeployTokenToDTOTokens(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	input := []*gitlab.DeployToken{
		{
			ID:        1,
			Name:      "deploy-1",
			Revoked:   false,
			ExpiresAt: &expiresAt,
		},
		{
			ID:        2,
			Name:      "deploy-2",
			Revoked:   true,
			ExpiresAt: nil,
		},
	}

	result := app.ConvertProjectDeployTokenToDTOTokens(input)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "deploy-1", result[0].Name)
	assert.Equal(t, "2024-12-31", result[0].ExpiresAt)
	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, "deploy-2", result[1].Name)
	assert.Equal(t, "", result[1].ExpiresAt)
}

func TestConvertPersonalGitlabTokenToDTOTokens(t *testing.T) {
	expiresAt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	
	input := []*gitlab.PersonalAccessToken{
		{
			ID:        1,
			Name:      "personal-1",
			Revoked:   false,
			ExpiresAt: (*gitlab.ISOTime)(&expiresAt),
		},
		{
			ID:        2,
			Name:      "personal-2",
			Revoked:   true,
			ExpiresAt: nil,
		},
	}

	result := app.ConvertPersonalGitlabTokenToDTOTokens(input)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].ID)
	assert.Equal(t, "personal-1", result[0].Name)
	assert.Equal(t, "2024-12-31", result[0].ExpiresAt)
	assert.Equal(t, "personal_access_token", result[0].Type)
	assert.Equal(t, int64(2), result[1].ID)
	assert.Equal(t, "personal-2", result[1].Name)
	assert.Equal(t, "", result[1].ExpiresAt)
	assert.Equal(t, "personal_access_token", result[1].Type)
}