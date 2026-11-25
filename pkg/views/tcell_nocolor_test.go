package views_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/pterm/pterm"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableOutput_Render_NoColorHasNoEscapeSequence(t *testing.T) {
	pterm.EnableColor()
	t.Cleanup(pterm.EnableColor)
	var buf bytes.Buffer
	pterm.SetDefaultOutput(&buf)
	t.Cleanup(func() { pterm.SetDefaultOutput(os.Stdout) })

	table := views.NewTableOutput(views.WithColorOption(false), views.WithHeaderOption(true))
	err := table.Render([]dto.Token{
		{ID: 1, Source: "a/b", Type: "access_token", Name: "ci", ExpiresAt: "2020-01-01"},
	})

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "ci")
	assert.NotContains(t, buf.String(), "\x1b[", "--no-color output must not contain ANSI escape sequences")
}
