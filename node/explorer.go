package node

import (
	"context"
	"fmt"

	"github.com/blocklayerhq/chainkit/config"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/util"
	"github.com/pkg/errors"
)

// explorerImage defines the container image to pull for running the Cosmos Explorer
const explorerImage = "samalba/cosmos-explorer-localdev:20181204"

func startExplorer(ctx context.Context, config *config.Config, p *project.Project) error {
	cmd := []string{
		"run", "--rm",
		"-p", fmt.Sprintf("%d:8080", config.Ports.Explorer),
		"-l", "chainkit.cosmos.explorer",
		explorerImage,
	}
	if err := util.Run(ctx, "docker", cmd...); err != nil {
		return errors.Wrap(err, "failed to start the explorer")
	}
	return nil
}
