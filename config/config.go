package config

import (
	"path"
)

// Config represents the node configuration.
type Config struct {
	RootDir string
	Ports   *PortMapper
}

// StateDir returns the state directory within the project.
func (c *Config) StateDir() string {
	return path.Join(c.RootDir, "state")
}

// LogFile returns the log file path
func (c *Config) LogFile() string {
	return path.Join(c.RootDir, "log")
}

// DataDir returns the data directory within the project state.
func (c *Config) DataDir() string {
	return path.Join(c.StateDir(), "data")
}

// ConfigDir returns the config directory within the project state.
func (c *Config) ConfigDir() string {
	return path.Join(c.StateDir(), "config")
}

// ConfigPath returns the path of the configuration file.
func (c *Config) ConfigPath() string {
	return path.Join(c.ConfigDir(), "config.toml")
}

// ManifestPath returns the manifest file.
func (c *Config) ManifestPath() string {
	return path.Join(c.RootDir, "chainkit.yml")
}

// GenesisPath returns the genesis path for the project.
func (c *Config) GenesisPath() string {
	return path.Join(c.ConfigDir(), "genesis.json")
}

// CLIDir returns the CLI directory within the project state.
func (c *Config) CLIDir() string {
	return path.Join(c.StateDir(), "cli")
}

// IPFSDir returns the IPFS data directory within the project state.
func (c *Config) IPFSDir() string {
	return path.Join(c.StateDir(), "ipfs")
}
