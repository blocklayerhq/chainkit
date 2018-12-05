package cmd

import (
	"os"
	"path"
	"path/filepath"

	"github.com/blocklayerhq/chainkit/ui"
	"github.com/spf13/cobra"
)

func getCwd(cmd *cobra.Command) string {
	cwd, err := cmd.Flags().GetString("cwd")
	if err != nil {
		ui.Fatal("unable to resolve --cwd: %v", err)
		return ""
	}
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			ui.Fatal("unable to determine current directory: %v", err)
			return ""
		}
	}
	abs, err := filepath.Abs(cwd)
	if err != nil {
		ui.Fatal("unable to parse %q: %v", cwd, err)
	}
	return abs
}

func goPath() string {
	p := os.Getenv("GOPATH")
	if p != "" {
		return p
	}
	return path.Join(os.Getenv("HOME"), "go")
}

func goSrc() string {
	return path.Join(goPath(), "src")
}
