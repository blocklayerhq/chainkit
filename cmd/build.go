package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the application",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := cmd.Flags().GetString("cwd")
		if err != nil {
			return err
		}

		return build(cwd)
	},
}

func init() {
	buildCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(buildCmd)
}

func build(path string) error {
	return run(path, "docker", "version")
}

func run(rootDir, command string, args ...string) error {
	cmd := exec.Command(command)
	cmd.Args = append([]string{command}, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = rootDir
	return cmd.Run()
}
