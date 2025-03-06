package app_test

import (
	"testing"
	"time"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
	"github.com/stretchr/testify/assert"
)

func TestConvertGroupAccessTokenToDTOToken(t *testing.T) {
	groupAccessToken := gitlab.GroupAccessToken{
		Id:        1,
		Name:      "test-token",
		ExpiresAt: time.Now().Format(time.RFC3339),
		Revoked:   false,
	}

	expectedToken := dto.Token{
		ID:        1,
		Name:      "test-token",
		ExpiresAt: groupAccessToken.ExpiresAt,
		Revoked:   false,
		Source:    "group",
		Type:      "access_token",
	}

	result := app.ConvertGroupAccessTokenToDTOToken(groupAccessToken)
	assert.Equal(t, expectedToken, result)
}
func TestConvertGroupDeployTokenToDTOToken(t *testing.T) {
	groupDeployToken := gitlab.GroupDeployToken{
		Id:        1,
		Name:      "test-deploy-token",
		ExpiresAt: time.Now().Format(time.RFC3339),
		Revoked:   false,
	}

	expectedToken := dto.Token{
		ID:        1,
		Name:      "test-deploy-token",
		ExpiresAt: groupDeployToken.ExpiresAt,
		Revoked:   false,
		Source:    "group",
		Type:      "deploy_token",
	}

	result := app.ConvertGroupDeployTokenToDTOToken(groupDeployToken)
	assert.Equal(t, expectedToken, result)
}
func TestConvertProjectAccessTokenToDTOToken(t *testing.T) {
	projectAccessToken := gitlab.ProjectAccessToken{
		Id:        1,
		Name:      "test-project-token",
		ExpiresAt: time.Now().Format(time.RFC3339),
		Revoked:   false,
	}

	expectedToken := dto.Token{
		ID:        1,
		Name:      "test-project-token",
		ExpiresAt: projectAccessToken.ExpiresAt,
		Revoked:   false,
		Source:    "project",
		Type:      "access_token",
	}

	result := app.ConvertProjectAccessTokenToDTOToken(projectAccessToken)
	assert.Equal(t, expectedToken, result)
}
func TestConvertProjectDeployTokenToDTOToken(t *testing.T) {
	projectDeployToken := gitlab.PersonalAccessToken{
		Id:        1,
		Name:      "test-project-deploy-token",
		ExpiresAt: time.Now().Format(time.RFC3339),
		Revoked:   false,
	}

	expectedToken := dto.Token{
		ID:        1,
		Name:      "test-project-deploy-token",
		ExpiresAt: projectDeployToken.ExpiresAt,
		Revoked:   false,
		Source:    "project",
		Type:      "deploy_token",
	}

	result := app.ConvertProjectDeployTokenToDTOToken(projectDeployToken)
	assert.Equal(t, expectedToken, result)
}
func TestConvertPersonalGitlabTokenToDTOToken(t *testing.T) {
	personalGitlabToken := gitlab.PersonalAccessToken{
		Id:        1,
		Name:      "test-personal-token",
		ExpiresAt: time.Now().Format(time.RFC3339),
		Revoked:   false,
	}

	expectedToken := dto.Token{
		ID:        1,
		Name:      "test-personal-token",
		ExpiresAt: personalGitlabToken.ExpiresAt,
		Revoked:   false,
		Source:    "",
		Type:      "personal_access_token",
	}

	result := app.ConvertPersonalGitlabTokenToDTOToken(personalGitlabToken)
	assert.Equal(t, expectedToken, result)
}
