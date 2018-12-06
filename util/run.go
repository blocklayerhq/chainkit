package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/blocklayerhq/chainkit/config"
	"github.com/blocklayerhq/chainkit/project"
)

// DockerRun runs a command within the project's container.
func DockerRun(ctx context.Context, config *config.Config, p *project.Project, args ...string) error {
	return DockerRunWithFD(ctx, config, p, os.Stdin, os.Stdout, os.Stderr, args...)
}

// DockerRunWithFD is like DockerRun but accepts stdin/stdout/stderr.
func DockerRunWithFD(ctx context.Context, config *config.Config, p *project.Project, stdin io.Reader, stdout, stderr io.Writer, args ...string) error {
	var (
		daemonDirContainer = path.Join("/", "root", "."+p.Binaries.Daemon)
		cliDirContainer    = path.Join("/", "root", "."+p.Binaries.CLI)
	)

	cmd := []string{
		"run", "--rm",
		"-p", fmt.Sprintf("%d:26656", config.Ports.TendermintP2P),
		"-p", fmt.Sprintf("%d:26657", config.Ports.TendermintRPC),
		"-v", config.StateDir() + ":" + daemonDirContainer,
		"-v", config.CLIDir() + ":" + cliDirContainer,
		p.Image + ":latest",
		p.Binaries.Daemon,
	}
	cmd = append(cmd, args...)

	return RunWithFD(ctx, stdin, stdout, stderr, "docker", cmd...)
}

// Run runs a system command.
func Run(ctx context.Context, command string, args ...string) error {
	return RunWithFD(ctx, os.Stdin, os.Stdout, os.Stderr, command, args...)
}

// RunWithFD is like Run, but accepts custom stdin/stdout/stderr.
func RunWithFD(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, command string, args ...string) error {
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
