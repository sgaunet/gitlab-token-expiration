package views

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"go.yaml.in/yaml/v3"
)

// Output formats accepted by --output.
const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatYAML  = "yaml"
)

// day is the length of a calendar day in UTC.
const day = 24 * time.Hour

// Formats returns the supported output formats, the default one first.
func Formats() []string {
	return []string{FormatTable, FormatJSON, FormatYAML}
}

// structuredToken is the format of a token in the JSON and YAML output.
// Scripts depend on it: fields may be added, but never renamed or removed.
type structuredToken struct {
	Source  string `json:"source"  yaml:"source"`
	Type    string `json:"type"    yaml:"type"`
	ID      int64  `json:"id"      yaml:"id"`
	Name    string `json:"name"    yaml:"name"`
	Revoked bool   `json:"revoked" yaml:"revoked"`
	// ExpiresAt is the expiration date (YYYY-MM-DD), nil when the token never expires.
	ExpiresAt *string `json:"expires_at" yaml:"expires_at"`
	// ExpiresInDays is the number of days from today (UTC) to ExpiresAt, nil
	// when the token never expires. GitLab tokens stop working at 00:00 UTC on
	// their expiration date, so 0 or less means the token has expired.
	ExpiresInDays *int `json:"expires_in_days" yaml:"expires_in_days"`
}

// StructuredOutput renders the tokens as a JSON or YAML list, for scripts.
// Colors, header and expiration highlighting do not apply.
type StructuredOutput struct {
	format string
	w      io.Writer
	now    func() time.Time
}

// NewStructuredOutput returns a renderer writing the tokens to w in the given
// format (FormatJSON or FormatYAML).
func NewStructuredOutput(format string, w io.Writer) (*StructuredOutput, error) {
	if format != FormatJSON && format != FormatYAML {
		return nil, fmt.Errorf("%w: %q", ErrUnknownFormat, format)
	}
	if w == nil {
		w = os.Stdout
	}
	return &StructuredOutput{format: format, w: w, now: time.Now}, nil
}

// Render writes every given token. An empty result is written as an empty list.
func (s *StructuredOutput) Render(tokens []dto.Token) error {
	out := make([]structuredToken, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, s.toStructured(t))
	}

	if s.format == FormatJSON {
		enc := json.NewEncoder(s.w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(out); err != nil {
			return fmt.Errorf("error rendering JSON: %w", err)
		}
		return nil
	}

	enc := yaml.NewEncoder(s.w)
	enc.SetIndent(2) //nolint:mnd // conventional YAML indentation
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("error rendering YAML: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("error rendering YAML: %w", err)
	}
	return nil
}

func (s *StructuredOutput) toStructured(t dto.Token) structuredToken {
	st := structuredToken{
		Source:  t.Source,
		Type:    t.Type,
		ID:      t.ID,
		Name:    t.Name,
		Revoked: t.Revoked,
	}
	if t.ExpiresAt != "" {
		st.ExpiresAt = new(t.ExpiresAt)
		st.ExpiresInDays = daysUntil(t.ExpiresAt, s.now())
	}
	return st
}

// daysUntil returns the number of days from the day of now (UTC) to the date
// expiresAt (YYYY-MM-DD), or nil when expiresAt is not a valid date.
func daysUntil(expiresAt string, now time.Time) *int {
	date, err := time.Parse(time.DateOnly, expiresAt)
	if err != nil {
		return nil
	}
	today := now.UTC().Truncate(day)
	return new(int(date.Sub(today) / day))
}
