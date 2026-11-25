package views_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

var structuredTokens = []dto.Token{
	{Source: "a/b", Type: "access_token", ID: 1, Name: "gat", ExpiresAt: "2027-01-01"},
	{Source: "a/p", Type: "access_token", ID: 2, Name: "old", Revoked: true, ExpiresAt: "2025-01-01"},
}

func TestStructuredOutput_JSON(t *testing.T) {
	var buf bytes.Buffer
	r, err := views.NewStructuredOutput(views.FormatJSON, &buf, false)
	require.NoError(t, err)

	require.NoError(t, r.Render(structuredTokens))

	assert.JSONEq(t, `[{"source":"a/b","type":"access_token","id":1,"name":"gat","revoked":false,"expires_at":"2027-01-01"}]`,
		buf.String())
}

func TestStructuredOutput_YAMLWithRevoked(t *testing.T) {
	var buf bytes.Buffer
	r, err := views.NewStructuredOutput(views.FormatYAML, &buf, true)
	require.NoError(t, err)

	require.NoError(t, r.Render(structuredTokens))

	var got []dto.Token
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, structuredTokens, got)
	assert.Contains(t, buf.String(), "expires_at: \"2027-01-01\"")
}

func TestStructuredOutput_EmptyIsAnEmptyList(t *testing.T) {
	var buf bytes.Buffer
	r, err := views.NewStructuredOutput(views.FormatJSON, &buf, false)
	require.NoError(t, err)

	require.NoError(t, r.Render(nil))

	var got []dto.Token
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.NotNil(t, got)
	assert.Empty(t, got)
	assert.Equal(t, "[]\n", buf.String())

	buf.Reset()
	r, err = views.NewStructuredOutput(views.FormatYAML, &buf, false)
	require.NoError(t, err)
	require.NoError(t, r.Render(nil))
	assert.Equal(t, "[]\n", buf.String())
}

func TestNewStructuredOutput_UnknownFormat(t *testing.T) {
	_, err := views.NewStructuredOutput("xml", nil, false)
	require.ErrorIs(t, err, views.ErrUnknownFormat)
}
