package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blocklayerhq/chainkit/pkg/httpfs"
	"github.com/blocklayerhq/chainkit/pkg/ui"
	"github.com/blocklayerhq/chainkit/templates"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type templateContext struct {
	Name    string
	RootDir string
	GoPkg   string
}

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize an application",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		rootDir := path.Join(getCwd(cmd), name)
		initialize(name, rootDir)
	},
}

func init() {
	initCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(initCmd)
}

func initialize(name, rootDir string) {
	ui.Info("Creating a new blockchain app in %s", ui.Emphasize(rootDir))

	if err := scaffold(name, rootDir); err != nil {
		ui.Fatal("Failed to initialize: %v", err)
	}

	build(name, rootDir, false)

	ui.Info("Generating configuration and gensis")
	if err := dockerRun(rootDir, name, "init"); err != nil {
		ui.Fatal("Initialization failed: %v", err)
	}
	if err := ui.Tree(path.Join(rootDir, "data")); err != nil {
		ui.Fatal("%v", err)
	}

	ui.Success("Sucess! Created %s at %s", ui.Emphasize(name), ui.Emphasize(rootDir))
	printGettingStarted(name)
}

func printGettingStarted(name string) {
	fmt.Printf(`
Inside that directory, you can run several commands:

  %s
    Starts the application.

  %s
    Build the application.

We suggest that you begin by typing:
  %s %s
  %s
`,
		ui.Emphasize("chainkit start"),
		ui.Emphasize("chainkit build"),
		ui.Emphasize("cd"),
		name,
		ui.Emphasize("chainkit start"),
	)
}

func scaffold(name, rootDir string) error {
	ui.Info("Scaffolding base application")

	gosource := goSrc()

	if !strings.HasPrefix(rootDir, gosource) {
		return fmt.Errorf("you must run this command within your GOPATH (%q)", goPath())
	}

	// Make sure the destination path doesn't exist.
	if _, err := os.Stat(rootDir); !os.IsNotExist(err) {
		return fmt.Errorf("destination path %q already exists", rootDir)
	}

	ctx := &templateContext{
		Name:    name,
		RootDir: rootDir,
		GoPkg:   strings.TrimPrefix(rootDir, gosource+"/"),
	}

	if err := extractFiles(ctx, rootDir); err != nil {
		return err
	}
	if err := ui.Tree(rootDir); err != nil {
		return err
	}

	return nil
}

func extractFiles(ctx *templateContext, dest string) error {
	err := httpfs.Walk(templates.Assets, "/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
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
	if filepath.Ext(dstPath) == ".tpl" {
		dstPath = strings.TrimSuffix(dstPath, ".tpl")
	}

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
