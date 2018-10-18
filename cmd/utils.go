package cmd

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

func getCwd(cmd *cobra.Command) (string, error) {
	cwd, err := cmd.Flags().GetString("cwd")
	if err != nil {
		return "", err
	}
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	return filepath.Abs(cwd)
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

func docker(rootDir string, args ...string) error {
	return run(rootDir, "docker", args...)
}

func run(rootDir, command string, args ...string) error {
	cmd := exec.Command(command)
	cmd.Args = append([]string{command}, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = rootDir
	return cmd.Run()
}
