package cmd

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/blocklayerhq/chainkit/project"
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

func dockerRun(ctx context.Context, p *project.Project, args ...string) error {
	var (
		daemonDirContainer = path.Join("/", "root", "."+p.Binaries.Daemon)
		cliDirContainer    = path.Join("/", "root", "."+p.Binaries.CLI)
	)

	cmd := []string{
		"run", "--rm",
		"-p", "26656:26656",
		"-p", "26657:26657",
		"-v", p.StateDir() + ":" + daemonDirContainer,
		"-v", p.CLIDir() + ":" + cliDirContainer,
		"--name", p.Image,
		p.Image + ":latest",
		p.Binaries.Daemon,
	}
	cmd = append(cmd, args...)

	return docker(ctx, p, cmd...)
}

func docker(ctx context.Context, p *project.Project, args ...string) error {
	return run(ctx, p, "docker", args...)
}

func run(ctx context.Context, p *project.Project, command string, args ...string) error {
	ui.Verbose("$ %s %s", command, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, command)
	cmd.Args = append([]string{command}, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = p.RootDir
	return cmd.Run()
}
