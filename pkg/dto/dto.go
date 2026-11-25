// Package dto contains data transfer objects for the GitLab token expiration tool.
package dto

// Token types, the values of Token.Type.
const (
	TypeGroupAccessToken    = "group_access_token"
	TypeGroupDeployToken    = "group_deploy_token"
	TypeProjectAccessToken  = "project_access_token"
	TypeProjectDeployToken  = "project_deploy_token"
	TypePersonalAccessToken = "personal_access_token"
)

// Token represents a Gitlab token (access token, deploy token or personal access token).
// Some fields of the GitLab API are omitted.
//
// The JSON and YAML output (-o json|yaml) does not marshal Token directly:
// its format is defined in pkg/views/structured.go.
type Token struct {
	Source    string `json:"source"` // group or project path, empty for personal access tokens
	Type      string `json:"type"`   // one of the Type* constants
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Revoked   bool   `json:"revoked"`
	ExpiresAt string `json:"expires_at"` // YYYY-MM-DD, empty when the token never expires
}
