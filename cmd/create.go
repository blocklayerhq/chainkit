package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blocklayerhq/chainkit/builder"
	"github.com/blocklayerhq/chainkit/httpfs"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/templates"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type templateContext struct {
	Name    string
	RootDir string
	GoPkg   string
}

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an application",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		p := project.New(getCwd(cmd), name)
		create(p)
	},
}

func init() {
	createCmd.Flags().String("cwd", ".", "specifies the current working directory")

	rootCmd.AddCommand(createCmd)
}

func create(p *project.Project) {
	ctx := context.Background()

	ui.Info("Creating a new blockchain app in %s", ui.Emphasize(p.RootDir))

	if err := scaffold(p); err != nil {
		ui.Fatal("Failed to initialize: %v", err)
	}

	b := builder.New(p)
	if err := b.Build(ctx, builder.BuildOpts{}); err != nil {
		ui.Fatal("Failed to build the application: %v", err)
	}

	ui.Success("Success! Created %s at %s", ui.Emphasize(p.Name), ui.Emphasize(p.RootDir))
	printGettingStarted(p)
}

func printGettingStarted(p *project.Project) {
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
		p.Name,
		ui.Emphasize("chainkit start"),
	)
}

func scaffold(p *project.Project) error {
	ui.Info("Scaffolding base application")

	gosource := goSrc()

	if !strings.HasPrefix(p.RootDir, gosource) {
		return fmt.Errorf("you must run this command within your GOPATH (%q)", goPath())
	}

	// Make sure the destination path doesn't exist.
	if _, err := os.Stat(p.RootDir); !os.IsNotExist(err) {
		return fmt.Errorf("destination path %q already exists", p.RootDir)
	}

	ctx := &templateContext{
		Name:    p.Name,
		RootDir: p.RootDir,
		GoPkg:   strings.TrimPrefix(p.RootDir, gosource+"/"),
	}

	if err := extractFiles(ctx, p); err != nil {
		return err
	}
	if err := ui.Tree(p.RootDir, []string{"k8s"}); err != nil {
		return err
	}

	return nil
}

func extractFiles(ctx *templateContext, p *project.Project) error {
	err := httpfs.Walk(templates.Assets, "/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return extractFile(ctx, path, p, fi)
	})
	return err
}

func extractFile(ctx *templateContext, src string, p *project.Project, fi os.FileInfo) error {
	// Templatize the file name.
	parsedSrc, err := templatize(ctx, src, src)
	if err != nil {
		return err
	}

	dstPath := path.Join(p.RootDir, string(parsedSrc))
	if fi.IsDir() {
		return os.MkdirAll(dstPath, fi.Mode())
	}

	// Save the project manifest on disk
	if err := p.Save(); err != nil {
		errMsg := fmt.Sprintf("Cannot create \"%s\"", project.ChainkitManifest)
		return errors.Wrap(err, errMsg)
	}

	data, err := httpfs.ReadFile(templates.Assets, src)
	if err != nil {
		return errors.Wrap(err, "unable to read template file")
	}

	// Handle templates
	if filepath.Ext(dstPath) == ".tmpl" {
		// Parse template
		data, err = templatize(ctx, dstPath, string(data))
		if err != nil {
			return errors.Wrap(err, "unable to templetaize")
		}

		// Remove .tpl from the file path
		dstPath = strings.TrimSuffix(dstPath, ".tmpl")

	}

	if err := ioutil.WriteFile(dstPath, data, fi.Mode()); err != nil {
		return errors.Wrap(err, "unable to write to destination")
	}

	return nil
}

func templatize(ctx *templateContext, name, input string) ([]byte, error) {
	t, err := template.New(name).Parse(input)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
