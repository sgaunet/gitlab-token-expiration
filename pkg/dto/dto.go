// Package dto contains data transfer objects for the GitLab token expiration tool.
package dto

// Token represents a Gitlab token (pat, deploy_token, access_token)
// some fields are omitted.
type Token struct {
	Source    string `json:"source"     yaml:"source"` // project or group or personal
	Type      string `json:"type"       yaml:"type"`   // pat or deploy_token or access_token
	ID        int64  `json:"id"         yaml:"id"`
	Name      string `json:"name"       yaml:"name"`
	Revoked   bool   `json:"revoked"    yaml:"revoked"`
	ExpiresAt string `json:"expires_at" yaml:"expires_at"`
}
