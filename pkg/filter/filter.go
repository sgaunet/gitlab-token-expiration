// Package filter selects tokens by name and source with regular expressions.
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
// and none of the Exclude patterns that are set. A nil pattern is not set,
// so the zero value keeps every token.
type TokenFilter struct {
	IncludeName   *regexp.Regexp
	ExcludeName   *regexp.Regexp
	IncludeSource *regexp.Regexp
	ExcludeSource *regexp.Regexp
}

// Apply returns the tokens kept by the filter, in their original order.
func (f TokenFilter) Apply(tokens []dto.Token) []dto.Token {
	if f.IncludeName == nil && f.ExcludeName == nil && f.IncludeSource == nil && f.ExcludeSource == nil {
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
