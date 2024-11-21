package views

import (
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
)

type Printer interface {
	PrintGitlabPersonalAccessToken(t []gitlab.PersonalAccessToken) error
	PrintGitlabProjectAccessToken(t []gitlab.ProjectAccessToken) error
	PrintGitlabGroupAccessToken(t []gitlab.GroupAccessToken) error
	PrintGitlabGroupDeployToken(t []gitlab.GroupDeployToken) error
}
