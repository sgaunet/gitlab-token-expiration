package views

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
	"github.com/sgaunet/gitlab-token-expiration/pkg/dto"
)

type TableOutput struct {
	HeaderOption bool
	ColorOption  bool
	printRevoked bool
}

type TableOutputOption func(*TableOutput)

func WithHeaderOption(headerOption bool) TableOutputOption {
	return func(t *TableOutput) {
		t.HeaderOption = headerOption
	}
}

func WithColorOption(colorOption bool) TableOutputOption {
	return func(t *TableOutput) {
		t.ColorOption = colorOption
	}
}

func WithPrintRevokedOption(printRevoked bool) TableOutputOption {
	return func(t *TableOutput) {
		t.printRevoked = printRevoked
	}
}

func NewTableOutput(opts ...TableOutputOption) TableOutput {
	t := TableOutput{}
	for _, opt := range opts {
		opt(&t)
	}
	return t
}

func (t TableOutput) Render(tokens []dto.Token) error {
	tData := pterm.TableData{}
	if t.HeaderOption {
		tData = append(tData, []string{"ID", "Source", "Type", "Name", "Revoked", "Expires at"})
	}
	for _, token := range tokens {
		if !t.printRevoked && token.Revoked {
			continue
		}
		tData = append(tData, []string{fmt.Sprintf("%d", token.ID),
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

func prettyPrintGitlabTime(t string) string {
	if t == "" {
		return ""
	}
	if len(t) < 19 {
		return t
	}
	// 2021-09-29T14:00:00Z to 2021-09-29 14:00:00
	return fmt.Sprintf("%s %s", t[:10], t[11:19])
}

// prettyPrintBool returns a string representation of a boolean value
// with red color if value is equal to coloredValue
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
	if now.AddDate(0, 0, 30).After(date) {
		return yellow(d)
	}
	return d
}
