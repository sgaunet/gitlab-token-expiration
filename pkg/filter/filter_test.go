package filter_test

import (
	"regexp"
	"testing"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tokenIDs returns the IDs of tokens, to compare results concisely.
func tokenIDs(tokens []dto.Token) []int64 {
	res := make([]int64, 0, len(tokens))
	for _, t := range tokens {
		res = append(res, t.ID)
	}
	return res
}

func TestTokenFilter_Apply(t *testing.T) {
	tokens := []dto.Token{
		{ID: 1, Name: "renovate", Source: "team-a/app"},
		{ID: 2, Name: "deploy-ci", Source: "team-a/infra"},
		{ID: 3, Name: "Renovate", Source: "team-b/app"},
		{ID: 4, Name: "tmp-debug", Source: "team-b/tools"},
	}

	type patterns struct {
		includeName, excludeName, includeSource, excludeSource string
	}
	tests := []struct {
		name       string
		patterns   patterns
		ignoreCase bool
		want       []int64
	}{
		{
			name: "no pattern keeps everything",
			want: []int64{1, 2, 3, 4},
		},
		{
			name:     "include name matches anywhere",
			patterns: patterns{includeName: "ploy"},
			want:     []int64{2},
		},
		{
			name:     "exclude name",
			patterns: patterns{excludeName: "renovate"},
			want:     []int64{2, 3, 4},
		},
		{
			name:     "include source with anchor",
			patterns: patterns{includeSource: "^team-a/"},
			want:     []int64{1, 2},
		},
		{
			name:     "exclude source",
			patterns: patterns{excludeSource: "/app$"},
			want:     []int64{2, 4},
		},
		{
			name:     "include source and exclude name",
			patterns: patterns{includeSource: "^team-a/", excludeName: "renovate"},
			want:     []int64{2},
		},
		{
			name:     "include source and include name must both match",
			patterns: patterns{includeSource: "/app$", includeName: "^Renovate$"},
			want:     []int64{3},
		},
		{
			name:       "ignore case on include",
			patterns:   patterns{includeName: "^renovate$"},
			ignoreCase: true,
			want:       []int64{1, 3},
		},
		{
			name:       "ignore case on exclude",
			patterns:   patterns{excludeName: "RENOVATE"},
			ignoreCase: true,
			want:       []int64{2, 4},
		},
		{
			name:       "alternation with ignore case",
			patterns:   patterns{includeName: "^RENOVATE$|^tmp-"},
			ignoreCase: true,
			want:       []int64{1, 3, 4},
		},
		{
			name:     "no match",
			patterns: patterns{includeName: "nothing"},
			want:     []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := filter.TokenFilter{
				IncludeName:   mustCompile(t, tt.patterns.includeName, tt.ignoreCase),
				ExcludeName:   mustCompile(t, tt.patterns.excludeName, tt.ignoreCase),
				IncludeSource: mustCompile(t, tt.patterns.includeSource, tt.ignoreCase),
				ExcludeSource: mustCompile(t, tt.patterns.excludeSource, tt.ignoreCase),
			}
			assert.Equal(t, tt.want, tokenIDs(f.Apply(tokens)))
		})
	}
}

func TestTokenFilter_Apply_EmptyInput(t *testing.T) {
	f := filter.TokenFilter{IncludeName: regexp.MustCompile("x")}
	assert.Empty(t, f.Apply(nil))
	assert.Empty(t, f.Apply([]dto.Token{}))
	assert.Nil(t, filter.TokenFilter{}.Apply(nil))
}

func TestCompilePattern_InvalidShowsPatternAsTyped(t *testing.T) {
	for _, ignoreCase := range []bool{false, true} {
		_, err := filter.CompilePattern("a(", ignoreCase)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regular expression")
		assert.Contains(t, err.Error(), "`a(`")
		assert.NotContains(t, err.Error(), "(?i)")
	}
}

// mustCompile compiles pattern, or returns nil (pattern not set) when it is empty.
func mustCompile(t *testing.T, pattern string, ignoreCase bool) *regexp.Regexp {
	t.Helper()
	if pattern == "" {
		return nil
	}
	re, err := filter.CompilePattern(pattern, ignoreCase)
	require.NoError(t, err)
	return re
}
