package cmd

import (
	"fmt"
	"os"

	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var rootCmd = &cobra.Command{
	Use:   "chainkit",
	Short: "ChainKit is a toolkit for blockchain development.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Enable/Disable text coloring.
		if cmd.Flags().Changed("no-color") {
			// --no-color overrides auto detection.
			noColor, err := cmd.Flags().GetBool("no-color")
			if err != nil {
				ui.Fatal("unable to resolve flag: %v", err)
			}
			ui.EnableColors(!noColor)
		} else {
			// By default, enable colors only if stdout is a tty.
			ui.EnableColors(terminal.IsTerminal(int(os.Stdout.Fd())))
		}
	},
}

func init() {
	rootCmd.PersistentFlags().Bool("no-color", false, "disable output coloring")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
