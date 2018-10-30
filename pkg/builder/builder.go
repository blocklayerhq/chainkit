package builder

import (
	"context"
	"io"
	"os/exec"
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

	cmdReader := io.MultiReader(outReader, errReader)

	errCh := make(chan error)
	go func() {
		defer close(errCh)
		errCh <- b.parser.Parse(cmdReader, opts)
	}()
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}
