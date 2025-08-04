package app

import (
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"gitlab.com/gitlab-org/api/client-go"
)

// ConvertGroupAccessTokenToDTOToken converts a GitLab group access token to a DTO token.
func ConvertGroupAccessTokenToDTOToken(groupAccessToken *gitlab.GroupAccessToken) dto.Token {
	// Convert time format
	var expiresAt string
	if groupAccessToken.ExpiresAt != nil {
		expiresAt = groupAccessToken.ExpiresAt.String()
		const dateFormatLength = 10
		if len(expiresAt) >= dateFormatLength {
			expiresAt = expiresAt[:dateFormatLength] // Extract YYYY-MM-DD part
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

// ConvertGroupDeployTokenToDTOToken converts a GitLab group deploy token to a DTO token.
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

// ConvertProjectAccessTokenToDTOToken converts a GitLab project access token to a DTO token.
func ConvertProjectAccessTokenToDTOToken(projectAccessToken *gitlab.ProjectAccessToken) dto.Token {
	// Convert time format
	var expiresAt string
	if projectAccessToken.ExpiresAt != nil {
		expiresAt = projectAccessToken.ExpiresAt.String()
		const dateFormatLength = 10
		if len(expiresAt) >= dateFormatLength {
			expiresAt = expiresAt[:dateFormatLength] // Extract YYYY-MM-DD part
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

// ConvertProjectDeployTokenToDTOToken converts a GitLab project deploy token to a DTO token.
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

// ConvertPersonalGitlabTokenToDTOToken converts a GitLab personal access token to a DTO token.
func ConvertPersonalGitlabTokenToDTOToken(personalGitlabToken *gitlab.PersonalAccessToken) dto.Token {
	// Convert time format
	var expiresAt string
	if personalGitlabToken.ExpiresAt != nil {
		expiresAt = personalGitlabToken.ExpiresAt.String()
		const dateFormatLength = 10
		if len(expiresAt) >= dateFormatLength {
			expiresAt = expiresAt[:dateFormatLength] // Extract YYYY-MM-DD part
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

// ConvertGroupAccessTokenToDTOTokens converts multiple GitLab group access tokens to DTO tokens.
func ConvertGroupAccessTokenToDTOTokens(groupAccessTokens []*gitlab.GroupAccessToken) []dto.Token {
	tokens := make([]dto.Token, 0, len(groupAccessTokens))
	for _, groupAccessToken := range groupAccessTokens {
		tokens = append(tokens, ConvertGroupAccessTokenToDTOToken(groupAccessToken))
	}
	return tokens
}

// ConvertGroupDeployTokenToDTOTokens converts multiple GitLab group deploy tokens to DTO tokens.
func ConvertGroupDeployTokenToDTOTokens(groupDeployTokens []*gitlab.DeployToken) []dto.Token {
	tokens := make([]dto.Token, 0, len(groupDeployTokens))
	for _, groupDeployToken := range groupDeployTokens {
		tokens = append(tokens, ConvertGroupDeployTokenToDTOToken(groupDeployToken))
	}
	return tokens
}

// ConvertProjectAccessTokenToDTOTokens converts multiple GitLab project access tokens to DTO tokens.
func ConvertProjectAccessTokenToDTOTokens(projectAccessTokens []*gitlab.ProjectAccessToken) []dto.Token {
	tokens := make([]dto.Token, 0, len(projectAccessTokens))
	for _, projectAccessToken := range projectAccessTokens {
		tokens = append(tokens, ConvertProjectAccessTokenToDTOToken(projectAccessToken))
	}
	return tokens
}

// ConvertProjectDeployTokenToDTOTokens converts multiple GitLab project deploy tokens to DTO tokens.
func ConvertProjectDeployTokenToDTOTokens(projectDeployTokens []*gitlab.DeployToken) []dto.Token {
	tokens := make([]dto.Token, 0, len(projectDeployTokens))
	for _, projectDeployToken := range projectDeployTokens {
		tokens = append(tokens, ConvertProjectDeployTokenToDTOToken(projectDeployToken))
	}
	return tokens
}

// ConvertPersonalGitlabTokenToDTOTokens converts multiple GitLab personal access tokens to DTO tokens.
func ConvertPersonalGitlabTokenToDTOTokens(personalGitlabTokens []*gitlab.PersonalAccessToken) []dto.Token {
	tokens := make([]dto.Token, 0, len(personalGitlabTokens))
	for _, personalGitlabToken := range personalGitlabTokens {
		tokens = append(tokens, ConvertPersonalGitlabTokenToDTOToken(personalGitlabToken))
	}
	return tokens
}
