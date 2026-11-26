// Package filter selects tokens by name and source with regular expressions,
// and hides revoked tokens.
package filter

import (
	"fmt"
	"regexp"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
)

// CompilePattern compiles a filter pattern (Go RE2 syntax, matched anywhere in
// the value). With ignoreCase, the match is case-insensitive.
func CompilePattern(pattern string, ignoreCase bool) (*regexp.Regexp, error) {
	// Compile the pattern as given first, so that an error shows exactly what
	// the user typed, without the (?i) prefix added for ignoreCase.
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regular expression: %w", err)
	}
	if ignoreCase {
		re = regexp.MustCompile("(?i)" + pattern)
	}
	return re, nil
}

// TokenFilter keeps the tokens that match every Include pattern that is set
// and none of the Exclude patterns that are set. A nil pattern is not set.
// With HideRevoked, revoked tokens are dropped too. The zero value keeps
// every token.
type TokenFilter struct {
	IncludeName   *regexp.Regexp
	ExcludeName   *regexp.Regexp
	IncludeSource *regexp.Regexp
	ExcludeSource *regexp.Regexp
	HideRevoked   bool
}

// Apply returns the tokens kept by the filter, in their original order.
func (f TokenFilter) Apply(tokens []dto.Token) []dto.Token {
	if f == (TokenFilter{}) {
		return tokens
	}
	kept := make([]dto.Token, 0, len(tokens))
	for _, t := range tokens {
		if f.keep(t) {
			kept = append(kept, t)
		}
	}
	return kept
}

func (f TokenFilter) keep(t dto.Token) bool {
	if f.HideRevoked && t.Revoked {
		return false
	}
	return includes(f.IncludeName, t.Name) && !excludes(f.ExcludeName, t.Name) &&
		includes(f.IncludeSource, t.Source) && !excludes(f.ExcludeSource, t.Source)
}

// includes reports whether value matches re, or true when re is not set.
func includes(re *regexp.Regexp, value string) bool {
	return re == nil || re.MatchString(value)
}

// excludes reports whether value matches re, or false when re is not set.
func excludes(re *regexp.Regexp, value string) bool {
	return re != nil && re.MatchString(value)
}
