// Package dto contains data transfer objects for the GitLab token expiration tool.
package dto

// Token represents a Gitlab token (pat, deploy_token, access_token)
// some fields are omitted.
type Token struct {
	Source    string `json:"source"` // project or group or personal
	Type      string `json:"type"`   // pat or deploy_token or access_token
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Revoked   bool   `json:"revoked"`
	ExpiresAt string `json:"expires_at"`
}
