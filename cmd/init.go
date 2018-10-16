package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [--template=...] <name>",
	Short: "Initialize an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		template, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}
		name := args[0]
		fmt.Println(name, template)

		return nil
	},
}

func init() {
	initCmd.Flags().StringP("template", "t", "cosmos-basecoin", "template to use for initialization")

	rootCmd.AddCommand(initCmd)
}
