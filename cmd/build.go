package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the application",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir, err := getCwd(cmd)
		if err != nil {
			return err
		}

		return build(rootDir)
	},
}

func init() {
	buildCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(buildCmd)
}

func build(rootDir string) error {
	return docker(rootDir, "build", "-t", filepath.Base(rootDir), rootDir)
}
