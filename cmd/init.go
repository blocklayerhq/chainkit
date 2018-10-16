package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	_ "github.com/blocklayerhq/chainkit/templates/build" // embed the static assets
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize an application",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		dest, err := cmd.Flags().GetString("dest")
		if err != nil {
			return err
		}

		return initialize(name, dest)
	},
}

func initialize(name, dest string) error {
	templates, err := fs.New()
	if err != nil {
		return err
	}

	if err := extractFiles(templates, dest); err != nil {
		return err
	}

	return nil
}

func extractFiles(templates http.FileSystem, dest string) error {
	err := fs.Walk(templates, "/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return extractFile(templates, path, dest, fi)
	})
	return err
}

func extractFile(templates http.FileSystem, src, dst string, fi os.FileInfo) error {
	dstPath := path.Join(dst, src)
	fmt.Println(dstPath)

	if fi.IsDir() {
		return os.MkdirAll(dstPath, fi.Mode())
	}

	data, err := fs.ReadFile(templates, src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dstPath, data, fi.Mode())
}

func init() {
	initCmd.Flags().StringP("dest", "d", ".", "destination path of the generated application")

	rootCmd.AddCommand(initCmd)
}
