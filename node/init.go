package node

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/blocklayerhq/chainkit/config"
	"github.com/blocklayerhq/chainkit/project"
	"github.com/blocklayerhq/chainkit/ui"
	"github.com/blocklayerhq/chainkit/util"
	"github.com/pkg/errors"
)

func initialize(ctx context.Context, config *config.Config, p *project.Project) error {
	_, err := os.Stat(config.GenesisPath())

	// Skip initialization if already initialized.
	if err == nil {
		return nil
	}

	// Make sure we got an ErrNotExist - fail otherwise.
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	ui.Info("Generating configuration and genesis files")
	if err := util.DockerRun(ctx, config, p, "init"); err != nil {
		//NOTE: some cosmos app (e.g. Gaia) take a --moniker option in the init command
		// if the normal init fail, rerun with `--moniker $(hostname)`
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		if err := util.DockerRun(ctx, config, p, "init", "--moniker", hostname); err != nil {
			return err
		}
	}

	if err := fixFsPermissions(ctx, config, p); err != nil {
		return err
	}

	if err := ui.Tree(config.StateDir(), []string{"ipfs"}); err != nil {
		return errors.Wrap(err, "Cannot print source tree")
	}

	return nil
}

func fixFsPermissions(ctx context.Context, config *config.Config, p *project.Project) error {
	u, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "Cannot get user id")
	}
	daemonDir := path.Join("/", "root", "."+p.Binaries.Daemon)
	cliDir := path.Join("/", "root", "."+p.Binaries.CLI)
	user := fmt.Sprintf("%s:%s", u.Uid, u.Gid)
	cmd := []string{
		"run", "--rm",
		"-v", config.StateDir() + ":" + daemonDir,
		"-v", config.CLIDir() + ":" + cliDir,
		"--name", p.Image,
		p.Image + ":latest",
		"chown", "-R", user, daemonDir, cliDir,
	}
	if err := util.Run(ctx, "docker", cmd...); err != nil {
		return errors.Wrap(err, "Cannot change directories permissions")
	}
	return nil
}

// updateConfig updates the config file for the node before starting.
func updateConfig(file string, vars map[string]string) error {
	config, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	output := bytes.NewBufferString("")
	scanner := bufio.NewScanner(bytes.NewReader(config))
	for scanner.Scan() {
		line := scanner.Text()
		// Scan vars to replace in the current line
		for k, v := range vars {
			if strings.HasPrefix(line+" = ", k) {
				line = fmt.Sprintf("%s = %s", k, v)
			}
		}
		if _, err := fmt.Fprintln(output, line); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, output); err != nil {
		return err
	}

	return nil
}
