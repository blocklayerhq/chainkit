package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type projectBinaries struct {
	CLI    string
	Daemon string
}

// Project represents a project
type Project struct {
	Name     string
	RootDir  string `yaml:"-"`
	Binaries *projectBinaries
}

// ChainkitManifest defines the name of the manifest file
const ChainkitManifest = "chainkit.yml"

// New will create a new project in the given directory.
func New(dir, name string) *Project {
	return &Project{
		Name: name,
		Binaries: &projectBinaries{
			CLI:    name + "cli",
			Daemon: name + "d",
		},
		RootDir: path.Join(dir, name),
	}
}

// Save serializes the project data on disk
func (p *Project) Save() error {
	ybuf, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	fp, err := os.Create(path.Join(p.RootDir, ChainkitManifest))
	if err != nil {
		return err
	}
	if _, err = fp.Write(ybuf); err != nil {
		return err
	}
	return nil
}

// Load will load a project from a given directory
func Load(dir string) (*Project, error) {
	errMsg := fmt.Sprintf("Cannot read manifest \"%s\"", ChainkitManifest)
	data, err := ioutil.ReadFile(path.Join(dir, ChainkitManifest))
	if err != nil {
		return nil, errors.Wrap(err, errMsg)
	}
	p := &Project{}
	if err = yaml.Unmarshal(data, p); err != nil {
		return nil, errors.Wrap(err, errMsg)
	}
	p.RootDir = dir
	return p, nil
}
