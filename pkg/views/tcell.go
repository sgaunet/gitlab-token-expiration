package views

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/pterm/pterm"
	"github.com/sgaunet/gitlab-token-expiration/pkg/gitlab"
)

type TableOutput struct {
	HeaderOption bool
	ColorOption  bool
}

func NewTableOutput(headerOption, colorOption bool) TableOutput {
	return TableOutput{
		HeaderOption: headerOption,
		ColorOption:  colorOption,
	}
}

func (t TableOutput) PrintGitlabPersonalAccessToken(tokens []gitlab.PersonalAccessToken) error {
	tData := pterm.TableData{
		{"ID", "Name", "Revoked", "Last used at", "Active", "Expires at"},
	}
	for _, token := range tokens {
		tData = append(tData, []string{fmt.Sprintf("%d", token.Id), token.Name,
			prettyPrintBool(token.Revoked, true), prettyPrintGitlabTime(token.LastUsedAt),
			prettyPrintBool(token.Active, false), prettyPrintExpiresAt(token.ExpiresAt)})
	}
	// Create a table with a header and the defined data, then render it
	err := pterm.DefaultTable.WithHasHeader().WithData(tData).Render()
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
func prettyPrintBool(b bool, coloredValue bool) string {
	red := color.New(color.FgRed).SprintFunc()
	bStr := strconv.FormatBool(b)
	if b == coloredValue {
		return red(bStr)
	}
	return bStr
}

func prettyPrintExpiresAt(d string) string {
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
	now := time.Now()
	if now.After(date) {
		return red(d)
	}
	if now.AddDate(0, 0, 30).After(date) {
		return yellow(d)
	}
	return d
}

func (t TableOutput) PrintGitlabProjectAccessToken(tokens []gitlab.ProjectAccessToken) error {
	tData := pterm.TableData{
		{"ID", "Name", "Revoked", "User ID", "Last used at", "Active", "Expires at"},
	}
	for _, token := range tokens {
		tData = append(tData, []string{fmt.Sprintf("%d", token.Id), token.Name,
			prettyPrintBool(token.Revoked, true),
			fmt.Sprintf("%d", token.UserId),
			prettyPrintGitlabTime(token.LastUsedAt), prettyPrintBool(token.Active, false), prettyPrintExpiresAt(token.ExpiresAt)})
	}
	// Create a table with a header and the defined data, then render it
	err := pterm.DefaultTable.WithHasHeader().WithData(tData).Render()
	if err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}
	return nil
}

func (t TableOutput) PrintGitlabGroupAccessToken(tokens []gitlab.GroupAccessToken) error {
	tData := pterm.TableData{
		{"ID", "Name", "Revoked", "User ID", "Last used at", "Active", "Expires at"},
	}
	for _, token := range tokens {
		tData = append(tData, []string{fmt.Sprintf("%d", token.Id), token.Name,
			prettyPrintBool(token.Revoked, true),
			fmt.Sprintf("%d", token.UserId),
			prettyPrintGitlabTime(token.LastUsedAt), prettyPrintBool(token.Active, false), prettyPrintExpiresAt(token.ExpiresAt)})
	}
	// Create a table with a header and the defined data, then render it
	err := pterm.DefaultTable.WithHasHeader().WithData(tData).Render()
	if err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}
	return nil
}

func (t TableOutput) PrintGitlabGroupDeployToken(tokens []gitlab.GroupDeployToken) error {
	tData := pterm.TableData{
		{"ID", "Name", "Username", "Expires at", "Revoked", "Expired"},
	}
	for _, token := range tokens {
		tData = append(tData, []string{fmt.Sprintf("%d", token.Id), token.Name, token.Username,
			prettyPrintExpiresAt(token.ExpiresAt), prettyPrintBool(token.Revoked, true), prettyPrintBool(token.Expired, true)})
	}
	// Create a table with a header and the defined data, then render it
	err := pterm.DefaultTable.WithHasHeader().WithData(tData).Render()
	if err != nil {
		return fmt.Errorf("error rendering table: %w", err)
	}
	return nil
}
