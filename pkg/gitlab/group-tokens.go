package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// GroupAccessToken is a struct that contains the information about a Gitlab group access token
type GroupAccessToken struct {
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

// GetGroupAccessTokens returns the access tokens of the group that matches the given ID
func (s *GitlabService) GetGroupAccessTokens(groupID int) ([]GroupAccessToken, error) {
	url := fmt.Sprintf("%s/groups/%d/access_tokens", s.gitlabApiEndpoint, groupID)
	resp, err := s.get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res []GroupAccessToken
	// Unmarshal the response
	if err = json.Unmarshal(body, &res); err != nil {
		// If the response is an error message, unmarshal it
		return res, UnmarshalErrorMessage(body)
	}
	return res, nil
}

// GroupDeployToken is a struct that contains the information about a Gitlab group deploy token
type GroupDeployToken struct {
	Id        int      `json:"id"`
	Name      string   `json:"name"`
	Username  string   `json:"username"`
	ExpiresAt string   `json:"expires_at"`
	Revoked   bool     `json:"revoked"`
	Expired   bool     `json:"expired"`
	Scopes    []string `json:"scopes"`
}

// GetGroupDeployTokens returns the deploy tokens of the group that matches the given ID
func (s *GitlabService) GetGroupDeployTokens(groupID int) ([]GroupDeployToken, error) {
	url := fmt.Sprintf("%s/groups/%d/deploy_tokens", s.gitlabApiEndpoint, groupID)
	resp, err := s.get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", err)
		}
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res []GroupDeployToken
	// Unmarshal the response
	if err = json.Unmarshal(body, &res); err != nil {
		// If the response is an error message, unmarshal it
		return res, UnmarshalErrorMessage(body)
	}
	return res, nil
}
