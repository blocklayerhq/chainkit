package project

import (
	"path"
	"path/filepath"
)

// Project represents a project
type Project struct {
	Name    string
	RootDir string
}

// Create will create a new project in the given directory.
func Create(dir, name string) (*Project, error) {
	return &Project{
		Name:    name,
		RootDir: path.Join(dir, name),
	}, nil
}

// Load will load a project from a given directory
func Load(dir string) (*Project, error) {
	return &Project{
		Name:    filepath.Base(dir),
		RootDir: dir,
	}, nil
}
