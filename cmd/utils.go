package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/schollz/progressbar"
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

func dockerBuild(ctx context.Context, rootDir, name string, verbose bool) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", name, rootDir)
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer outReader.Close()
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer errReader.Close()

	cmdReader := io.MultiReader(outReader, errReader)

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		var (
			progress *progressbar.ProgressBar
		)

		// Clear the console on exit.
		defer ui.Live("")

		for scanner.Scan() {
			text := stripansi.Strip(scanner.Text())

			// Print the current build step.
			if strings.HasPrefix(text, "Step ") {
				switch {
				case strings.Contains(text, "RUN apk add --no-cache"):
					fmt.Println(ui.Small("[1/4]"), "📦 Setting up the build environment...")
				case strings.Contains(text, "RUN dep ensure"):
					fmt.Println(ui.Small("[2/4]"), "🔎 Fetching dependencies...")
				case strings.Contains(text, "RUN find vendor"):
					fmt.Println(ui.Small("[3/4]"), "🔗 Installing dependencies...")
				case strings.Contains(text, "RUN     CGO_ENABLED=0 go build"):
					fmt.Println(ui.Small("[4/4]"), "🔨 Compiling application...")
				}
			}

			// Non-verbose output
			if !verbose {
				var (
					step  int
					total int
				)
				sr := strings.NewReader(text)

				// Check if this is a progressbar-style output (e.g. "X out of Y").
				if n, _ := fmt.Fscanf(sr, "(%d/%d) Wrote", &step, &total); n == 2 {
					if progress == nil {
						// Clear current line.
						ui.Live("")
						progress = progressbar.NewOptions(
							total,
							progressbar.OptionSetTheme(progressbar.Theme{
								Saucer:        "#",
								SaucerPadding: "-",
								BarStart:      "[",
								BarEnd:        "]",
							}),
							progressbar.OptionSetWidth(ui.ConsoleWidth()/2),
						)
					}
					progress.Add(1)
					if step == total {
						progress.Finish()
						progress.Clear()
						progress = nil
					}
				} else {
					// Otherwise, live print the line (this will replace the previous output line)
					ui.Live(text)
				}
			}

			// In verbose mode, just print the line.
			if verbose {
				ui.Verbose(text)
			}
		}
	}()
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
