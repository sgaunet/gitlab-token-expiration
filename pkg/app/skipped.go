package app

import (
	"cmp"
	"net/http"
	"slices"

	"gitlab.com/gitlab-org/api/client-go"
)

// SkippedResource is a group or a project whose tokens were not listed
// because GitLab denied access (401 Unauthorized or 403 Forbidden).
type SkippedResource struct {
	Kind string // "group" or "project"
	Path string
	Err  error
}

// WithSkipDenied makes the App skip, instead of failing, the groups and
// projects for which GitLab denies access to the tokens (401 or 403).
// Skipped resources are available through Skipped.
func WithSkipDenied(skip bool) Option {
	return func(a *App) {
		a.skipDenied = skip
	}
}

// Skipped returns the groups and projects skipped because access was denied,
// sorted by kind then path.
func (a *App) Skipped() []SkippedResource {
	a.skippedMu.Lock()
	defer a.skippedMu.Unlock()

	res := slices.Clone(a.skipped)
	slices.SortFunc(res, func(x, y SkippedResource) int {
		return cmp.Or(cmp.Compare(x.Kind, y.Kind), cmp.Compare(x.Path, y.Path))
	})
	return res
}

// skipIfDenied records the resource as skipped and returns true when skipping
// is enabled and err is a 401 or 403 response. It is safe for concurrent use.
func (a *App) skipIfDenied(kind, path string, err error) bool {
	if !a.skipDenied || !isAccessDenied(err) {
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
