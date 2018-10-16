package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path"

	_ "github.com/blocklayerhq/chainkit/templates/build" // embed the static assets
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [--template=...] <name>",
	Short: "Initialize an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		template, err := cmd.Flags().GetString("template")
		if err != nil {
			return err
		}

		dest, err := cmd.Flags().GetString("dest")
		if err != nil {
			return err
		}

		return initialize(name, template, dest)
	},
}

func initialize(name, template, dest string) error {
	templates, err := fs.New()
	if err != nil {
		return err
	}

	templateRoot := path.Join("/", template)
	if _, err := templates.Open(templateRoot); err != nil {
		return errors.Wrap(err, "unable to locate template")
	}

	if err := extractFiles(templates, templateRoot, dest); err != nil {
		return err
	}

	return nil
}

func extractFiles(templates http.FileSystem, root, dest string) error {
	err := fs.Walk(templates, root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	})
	return err
}

func init() {
	initCmd.Flags().StringP("template", "t", "cosmos-basecoin", "template to use for initialization")
	initCmd.Flags().StringP("dest", "d", ".", "destination path of the generated application")

	rootCmd.AddCommand(initCmd)
}
