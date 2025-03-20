package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
)

// PersonalAccessToken represents a Gitlab personal access token
type PersonalAccessToken struct {
	Id         int      `json:"id"`
	Name       string   `json:"name"`
	Revoked    bool     `json:"revoked"`
	CreatedAt  string   `json:"created_at"`
	Scopes     []string `json:"scopes"`
	UserId     int      `json:"user_id"`
	LastUsedAt string   `json:"last_used_at"`
	Active     bool     `json:"active"`
	ExpiresAt  string   `json:"expires_at"`
}

// GetPersonalAccessTokens returns the list of personal access tokens
func (s *GitlabService) GetPersonalAccessTokens() (res []PersonalAccessToken, err error) {
	url := fmt.Sprintf("%s/personal_access_tokens", s.gitlabApiEndpoint)
	resp, err := s.get(url)
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	// Unmarshal the response
	if err = json.Unmarshal(body, &res); err != nil {
		// If the response is an error message, unmarshal it
		return res, UnmarshalErrorMessage(body)
	}
	return res, err
}
