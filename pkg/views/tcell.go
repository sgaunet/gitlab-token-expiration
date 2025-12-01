// Package views provides rendering functionality for displaying token information.
package views

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
)

// TableOutput represents a table renderer for token display.
type TableOutput struct {
	HeaderOption    bool
	ColorOption     bool
	printRevoked    bool
	nbDaysBeforeExp uint
}

// TableOutputOption is a function that configures TableOutput.
type TableOutputOption func(*TableOutput)

// WithHeaderOption configures whether to display table headers.
func WithHeaderOption(headerOption bool) TableOutputOption {
	return func(t *TableOutput) {
		t.HeaderOption = headerOption
	}
}

// WithColorOption configures whether to use colored output.
func WithColorOption(colorOption bool) TableOutputOption {
	return func(t *TableOutput) {
		t.ColorOption = colorOption
	}
}

// WithPrintRevokedOption configures whether to display revoked tokens.
func WithPrintRevokedOption(printRevoked bool) TableOutputOption {
	return func(t *TableOutput) {
		t.printRevoked = printRevoked
	}
}

// WithNbDaysBeforeExp configures the number of days before expiration to highlight.
func WithNbDaysBeforeExp(nbDaysBeforeExp uint) TableOutputOption {
	return func(t *TableOutput) {
		t.nbDaysBeforeExp = nbDaysBeforeExp
	}
}

// NewTableOutput creates a new TableOutput with the given options.
func NewTableOutput(opts ...TableOutputOption) TableOutput {
	t := TableOutput{}
	for _, opt := range opts {
		opt(&t)
	}
	return t
}

// Render displays the tokens in a table format.
func (t TableOutput) Render(tokens []dto.Token) error {
	tData := pterm.TableData{}
	if t.HeaderOption {
		tData = append(tData, []string{"ID", "Source", "Type", "Name", "Revoked", "Expires at"})
	}
	for _, token := range tokens {
		if !t.printRevoked && token.Revoked {
			continue
		}
		tData = append(tData, []string{strconv.FormatInt(token.ID, 10),
			token.Source, token.Type, token.Name,
			t.prettyPrintBool(token.Revoked, true),
			t.prettyPrintExpiresAt(token.ExpiresAt)})
	}
	// Create a table with a header and the defined data, then render it
	table := pterm.DefaultTable
	if t.HeaderOption {
		table = *table.WithHasHeader()
	}
	err := table.WithData(tData).Render()
	if err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}
	return nil
}


// prettyPrintBool returns a string representation of a boolean value
// with red color if value is equal to coloredValue.
func (t TableOutput) prettyPrintBool(b bool, coloredValue bool) string {
	red := color.New(color.FgRed).SprintFunc()
	bStr := strconv.FormatBool(b)
	if b == coloredValue && t.ColorOption {
		return red(bStr)
	}
	return bStr
}

func (t TableOutput) prettyPrintExpiresAt(d string) string {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	// convert YYYY-MM-DD to time.Time
	// then compare with now
	// if now > time.Time, print in red
	// if now + 30 days > time.Time, print in yellow
	// else return the date
	date, err := time.Parse("2006-01-02", d)
	if err != nil {
		return d
	}
	// if no color option, return the date
	if !t.ColorOption {
		return d
	}
	now := time.Now()
	if now.After(date) {
		return red(d)
	}
	if t.nbDaysBeforeExp <= math.MaxInt32 && now.AddDate(0, 0, int(t.nbDaysBeforeExp)).After(date) {
		return yellow(d)
	}
	return d
}
