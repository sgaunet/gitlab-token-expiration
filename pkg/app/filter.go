package app

import (
	"fmt"
	"regexp"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
)

// NameFilter selects tokens by name with a regular expression, like grep.
// The zero value (or a nil *NameFilter) keeps every token.
type NameFilter struct {
	re     *regexp.Regexp
	invert bool
}

// NewNameFilter compiles pattern (Go RE2 syntax, matched anywhere in the token
// name) into a NameFilter. With invert, the tokens that do NOT match are kept.
// With ignoreCase, the match is case-insensitive.
// An empty pattern returns a nil filter, which keeps every token.
func NewNameFilter(pattern string, invert, ignoreCase bool) (*NameFilter, error) {
	if pattern == "" {
		return nil, nil //nolint:nilnil // a nil filter is a valid "no filter"
	}
	if ignoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid name filter %q: %w", pattern, err)
	}
	return &NameFilter{re: re, invert: invert}, nil
}

// Apply returns the tokens kept by the filter, in their original order.
func (f *NameFilter) Apply(tokens []dto.Token) []dto.Token {
	if f == nil || f.re == nil {
		return tokens
	}
	kept := make([]dto.Token, 0, len(tokens))
	for _, t := range tokens {
		if f.re.MatchString(t.Name) != f.invert {
			kept = append(kept, t)
		}
	}
	return kept
}
