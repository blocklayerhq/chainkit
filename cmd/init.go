package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/blocklayerhq/chainkit/pkg/httpfs"
	"github.com/blocklayerhq/chainkit/templates"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type templateContext struct {
	Name    string
	WorkDir string
}

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

func init() {
	initCmd.Flags().StringP("dest", "d", ".", "destination path of the generated application")

	rootCmd.AddCommand(initCmd)
}

func initialize(name, dest string) error {
	workDir := path.Join(dest, name)

	// Make sure the destination path doesn't exist.
	if _, err := os.Stat(workDir); !os.IsNotExist(err) {
		return fmt.Errorf("destination path %q already exists", workDir)
	}

	ctx := &templateContext{
		Name:    name,
		WorkDir: workDir,
	}

	if err := extractFiles(ctx, workDir); err != nil {
		return err
	}

	return nil
}

func extractFiles(ctx *templateContext, dest string) error {
	err := httpfs.Walk(templates.Assets, "/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path)
		return extractFile(ctx, path, dest, fi)
	})
	return err
}

func extractFile(ctx *templateContext, src, dst string, fi os.FileInfo) error {
	// Templatize the file name.
	parsedSrc, err := templatize(ctx, src)
	if err != nil {
		return err
	}

	dstPath := path.Join(dst, string(parsedSrc))
	fmt.Println(src, dst, dstPath)

	if fi.IsDir() {
		return os.MkdirAll(dstPath, fi.Mode())
	}

	data, err := httpfs.ReadFile(templates.Assets, src)
	if err != nil {
		return errors.Wrap(err, "unable to read template file")
	}
	output, err := templatize(ctx, string(data))
	if err != nil {
		return errors.Wrap(err, "unable to templetaize")
	}

	if err := ioutil.WriteFile(dstPath, output, fi.Mode()); err != nil {
		return errors.Wrap(err, "unable to write to destination")
	}

	return nil
}

func templatize(ctx *templateContext, input string) ([]byte, error) {
	t, err := template.New("chainkit").Parse(input)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
