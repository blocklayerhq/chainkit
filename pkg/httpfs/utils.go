package httpfs

import (
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
)

// Walk walks the file tree rooted at root, calling walkFn for each file or directory in the tree, including root.
// All errors that arise visiting files and directories are filtered by walkFn. The files are walked in lexical order, which makes the output deterministic but means that for very large directories Walk can be inefficient. Walk does not follow symbolic links.
func Walk(hfs http.FileSystem, root string, walkFn filepath.WalkFunc) error {
	dh, err := hfs.Open(root)
	if err != nil {
		return err
	}
	di, err := dh.Stat()
	if err != nil {
		return err
	}
	fis, err := dh.Readdir(-1)
	dh.Close()
	if err = walkFn(root, di, err); err != nil {
		if err == filepath.SkipDir {
			return nil
		}
		return err
	}
	for _, fi := range fis {
		fn := path.Join(root, fi.Name())
		if fi.IsDir() {
			if err = Walk(hfs, fn, walkFn); err != nil {
				if err == filepath.SkipDir {
					continue
				}
				return err
			}
			continue
		}
		if err = walkFn(fn, fi, nil); err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}
	}
	return nil
}

// ReadFile reads the file named by filename and returns the contents.
// A successful call returns err == nil, not err == EOF. Because ReadFile reads the whole file, it does not treat an EOF from Read as an error to be reported.
func ReadFile(hfs http.FileSystem, name string) ([]byte, error) {
	fh, err := hfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	return ioutil.ReadAll(fh)
}
