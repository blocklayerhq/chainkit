package builder

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/schollz/progressbar"
)

// Parser is the build output parser
type Parser struct {
	progress *progressbar.ProgressBar
}

// Parse parses the build output
func (p *Parser) Parse(r io.Reader, opts BuildOpts) error {
	scanner := bufio.NewScanner(r)

	// Clear the console on exit.
	defer ui.Live("")

	for scanner.Scan() {
		text := stripansi.Strip(scanner.Text())
		p.processLine(text, opts)
	}

	return scanner.Err()
}

func (p *Parser) processLine(text string, opts BuildOpts) {
	// Print the current build step.
	if strings.HasPrefix(text, "Step ") {
		p.processStep(text)
	}

	// If we're in verbose mode, just print the line.
	if opts.Verbose {
		ui.Verbose(text)
		return
	}

	// Otherwise, check if it's a progress update.
	if p.processProgress(text) {
		return
	}

	// If not, live print the line (this will replace the previous output line)
	ui.Live(text)
}

func (p *Parser) processStep(text string) {
	switch {
	case strings.Contains(text, "RUN apk add --no-cache"):
		fmt.Println(ui.Small("[1/4]"), "📦 Setting up the build environment...")
	case strings.Contains(text, "RUN dep ensure"):
		fmt.Println(ui.Small("[2/4]"), "🔎 Fetching dependencies...")
	case strings.Contains(text, "RUN find vendor"):
		fmt.Println(ui.Small("[3/4]"), "🔗 Installing dependencies...")
	case strings.Contains(text, "RUN     CGO_ENABLED=0 go build"):
		fmt.Println(ui.Small("[4/4]"), "🔨 Compiling application...")
	}
}

func (p *Parser) processProgress(text string) bool {
	var (
		step  int
		total int
	)

	// Don't show progress bars on small terminals.
	if ui.ConsoleWidth() < 80 {
		return false
	}

	sr := strings.NewReader(text)
	// Check if this is a progressbar-style output (e.g. "X out of Y").
	if n, _ := fmt.Fscanf(sr, "(%d/%d) Wrote", &step, &total); n != 2 {
		return false
	}

	// This is a progress output. Create the progress bar if it doesn't exist.
	if p.progress == nil {
		// Clear current line.
		ui.Live("")
		p.progress = progressbar.NewOptions(
			total,
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "#",
				SaucerPadding: "-",
				BarStart:      "[",
				BarEnd:        "]",
			}),
			progressbar.OptionSetWidth(ui.ConsoleWidth()-20),
		)
	}

	// Update the progress bar.
	p.progress.Add(1)

	// If it's the last step, clean it up.
	if step == total {
		p.progress.Finish()
		p.progress.Clear()
		p.progress = nil
	}

	return true
}
