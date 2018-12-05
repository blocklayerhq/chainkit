package node

import (
	"context"
	"os"

	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/blocklayerhq/chainkit/util"
)

func initialize(ctx context.Context, p *project.Project) error {
	_, err := os.Stat(p.StateDir())

	// Skip initialization if already initialized.
	if err == nil {
		return nil
	}

	// Make sure we got an ErrNotExist - fail otherwise.
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	ui.Info("Generating configuration and genesis files")
	if err := util.DockerRun(ctx, p, "init"); err != nil {
		//NOTE: some cosmos app (e.g. Gaia) take a --moniker option in the init command
		// if the normal init fail, rerun with `--moniker $(hostname)`
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		if err := util.DockerRun(ctx, p, "init", "--moniker", hostname); err != nil {
			return err
		}
	}

	if err := ui.Tree(p.StateDir(), nil); err != nil {
		return err
	}

	return nil
}
