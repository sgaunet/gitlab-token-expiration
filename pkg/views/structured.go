package views

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"go.yaml.in/yaml/v3"
)

// Output formats supported by NewRenderer.
const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatYAML  = "yaml"
)

// Formats lists the supported output formats, the default one first.
var Formats = []string{FormatTable, FormatJSON, FormatYAML}

// StructuredOutput renders the tokens as a JSON or YAML list, for scripts.
// Colors, header and expiration highlighting do not apply.
type StructuredOutput struct {
	format       string
	w            io.Writer
	printRevoked bool
}

// NewStructuredOutput returns a renderer writing the tokens to w in the given
// format (FormatJSON or FormatYAML). Revoked tokens are skipped unless
// printRevoked is set, like in the table output.
func NewStructuredOutput(format string, w io.Writer, printRevoked bool) (*StructuredOutput, error) {
	if format != FormatJSON && format != FormatYAML {
		return nil, fmt.Errorf("%w: %q", ErrUnknownFormat, format)
	}
	if w == nil {
		w = os.Stdout
	}
	return &StructuredOutput{format: format, w: w, printRevoked: printRevoked}, nil
}

// Render writes the tokens. An empty result is written as an empty list.
func (s *StructuredOutput) Render(tokens []dto.Token) error {
	kept := make([]dto.Token, 0, len(tokens))
	for _, t := range tokens {
		if !s.printRevoked && t.Revoked {
			continue
		}
		kept = append(kept, t)
	}

	if s.format == FormatJSON {
		enc := json.NewEncoder(s.w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(kept); err != nil {
			return fmt.Errorf("error rendering JSON: %w", err)
		}
		return nil
	}

	enc := yaml.NewEncoder(s.w)
	enc.SetIndent(2) //nolint:mnd // conventional YAML indentation
	if err := enc.Encode(kept); err != nil {
		return fmt.Errorf("error rendering YAML: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("error rendering YAML: %w", err)
	}
	return nil
}
