package ui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	spin "github.com/tj/go-spin"
	"github.com/ttacon/chalk"
	"github.com/xlab/treeprint"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	spinner = spin.New()
)

func Info(msg string, args ...interface{}) {
	fmt.Printf("%s %s\n", chalk.Bold.TextStyle(chalk.Blue.Color("==>")), chalk.Bold.TextStyle(fmt.Sprintf(msg, args...)))
}

func Verbose(msg string, args ...interface{}) {
	fmt.Println(chalk.Dim.TextStyle(fmt.Sprintf(msg, args...)))
}

func Success(msg string, args ...interface{}) {
	fmt.Printf("  %s %s\n", chalk.Bold.TextStyle(chalk.Green.Color("✔")), chalk.Bold.TextStyle(fmt.Sprintf(msg, args...)))
}

func Error(msg string, args ...interface{}) {
	fmt.Printf("  %s %s\n", chalk.Bold.TextStyle(chalk.Red.Color("✗")), chalk.Bold.TextStyle(fmt.Sprintf(msg, args...)))
}

func Fatal(msg string, args ...interface{}) {
	Error(msg, args...)
	os.Exit(1)
}

func Small(msg string) string {
	return chalk.Dim.TextStyle(msg)
}

func Emphasize(msg string) string {
	return chalk.Bold.TextStyle(chalk.Yellow.Color(msg))
}

func ConsoleWidth() int {
	width, _, err := terminal.GetSize(0)
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func Live(msg string) {
	lineLength := ConsoleWidth() - 5
	msg = strings.TrimSpace(msg)

	// Truncate length
	if len(msg) > lineLength {
		msg = msg[0:lineLength-2] + "…"
	}

	// Pad with spaces to clear previous line.
	for len(msg) < lineLength {
		msg += " "
	}

	fmt.Printf("%s %s\r", spinner.Next(), Small(msg))
}

func Tree(p string, ignore []string) error {
	root := treeprint.New()
	root.SetValue(p)
	if err := walk(p, root, ignore); err != nil {
		return err
	}
	Verbose(strings.TrimSpace(root.String()))
	return nil
}

func walk(p string, node treeprint.Tree, ignore []string) error {
	shouldIgnore := func(f os.FileInfo) bool {
		for _, i := range ignore {
			if f.Name() == i {
				return true
			}
		}
		return false
	}

	files, err := ioutil.ReadDir(p)
	if err != nil {
		return err
	}
	for _, file := range files {
		if shouldIgnore(file) {
			continue
		}
		if file.IsDir() {
			sub := node.AddBranch(file.Name())
			if err := walk(path.Join(p, file.Name()), sub, ignore); err != nil {
				return err
			}
			continue
		}

		node.AddNode(file.Name())
	}

	return nil
}
