package app

import (
	"cmp"
	"net/http"
	"slices"

	"gitlab.com/gitlab-org/api/client-go"
)

// Kinds of token lists, the values of SkippedResource.Kind.
const (
	KindGroupAccessTokens   = "group access tokens"
	KindGroupDeployTokens   = "group deploy tokens"
	KindProjectAccessTokens = "project access tokens"
	KindProjectDeployTokens = "project deploy tokens"
)

// SkippedResource is a list of tokens of a group or a project that was not
// read because GitLab denied access to it (401 Unauthorized or 403 Forbidden),
// for example the access tokens of a group where the user is only Maintainer.
type SkippedResource struct {
	Kind string // one of the Kind* constants
	Path string // path of the group or project
	Err  error
}

// Skipped returns the token lists skipped because access was denied,
// sorted by path then kind.
func (a *App) Skipped() []SkippedResource {
	a.skippedMu.Lock()
	defer a.skippedMu.Unlock()

	res := slices.Clone(a.skipped)
	slices.SortFunc(res, func(x, y SkippedResource) int {
		return cmp.Or(cmp.Compare(x.Path, y.Path), cmp.Compare(x.Kind, y.Kind))
	})
	return res
}

// Listed returns the number of token lists of groups and projects read
// successfully. With Skipped, it tells a partial result (some lists denied)
// from a run where nothing could be read.
func (a *App) Listed() int {
	return int(a.listed.Load())
}

// skipIfDenied records the token list as skipped and returns true when err is
// a 401 or 403 response. It is safe for concurrent use.
func (a *App) skipIfDenied(kind, path string, err error) bool {
	if !isAccessDenied(err) {
		return false
	}
	a.skippedMu.Lock()
	defer a.skippedMu.Unlock()
	a.skipped = append(a.skipped, SkippedResource{Kind: kind, Path: path, Err: err})
	return true
}

func isAccessDenied(err error) bool {
	return gitlab.HasStatusCode(err, http.StatusUnauthorized) ||
		gitlab.HasStatusCode(err, http.StatusForbidden)
}
