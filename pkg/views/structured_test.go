package views_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

// fixedNow is the clock used by the tests: expires_in_days is computed from it.
func fixedNow() time.Time {
	return time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
}

var structuredTokens = []dto.Token{
	{Source: "team-a", Type: dto.TypeGroupAccessToken, ID: 1, Name: "ci&deploy <prod>", ExpiresAt: "2026-12-05"},
	{Source: "team-a/app", Type: dto.TypeProjectAccessToken, ID: 2, Name: "forever"},
	{Source: "team-a", Type: dto.TypeGroupDeployToken, ID: 3, Name: "expired", ExpiresAt: "2026-11-20"},
	{Source: "team-a/app", Type: dto.TypeProjectAccessToken, ID: 4, Name: "today", ExpiresAt: "2026-11-25"},
	{Source: "team-b", Type: dto.TypeGroupAccessToken, ID: 5, Name: "old", Revoked: true, ExpiresAt: "2025-01-01"},
}

// contractToken mirrors the documented JSON/YAML format of a token.
type contractToken struct {
	Source        string  `yaml:"source"`
	Type          string  `yaml:"type"`
	ID            int64   `yaml:"id"`
	Name          string  `yaml:"name"`
	Revoked       bool    `yaml:"revoked"`
	ExpiresAt     *string `yaml:"expires_at"`
	ExpiresInDays *int    `yaml:"expires_in_days"`
}

func render(t *testing.T, format string, now func() time.Time, tokens []dto.Token) string {
	t.Helper()
	var buf bytes.Buffer
	r, err := views.NewStructuredOutput(format, &buf)
	require.NoError(t, err)
	views.SetNow(r, now)
	require.NoError(t, r.Render(tokens))
	return buf.String()
}

func TestStructuredOutput_JSONContract(t *testing.T) {
	out := render(t, views.FormatJSON, fixedNow, structuredTokens[:4])

	assert.JSONEq(t, `[
		{"source": "team-a", "type": "group_access_token", "id": 1, "name": "ci&deploy <prod>",
		 "revoked": false, "expires_at": "2026-12-05", "expires_in_days": 10},
		{"source": "team-a/app", "type": "project_access_token", "id": 2, "name": "forever",
		 "revoked": false, "expires_at": null, "expires_in_days": null},
		{"source": "team-a", "type": "group_deploy_token", "id": 3, "name": "expired",
		 "revoked": false, "expires_at": "2026-11-20", "expires_in_days": -5},
		{"source": "team-a/app", "type": "project_access_token", "id": 4, "name": "today",
		 "revoked": false, "expires_at": "2026-11-25", "expires_in_days": 0}
	]`, out)
	assert.Contains(t, out, `"ci&deploy <prod>"`, "HTML characters must not be escaped")
}

// The renderer writes every token it is given, revoked ones included:
// revoked tokens are hidden by the filter (pkg/filter) unless -r is set.
func TestStructuredOutput_YAMLWithRevoked(t *testing.T) {
	out := render(t, views.FormatYAML, fixedNow, structuredTokens)

	var got []contractToken
	require.NoError(t, yaml.Unmarshal([]byte(out), &got))
	require.Len(t, got, len(structuredTokens))
	assert.Equal(t, "old", got[4].Name)
	assert.True(t, got[4].Revoked)
	assert.Nil(t, got[1].ExpiresAt)
	assert.Nil(t, got[1].ExpiresInDays)
	require.NotNil(t, got[0].ExpiresInDays)
	assert.Equal(t, 10, *got[0].ExpiresInDays)
	assert.Contains(t, out, `expires_at: "2026-12-05"`, "dates must stay strings")
	assert.Contains(t, out, "expires_at: null")
}

func TestStructuredOutput_DaysCountedFromTodayInUTC(t *testing.T) {
	// 23:30 in New York on 2026-11-25 is already 2026-11-26 in UTC.
	lateEvening := func() time.Time {
		return time.Date(2026, 11, 25, 23, 30, 0, 0, time.FixedZone("EST", -5*60*60))
	}
	out := render(t, views.FormatYAML, lateEvening, structuredTokens[:1])

	var got []contractToken
	require.NoError(t, yaml.Unmarshal([]byte(out), &got))
	require.NotNil(t, got[0].ExpiresInDays)
	assert.Equal(t, 9, *got[0].ExpiresInDays)
}

func TestStructuredOutput_InvalidDateHasNoDays(t *testing.T) {
	out := render(t, views.FormatJSON, fixedNow, []dto.Token{{ID: 1, Name: "x", ExpiresAt: "someday"}})

	assert.JSONEq(t, `[{"source": "", "type": "", "id": 1, "name": "x", "revoked": false,
		"expires_at": "someday", "expires_in_days": null}]`, out)
}

func TestStructuredOutput_EmptyIsAnEmptyList(t *testing.T) {
	assert.Equal(t, "[]\n", render(t, views.FormatJSON, fixedNow, nil))
	assert.Equal(t, "[]\n", render(t, views.FormatYAML, fixedNow, nil))
}

func TestNewStructuredOutput_UnknownFormat(t *testing.T) {
	_, err := views.NewStructuredOutput("xml", nil)
	require.ErrorIs(t, err, views.ErrUnknownFormat)
}

func TestFormats(t *testing.T) {
	assert.Equal(t, []string{"table", "json", "yaml"}, views.Formats())
}
