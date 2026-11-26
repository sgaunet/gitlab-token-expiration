package cmd

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pterm/pterm"
)

// progress shows the progress of a step on stderr. When stderr is a terminal,
// it is a pterm spinner. Otherwise (CI logs, a file), a single plain line is
// printed when the step ends, without colors or redraws: pterm itself does not
// check whether its output is a terminal.
type progress struct {
	text    string
	spinner *pterm.SpinnerPrinter
}

// startProgress starts showing the step described by text.
func startProgress(text string) *progress {
	p := &progress{text: text}
	if stderrIsTerminal() {
		spinner, err := pterm.DefaultSpinner.Start(text)
		if err == nil {
			p.spinner = spinner
		}
	}
	return p
}

// Success ends the step with a success message.
func (p *progress) Success(msg string) {
	if p.spinner != nil {
		p.spinner.Success(msg)
		return
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", p.text, msg)
}

// Fail ends the step with a failure message.
func (p *progress) Fail(msg string) {
	if p.spinner != nil {
		p.spinner.Fail(msg)
		return
	}
	fmt.Fprintf(os.Stderr, "%s: %s\n", p.text, msg)
}

func stderrIsTerminal() bool {
	fd := os.Stderr.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
