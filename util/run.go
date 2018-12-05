package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
)

// DockerRun runs a command within the project's container.
func DockerRun(ctx context.Context, p *project.Project, args ...string) error {
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

	return Run(ctx, "docker", cmd...)
}

// Run runs a system command.
func Run(ctx context.Context, command string, args ...string) error {
	return RunWithFD(ctx, os.Stdin, os.Stdout, os.Stderr, command, args...)
}

// RunWithFD is like Run, but accepts custom stdin/stdout/stderr.
func RunWithFD(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, command string, args ...string) error {
	ui.Verbose("$ %s %s", command, strings.Join(args, " "))
	cmd := exec.Command(command)
	cmd.Args = append([]string{command}, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

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
