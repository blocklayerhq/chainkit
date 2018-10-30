package builder

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os/exec"

	"github.com/blocklayerhq/chainkit/pkg/ui"
)

// Builder is a wrapper around `docker build` which provides a better UX.
type Builder struct {
	rootDir string
	name    string
	parser  *Parser
}

// BuildOpts contains a list of build options.
type BuildOpts struct {
	Verbose bool
}

// New creates a new Builder.
func New(rootDir, name string) *Builder {
	return &Builder{
		rootDir: rootDir,
		name:    name,
		parser:  &Parser{},
	}
}

// Build executes a build.
func (b *Builder) Build(ctx context.Context, opts BuildOpts) error {
	cmd := exec.CommandContext(ctx, "docker", "build", "-t", b.name, b.rootDir)
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

	// Combine stdout and stderr into a single reader.
	cmdReader := io.MultiReader(outReader, errReader)

	// Keep the build output as a buffer.
	// We'll need it to log build errors.
	var output bytes.Buffer
	tee := io.TeeReader(cmdReader, &output)

	errCh := make(chan error)
	go func() {
		defer close(errCh)
		errCh <- b.parser.Parse(tee, opts)
	}()
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		b.buildLog(output)
		return err
	}

	if err := <-errCh; err != nil {
		b.buildLog(output)
		return err
	}

	return nil
}

func (b *Builder) buildLog(output bytes.Buffer) error {
	logfile, err := ioutil.TempFile("", "chainkit-build.*.log")
	if err != nil {
		return err
	}
	defer logfile.Close()

	if _, err := logfile.Write(output.Bytes()); err != nil {
		return err
	}

	ui.Error("A complete log of this build can be found in:")
	ui.Error("    %s", logfile.Name())

	return nil
}
