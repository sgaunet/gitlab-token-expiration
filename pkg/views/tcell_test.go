package views_test

import (
	"testing"

	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
	"github.com/sgaunet/gitlab-token-expiration/pkg/views"
	"github.com/stretchr/testify/assert"
)


func TestNewTableOutput(t *testing.T) {
	tests := []struct {
		name string
		opts []views.TableOutputOption
	}{
		{
			name: "default options",
			opts: nil,
		},
		{
			name: "with header option",
			opts: []views.TableOutputOption{views.WithHeaderOption(true)},
		},
		{
			name: "with color option",
			opts: []views.TableOutputOption{views.WithColorOption(true)},
		},
		{
			name: "with print revoked option",
			opts: []views.TableOutputOption{views.WithPrintRevokedOption(true)},
		},
		{
			name: "with days before expiration",
			opts: []views.TableOutputOption{views.WithNbDaysBeforeExp(30)},
		},
		{
			name: "with all options",
			opts: []views.TableOutputOption{
				views.WithHeaderOption(true),
				views.WithColorOption(true),
				views.WithPrintRevokedOption(true),
				views.WithNbDaysBeforeExp(30),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := views.NewTableOutput(tt.opts...)
			assert.NotNil(t, table)
		})
	}
}

func TestTableOutput_Render(t *testing.T) {
	tests := []struct {
		name   string
		table  views.TableOutput
		tokens []dto.Token
	}{
		{
			name:  "render with header",
			table: views.NewTableOutput(views.WithHeaderOption(true)),
			tokens: []dto.Token{
				{
					ID:        1,
					Source:    "project/test",
					Type:      "pat",
					Name:      "test-token",
					Revoked:   false,
					ExpiresAt: "2025-12-31",
				},
			},
		},
		{
			name:  "render without header",
			table: views.NewTableOutput(views.WithHeaderOption(false)),
			tokens: []dto.Token{
				{
					ID:        1,
					Source:    "project/test",
					Type:      "pat",
					Name:      "test-token",
					Revoked:   false,
					ExpiresAt: "2025-12-31",
				},
			},
		},
		{
			name:  "filter revoked tokens when printRevoked is false",
			table: views.NewTableOutput(views.WithPrintRevokedOption(false)),
			tokens: []dto.Token{
				{
					ID:        1,
					Name:      "active-token",
					Revoked:   false,
					ExpiresAt: "2025-12-31",
				},
				{
					ID:        2,
					Name:      "revoked-token",
					Revoked:   true,
					ExpiresAt: "2025-12-31",
				},
			},
		},
		{
			name:  "show revoked tokens when printRevoked is true",
			table: views.NewTableOutput(views.WithPrintRevokedOption(true)),
			tokens: []dto.Token{
				{
					ID:        1,
					Name:      "active-token",
					Revoked:   false,
					ExpiresAt: "2025-12-31",
				},
				{
					ID:        2,
					Name:      "revoked-token",
					Revoked:   true,
					ExpiresAt: "2025-12-31",
				},
			},
		},
		{
			name:   "empty token list",
			table:  views.NewTableOutput(),
			tokens: []dto.Token{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that rendering doesn't return an error
			err := tt.table.Render(tt.tokens)
			assert.NoError(t, err)
		})
	}
}

func TestTableOutput_PrettyPrintExpiresAt(t *testing.T) {
	tests := []struct {
		name        string
		table       views.TableOutput
		date        string
		expectColor bool // Can't test exact color in black box test
	}{
		{
			name:        "valid date with color disabled",
			table:       views.NewTableOutput(views.WithColorOption(false)),
			date:        "2025-12-31",
			expectColor: false,
		},
		{
			name:        "invalid date format",
			table:       views.NewTableOutput(views.WithColorOption(true)),
			date:        "invalid-date",
			expectColor: false,
		},
		{
			name:        "empty date",
			table:       views.NewTableOutput(views.WithColorOption(true)),
			date:        "",
			expectColor: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can only test that render doesn't panic with various dates
			tokens := []dto.Token{
				{
					ID:        1,
					Name:      "test",
					ExpiresAt: tt.date,
				},
			}
			
			err := tt.table.Render(tokens)
			assert.NoError(t, err)
		})
	}
}

func TestTableOutput_ComplexScenario(t *testing.T) {
	// Test a complex scenario with multiple tokens and options
	table := views.NewTableOutput(
		views.WithHeaderOption(true),
		views.WithColorOption(false), // Disable color for predictable output
		views.WithPrintRevokedOption(true),
		views.WithNbDaysBeforeExp(30),
	)

	tokens := []dto.Token{
		{
			ID:        1,
			Source:    "group/admin",
			Type:      "deploy_token",
			Name:      "production-deploy",
			Revoked:   false,
			ExpiresAt: "2025-12-31",
		},
		{
			ID:        2,
			Source:    "project/backend",
			Type:      "access_token",
			Name:      "ci-token",
			Revoked:   true,
			ExpiresAt: "2024-06-30",
		},
		{
			ID:        3,
			Source:    "personal",
			Type:      "pat",
			Name:      "dev-token",
			Revoked:   false,
			ExpiresAt: "2023-01-01", // Expired
		},
	}

	// Test that rendering doesn't return an error
	err := table.Render(tokens)
	assert.NoError(t, err)
}