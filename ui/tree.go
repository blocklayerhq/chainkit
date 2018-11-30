package ui

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/xlab/treeprint"
)

// Tree prints a source tree.
func Tree(p string, ignore []string) error {
	root := treeprint.New()
	root.SetValue(p)
	if err := walk(p, root, ignore); err != nil {
		return err
	}
	Verbose(strings.TrimSpace(root.String()))
	return nil
}

func walk(p string, node treeprint.Tree, ignore []string) error {
	shouldIgnore := func(f os.FileInfo) bool {
		for _, i := range ignore {
			if f.Name() == i {
				return true
			}
		}
		return false
	}

	files, err := ioutil.ReadDir(p)
	if err != nil {
		return err
	}
	for _, file := range files {
		if shouldIgnore(file) {
			continue
		}
		if file.IsDir() {
			sub := node.AddBranch(file.Name())
			if err := walk(path.Join(p, file.Name()), sub, ignore); err != nil {
				return err
			}
			continue
		}

		node.AddNode(file.Name())
	}

	return nil
}
