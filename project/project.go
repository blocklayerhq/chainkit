package project

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const manifestFile = "chainkit.yml"

type binaries struct {
	CLI    string
	Daemon string
}

// Project represents a project
type Project struct {
	Name     string
	Image    string
	Binaries *binaries
}

// New will create a new project in the given directory.
func New(name string) *Project {
	p := &Project{
		Name:  name,
		Image: fmt.Sprintf("chainkit-%s", name),
		Binaries: &binaries{
			CLI:    name + "cli",
			Daemon: name + "d",
		},
	}
	return p
}

// Save serializes the project data on disk
func (p *Project) Save(path string) error {
	ybuf, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err = fp.Write(ybuf); err != nil {
		return err
	}
	return nil
}

// Validate runs sanity checks against the project
func (p *Project) Validate() error {
	errorOut := func(field string) error {
		return fmt.Errorf("missing required field %q", field)
	}

	switch {
	case p.Name == "":
		return errorOut("name")
	case p.Image == "":
		return errorOut("image")
	case p.Binaries == nil:
		return errorOut("binaries")
	case p.Binaries.CLI == "":
		return errorOut("binaries.cli")
	case p.Binaries.Daemon == "":
		return errorOut("binaries.daemon")
	}

	return nil
}

// Parse parses a manifest.
func Parse(r io.Reader) (*Project, error) {
	errMsg := fmt.Sprintf("Cannot read manifest %q", manifestFile)

	dec := yaml.NewDecoder(r)
	p := &Project{}
	if err := dec.Decode(p); err != nil {
		return nil, errors.Wrap(err, errMsg)
	}

	if err := p.Validate(); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%s validation", manifestFile))
	}

	return p, nil
}

// Load will load a project from a given directory
func Load(dir string) (*Project, error) {
	f, err := os.Open(path.Join(dir, manifestFile))
	if err != nil {
		return nil, errors.Wrap(err, "unable to open manifest: %v")
	}
	defer f.Close()
	return Parse(f)
}
