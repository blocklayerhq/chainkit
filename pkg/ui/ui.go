package ui

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/mitchellh/colorstring"
	spin "github.com/tj/go-spin"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	spinner  = spin.New()
	colorize = colorstring.Colorize{
		Colors: colorstring.DefaultColors,
		Reset:  true,
	}
)

func init() {
	spinner.Set(spin.Spin1)
}

// EnableColors enables or disable output coloring.
func EnableColors(enabled bool) {
	colorize.Disable = !enabled
}

// Info prints an info message.
func Info(msg string, args ...interface{}) {
	fmt.Printf(colorize.Color("[bold][blue]==> [reset][bold]%s\n"), fmt.Sprintf(msg, args...))
}

// Verbose prints a verbose message.
func Verbose(msg string, args ...interface{}) {
	fmt.Printf(colorize.Color("[dim]%s\n"), fmt.Sprintf(msg, args...))
}

// Success prints a success message.
func Success(msg string, args ...interface{}) {
	fmt.Printf(colorize.Color("[bold][green]✔[reset][bold] %s\n"), fmt.Sprintf(msg, args...))
}

// Error prints an error message.
func Error(msg string, args ...interface{}) {
	fmt.Printf(colorize.Color("[bold][red]✗[reset][bold] %s\n"), fmt.Sprintf(msg, args...))
}

// Fatal prints an error message and exits.
func Fatal(msg string, args ...interface{}) {
	Error(msg, args...)
	os.Exit(1)
}

// Small returns a `small` colored string.
func Small(msg string) string {
	return colorize.Color("[dim]" + msg)
}

// Emphasize returns a `emphasized` colored string.
func Emphasize(msg string) string {
	return colorize.Color("[bold][yellow]" + msg)
}

// ConsoleWidth returns the console's width.
func ConsoleWidth() int {
	width, _, err := terminal.GetSize(0)
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

// Live is used to print a live message. Subsequent calls will replace the line.
func Live(msg string) {
	// Format the message.
	msg = fmt.Sprintf("%s %s", spinner.Next(), strings.TrimSpace(msg))

	// Get the actual console width.
	lineLength := ConsoleWidth()

	// Shorten the message until it fits.
	for utf8.RuneCountInString(msg) > lineLength {
		msg = msg[0:len(msg)-4] + "…"
	}

	// Pad the message with spaces until it takes the entire line.
	// This is in order to clear the previous line.
	for utf8.RuneCountInString(msg) < lineLength {
		msg = msg + " "
	}

	fmt.Printf("%s\r", Small(msg))
}
