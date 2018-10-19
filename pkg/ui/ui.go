package ui

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ttacon/chalk"
	"github.com/xlab/treeprint"
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

func Tree(p string) error {
	root := treeprint.New()
	root.SetValue(p)
	if err := walk(p, root); err != nil {
		return err
	}
	Verbose(strings.TrimSpace(root.String()))
	return nil
}
func walk(p string, node treeprint.Tree) error {
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			sub := node.AddBranch(file.Name())
			if err := walk(path.Join(p, file.Name()), sub); err != nil {
				return err
			}
			continue
		}

		node.AddNode(file.Name())
	}

	return nil
}
