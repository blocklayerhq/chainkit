package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the application",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		rootDir := getCwd(cmd)
		name := filepath.Base(rootDir)
		upgrade, err := cmd.Flags().GetBool("upgrade")
		if err != nil {
			ui.Fatal("Error parsing flags: %v", err)
		}

		delete, err := cmd.Flags().GetBool("delete")
		if err != nil {
			ui.Fatal("Error parsing flags: %v", err)
		}

		if delete {
			ui.Info("Deleting %s", name)
			if err := run(ctx, rootDir, "helm", "del", "--purge", name); err != nil {
				ui.Fatal("failed to delete %s: %v", name, err)
			}
			ui.Success("Deployment successfully deleted")
			os.Exit(0)
		}

		registry, err := cmd.Flags().GetString("registry")
		if err != nil {
			ui.Fatal("Error parsing flags: %v", err)
		}
		if registry == "" {
			ui.Fatal("--registry must be specified")
		}
		image := path.Join(registry, name)

		ui.Info("Deploying %s", name)

		ui.Info("Pushing %s to %s", name, ui.Emphasize(image))
		docker(ctx, rootDir, "tag", name, image)
		docker(ctx, rootDir, "push", image)
		ui.Success("Image pushed")

		ui.Info("Deploying application")
		ui.Verbose("App will be pushed to the current Kubernetes context")
		if err := run(ctx, rootDir, "kubectl", "config", "current-context"); err != nil {
			ui.Fatal("Error running kubectl: %v", err)
		}
		// FIXME: Symlink the configuration.
		if _, err := os.Stat(path.Join(rootDir, "k8s", "config")); os.IsNotExist(err) {
			if err := os.Symlink(path.Join(rootDir, "data", name+"d", "config"), path.Join(rootDir, "k8s", "config")); err != nil {
				ui.Fatal("Unable to symlink configuration files: %v", err)
			}
		}

		action := "install"
		if upgrade {
			action = "upgrade"
		}
		if err := run(ctx, rootDir, "helm", action,
			// "--dry-run", "--debug",
			"--name", name,
			"--set", fmt.Sprintf("command=%sd", name),
			"--set", fmt.Sprintf("rootDir=/root/.%sd", name),
			"--set", fmt.Sprintf("image.repository=%s", image),
			"./k8s",
		); err != nil {
			ui.Fatal("Deployment failed: %v", err)
		}

		ui.Success("Deploy successfull")
	},
}

func init() {
	deployCmd.Flags().String("cwd", ".", "specifies the current working directory")
	deployCmd.Flags().StringP("registry", "r", "", "registry for storing docker images")
	deployCmd.Flags().Bool("upgrade", false, "upgrade the deployment")
	deployCmd.Flags().Bool("delete", false, "delete the deployment")

	rootCmd.AddCommand(deployCmd)
}
