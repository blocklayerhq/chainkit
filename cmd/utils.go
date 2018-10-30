package cmd

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/blocklayerhq/chainkit/pkg/ui"
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

func dockerRun(ctx context.Context, rootDir, name string, args ...string) error {
	dataDir := path.Join(rootDir, "data")

	daemonName := name + "d"
	cliName := name + "cli"

	// -v "${data_dir}/${APP_NAME}d:/root/.${APP_NAME}d"
	daemonDir := path.Join(dataDir, daemonName)
	daemonDirContainer := path.Join("/", "root", "."+daemonName)

	// -v "${data_dir}/${APP_NAME}cli:/root/.${APP_NAME}cli"
	cliDir := path.Join(dataDir, cliName)
	cliDirContainer := path.Join("/", "root", "."+cliName)

	cmd := []string{
		"run", "--rm",
		"-p", "26656:26656",
		"-p", "26657:26657",
		"-p", "8080:8080",
		"-v", daemonDir + ":" + daemonDirContainer,
		"-v", cliDir + ":" + cliDirContainer,
		"--name", name,
		name + ":latest",
		daemonName,
	}
	cmd = append(cmd, args...)

	return docker(ctx, rootDir, cmd...)
}

func docker(ctx context.Context, rootDir string, args ...string) error {
	return run(ctx, rootDir, "docker", args...)
}

func run(ctx context.Context, rootDir, command string, args ...string) error {
	ui.Verbose("$ %s %s", command, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, command)
	cmd.Args = append([]string{command}, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = rootDir
	return cmd.Run()
}
