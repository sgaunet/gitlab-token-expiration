package gitlab

import (
	"encoding/json"
	"fmt"
)

// GitlabProject represents a Gitlab project
// https://docs.gitlab.com/ee/api/projects.html
// struct fields are not exhaustive - most of them won't be used
type GitlabProject struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	DescriptionHtml string `json:"description_html"`
	DefaultBranch   string `json:"default_branch"`
	Visibility      string `json:"visibility"`
	SshUrlToRepo    string `json:"ssh_url_to_repo"`
	HttpUrlToRepo   string `json:"http_url_to_repo"`
	WebUrl          string `json:"web_url"`
	Archived        bool   `json:"archived"`
}

// GetProject returns the gitlab project from the given ID
type ProjectAccessToken struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	Revoked     bool     `json:"revoked"`
	CreatedAt   string   `json:"created_at"`
	Scopes      []string `json:"scopes"`
	UserId      int      `json:"user_id"`
	LastUsedAt  string   `json:"last_used_at"`
	Active      bool     `json:"active"`
	ExpiresAt   string   `json:"expires_at"`
	AccessLevel int      `json:"access_level"`
}

// GetProjectAccessTokens returns the list of access tokens of the project
func (s *GitlabService) GetProjectAccessTokens(projectID int) ([]ProjectAccessToken, error) {
	url := fmt.Sprintf("%s/projects/%d/access_tokens", s.gitlabApiEndpoint, projectID)
	resp, err := s.get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res []ProjectAccessToken
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
