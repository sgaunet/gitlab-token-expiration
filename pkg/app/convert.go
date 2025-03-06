package app

import (
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
)

func ConvertGroupAccessTokenToDTOToken(groupAccessToken gitlab.GroupAccessToken) dto.Token {
	return dto.Token{
		ID:        groupAccessToken.Id,
		Name:      groupAccessToken.Name,
		ExpiresAt: groupAccessToken.ExpiresAt,
		Revoked:   groupAccessToken.Revoked,
		Source:    "group",
		Type:      "access_token",
	}
}

func ConvertGroupDeployTokenToDTOToken(groupDeployToken gitlab.GroupDeployToken) dto.Token {
	return dto.Token{
		ID:        groupDeployToken.Id,
		Name:      groupDeployToken.Name,
		ExpiresAt: groupDeployToken.ExpiresAt,
		Revoked:   groupDeployToken.Revoked,
		Source:    "group",
		Type:      "deploy_token",
	}
}

func ConvertProjectAccessTokenToDTOToken(projectAccessToken gitlab.ProjectAccessToken) dto.Token {
	return dto.Token{
		ID:        projectAccessToken.Id,
		Name:      projectAccessToken.Name,
		ExpiresAt: projectAccessToken.ExpiresAt,
		Revoked:   projectAccessToken.Revoked,
		Source:    "project",
		Type:      "access_token",
	}
}

func ConvertProjectDeployTokenToDTOToken(projectDeployToken gitlab.PersonalAccessToken) dto.Token {
	return dto.Token{
		ID:        projectDeployToken.Id,
		Name:      projectDeployToken.Name,
		ExpiresAt: projectDeployToken.ExpiresAt,
		Revoked:   projectDeployToken.Revoked,
		Source:    "project",
		Type:      "deploy_token",
	}
}

func ConvertPersonalGitlabTokenToDTOToken(personalGitlabToken gitlab.PersonalAccessToken) dto.Token {
	return dto.Token{
		ID:        personalGitlabToken.Id,
		Name:      personalGitlabToken.Name,
		ExpiresAt: personalGitlabToken.ExpiresAt,
		Revoked:   personalGitlabToken.Revoked,
		Source:    "",
		Type:      "personal_access_token",
	}
}

func ConvertGroupAccessTokenToDTOTokens(groupAccessTokens []gitlab.GroupAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, groupAccessToken := range groupAccessTokens {
		tokens = append(tokens, ConvertGroupAccessTokenToDTOToken(groupAccessToken))
	}
	return tokens
}

func ConvertGroupDeployTokenToDTOTokens(groupDeployTokens []gitlab.GroupDeployToken) []dto.Token {
	var tokens []dto.Token
	for _, groupDeployToken := range groupDeployTokens {
		tokens = append(tokens, ConvertGroupDeployTokenToDTOToken(groupDeployToken))
	}
	return tokens
}

func ConvertProjectAccessTokenToDTOTokens(projectAccessTokens []gitlab.ProjectAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, projectAccessToken := range projectAccessTokens {
		tokens = append(tokens, ConvertProjectAccessTokenToDTOToken(projectAccessToken))
	}
	return tokens
}

func ConvertProjectDeployTokenToDTOTokens(projectDeployTokens []gitlab.PersonalAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, projectDeployToken := range projectDeployTokens {
		tokens = append(tokens, ConvertProjectDeployTokenToDTOToken(projectDeployToken))
	}
	return tokens
}

func convertPersonalGitlabTokenToDTOTokens(personalGitlabTokens []gitlab.PersonalAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, personalGitlabToken := range personalGitlabTokens {
		tokens = append(tokens, ConvertPersonalGitlabTokenToDTOToken(personalGitlabToken))
	}
	return tokens
}
