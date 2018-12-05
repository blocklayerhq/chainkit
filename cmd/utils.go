package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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
		"-p", fmt.Sprintf("%d:26656", p.Ports.TendermintP2P),
		"-p", fmt.Sprintf("%d:26657", p.Ports.TendermintRPC),
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
	return runWithFD(ctx, p, os.Stdin, os.Stdout, os.Stderr, command, args...)
}

func runWithFD(ctx context.Context, p *project.Project, stdin io.Reader, stdout, stderr io.Writer, command string, args ...string) error {
	ui.Verbose("$ %s %s", command, strings.Join(args, " "))
	cmd := exec.Command(command)
	cmd.Args = append([]string{command}, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Dir = p.RootDir

	if err := cmd.Start(); err != nil {
		return err
	}

	// We don't use exec.CommandContext here because it will
	// SIGKILL the process. Instead, we handle the context
	// on our own and try to gracefully shutdown the command.
	waitDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			cmd.Process.Signal(syscall.SIGTERM)
			select {
			case <-time.After(5 * time.Second):
				cmd.Process.Kill()
			case <-waitDone:
			}
		case <-waitDone:
		}
	}()

	err := cmd.Wait()
	close(waitDone)
	return err
}
