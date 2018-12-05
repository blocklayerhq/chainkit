package node

import (
	"context"
	"fmt"

	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/util"
	"github.com/pkg/errors"
)

// explorerImage defines the container image to pull for running the Cosmos Explorer
const explorerImage = "samalba/cosmos-explorer-localdev:20181204"

func startExplorer(ctx context.Context, p *project.Project) error {
	containerName := fmt.Sprintf("%s-explorer", p.Image)
	// TODO: Leaving disabled for now as it helps finding leaks.
	// defer func() {
	// 	// Failsafe: Sometimes, if we stop a `docker run --rm`, it will leave
	// 	// the container behind.
	// 	util.Run(ctx, "docker", "rm", "-f", containerName)
	// }()

	cmd := []string{
		"run", "--rm",
		"--name", containerName,
		"-p", fmt.Sprintf("%d:8080", p.Ports.Explorer),
		explorerImage,
	}
	if err := util.Run(ctx, "docker", cmd...); err != nil {
		return errors.Wrap(err, "failed to start the explorer")
	}
	return nil
}
