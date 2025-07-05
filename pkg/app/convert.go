package app

import (
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"gitlab.com/gitlab-org/api/client-go"
)

func ConvertGroupAccessTokenToDTOToken(groupAccessToken *gitlab.GroupAccessToken) dto.Token {
	// Convert time format
	var expiresAt string
	if groupAccessToken.ExpiresAt != nil {
		expiresAt = groupAccessToken.ExpiresAt.String()
		if len(expiresAt) >= 10 {
			expiresAt = expiresAt[:10] // Extract YYYY-MM-DD part
		}
	}

	return dto.Token{
		ID:        groupAccessToken.ID,
		Name:      groupAccessToken.Name,
		ExpiresAt: expiresAt,
		Revoked:   groupAccessToken.Revoked,
		Source:    "group",
		Type:      "access_token",
	}
}

func ConvertGroupDeployTokenToDTOToken(groupDeployToken *gitlab.DeployToken) dto.Token {
	// Convert time format
	var expiresAt string
	if groupDeployToken.ExpiresAt != nil {
		expiresAt = groupDeployToken.ExpiresAt.Format("2006-01-02")
	}

	return dto.Token{
		ID:        groupDeployToken.ID,
		Name:      groupDeployToken.Name,
		ExpiresAt: expiresAt,
		Revoked:   groupDeployToken.Revoked,
		Source:    "group",
		Type:      "deploy_token",
	}
}

func ConvertProjectAccessTokenToDTOToken(projectAccessToken *gitlab.ProjectAccessToken) dto.Token {
	// Convert time format
	var expiresAt string
	if projectAccessToken.ExpiresAt != nil {
		expiresAt = projectAccessToken.ExpiresAt.String()
		if len(expiresAt) >= 10 {
			expiresAt = expiresAt[:10] // Extract YYYY-MM-DD part
		}
	}

	return dto.Token{
		ID:        projectAccessToken.ID,
		Name:      projectAccessToken.Name,
		ExpiresAt: expiresAt,
		Revoked:   projectAccessToken.Revoked,
		Source:    "project",
		Type:      "access_token",
	}
}

func ConvertProjectDeployTokenToDTOToken(projectDeployToken *gitlab.DeployToken) dto.Token {
	// Convert time format
	var expiresAt string
	if projectDeployToken.ExpiresAt != nil {
		expiresAt = projectDeployToken.ExpiresAt.Format("2006-01-02")
	}

	return dto.Token{
		ID:        projectDeployToken.ID,
		Name:      projectDeployToken.Name,
		ExpiresAt: expiresAt,
		Revoked:   projectDeployToken.Revoked,
		Source:    "project",
		Type:      "deploy_token",
	}
}

func ConvertPersonalGitlabTokenToDTOToken(personalGitlabToken *gitlab.PersonalAccessToken) dto.Token {
	// Convert time format
	var expiresAt string
	if personalGitlabToken.ExpiresAt != nil {
		expiresAt = personalGitlabToken.ExpiresAt.String()
		if len(expiresAt) >= 10 {
			expiresAt = expiresAt[:10] // Extract YYYY-MM-DD part
		}
	}

	return dto.Token{
		ID:        personalGitlabToken.ID,
		Name:      personalGitlabToken.Name,
		ExpiresAt: expiresAt,
		Revoked:   personalGitlabToken.Revoked,
		Source:    "",
		Type:      "personal_access_token",
	}
}

func ConvertGroupAccessTokenToDTOTokens(groupAccessTokens []*gitlab.GroupAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, groupAccessToken := range groupAccessTokens {
		tokens = append(tokens, ConvertGroupAccessTokenToDTOToken(groupAccessToken))
	}
	return tokens
}

func ConvertGroupDeployTokenToDTOTokens(groupDeployTokens []*gitlab.DeployToken) []dto.Token {
	var tokens []dto.Token
	for _, groupDeployToken := range groupDeployTokens {
		tokens = append(tokens, ConvertGroupDeployTokenToDTOToken(groupDeployToken))
	}
	return tokens
}

func ConvertProjectAccessTokenToDTOTokens(projectAccessTokens []*gitlab.ProjectAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, projectAccessToken := range projectAccessTokens {
		tokens = append(tokens, ConvertProjectAccessTokenToDTOToken(projectAccessToken))
	}
	return tokens
}

func ConvertProjectDeployTokenToDTOTokens(projectDeployTokens []*gitlab.DeployToken) []dto.Token {
	var tokens []dto.Token
	for _, projectDeployToken := range projectDeployTokens {
		tokens = append(tokens, ConvertProjectDeployTokenToDTOToken(projectDeployToken))
	}
	return tokens
}

func convertPersonalGitlabTokenToDTOTokens(personalGitlabTokens []*gitlab.PersonalAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, personalGitlabToken := range personalGitlabTokens {
		tokens = append(tokens, ConvertPersonalGitlabTokenToDTOToken(personalGitlabToken))
	}
	return tokens
}
