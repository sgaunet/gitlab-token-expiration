package app

import (
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
)

func convertGroupAccessTokenToDTOToken(groupAccessToken gitlab.GroupAccessToken) dto.Token {
	return dto.Token{
		ID:        groupAccessToken.Id,
		Name:      groupAccessToken.Name,
		ExpiresAt: groupAccessToken.ExpiresAt,
		Revoked:   groupAccessToken.Revoked,
		Source:    "group",
		Type:      "access_token",
	}
}

func convertGroupDeployTokenToDTOToken(groupDeployToken gitlab.GroupDeployToken) dto.Token {
	// fmt.Println(termlink.Link(fmt.Sprintf("# Group Deploy Tokens - %s", group.Path), fmt.Sprintf("%s/-/settings/repository", group.WebUrl)))
	return dto.Token{
		ID:        groupDeployToken.Id,
		Name:      groupDeployToken.Name,
		ExpiresAt: groupDeployToken.ExpiresAt,
		Revoked:   groupDeployToken.Revoked,
		Source:    "group",
		Type:      "deploy_token",
	}
}

func convertProjectAccessTokenToDTOToken(projectAccessToken gitlab.ProjectAccessToken) dto.Token {
	return dto.Token{
		ID:        projectAccessToken.Id,
		Name:      projectAccessToken.Name,
		ExpiresAt: projectAccessToken.ExpiresAt,
		Revoked:   projectAccessToken.Revoked,
		Source:    "project",
		Type:      "access_token",
	}
}

func convertProjectDeployTokenToDTOToken(projectDeployToken gitlab.PersonalAccessToken) dto.Token {
	return dto.Token{
		ID:        projectDeployToken.Id,
		Name:      projectDeployToken.Name,
		ExpiresAt: projectDeployToken.ExpiresAt,
		Revoked:   projectDeployToken.Revoked,
		Source:    "project",
		Type:      "deploy_token",
	}
}

func convertPersonalGitlabTokenToDTOToken(personalGitlabToken gitlab.PersonalAccessToken) dto.Token {
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
		tokens = append(tokens, convertGroupAccessTokenToDTOToken(groupAccessToken))
	}
	return tokens
}

func ConvertGroupDeployTokenToDTOTokens(groupDeployTokens []gitlab.GroupDeployToken) []dto.Token {
	var tokens []dto.Token
	for _, groupDeployToken := range groupDeployTokens {
		tokens = append(tokens, convertGroupDeployTokenToDTOToken(groupDeployToken))
	}
	return tokens
}

func ConvertProjectAccessTokenToDTOTokens(projectAccessTokens []gitlab.ProjectAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, projectAccessToken := range projectAccessTokens {
		tokens = append(tokens, convertProjectAccessTokenToDTOToken(projectAccessToken))
	}
	return tokens
}

func ConvertProjectDeployTokenToDTOTokens(projectDeployTokens []gitlab.PersonalAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, projectDeployToken := range projectDeployTokens {
		tokens = append(tokens, convertProjectDeployTokenToDTOToken(projectDeployToken))
	}
	return tokens
}

func convertPersonalGitlabTokenToDTOTokens(personalGitlabTokens []gitlab.PersonalAccessToken) []dto.Token {
	var tokens []dto.Token
	for _, personalGitlabToken := range personalGitlabTokens {
		tokens = append(tokens, convertPersonalGitlabTokenToDTOToken(personalGitlabToken))
	}
	return tokens
}
