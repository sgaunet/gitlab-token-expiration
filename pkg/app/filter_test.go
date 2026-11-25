package app_test

import (
	"testing"

	"github.com/sgaunet/gitlab-token-expiration/pkg/app"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func names(tokens []dto.Token) []string {
	res := make([]string, 0, len(tokens))
	for _, t := range tokens {
		res = append(res, t.Name)
	}
	return res
}

func TestNameFilter(t *testing.T) {
	tokens := []dto.Token{
		{Name: "hjh/supervision/dev/nicolas-1"},
		{Name: "renovate"},
		{Name: "HJH/supervision/dev/nicolas-2"},
		{Name: "deploy-ci"},
	}

	tests := []struct {
		name       string
		pattern    string
		invert     bool
		ignoreCase bool
		want       []string
	}{
		{
			name: "no pattern keeps everything",
			want: []string{"hjh/supervision/dev/nicolas-1", "renovate", "HJH/supervision/dev/nicolas-2", "deploy-ci"},
		},
		{
			name: "substring match", pattern: "hjh/supervision/dev/nicolas",
			want: []string{"hjh/supervision/dev/nicolas-1"},
		},
		{
			name: "invert", pattern: "hjh/supervision/dev/nicolas", invert: true,
			want: []string{"renovate", "HJH/supervision/dev/nicolas-2", "deploy-ci"},
		},
		{
			name: "ignore case", pattern: "hjh/supervision/dev/nicolas", ignoreCase: true,
			want: []string{"hjh/supervision/dev/nicolas-1", "HJH/supervision/dev/nicolas-2"},
		},
		{
			name: "invert and ignore case", pattern: "^hjh/", invert: true, ignoreCase: true,
			want: []string{"renovate", "deploy-ci"},
		},
		{
			name: "alternation and anchors", pattern: "^renovate$|-ci$",
			want: []string{"renovate", "deploy-ci"},
		},
		{
			name: "no match", pattern: "nothing",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := app.NewNameFilter(tt.pattern, tt.invert, tt.ignoreCase)
			require.NoError(t, err)
			assert.Equal(t, tt.want, names(f.Apply(tokens)))
		})
	}
}

func TestNameFilter_InvalidPattern(t *testing.T) {
	_, err := app.NewNameFilter("nicolas(", false, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid name filter")
}
