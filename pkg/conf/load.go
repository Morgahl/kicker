package conf

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

const (
	// DefaultConfigFileName is the name the default configuration file when not specified
	DefaultConfigFileName = "kicker.yaml"
)

// LoadConf will load the configuration file from the path specified or look for DefaultConfigFileName in the launching
// directory
func LoadConf(path string) (Conf, error) {
	if path == "" {
		path = filepath.Join(".", DefaultConfigFileName)
	}

	f, err := os.Open(path)
	if err != nil {
		return Conf{}, fmt.Errorf("error loading config file a '%s': %s", path, err)
	}
	defer f.Close()

	var c Conf
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		return Conf{}, fmt.Errorf("error parsing config file '%s': %s", path, err)
	}

	if err := c.validate(); err != nil {
		return Conf{}, fmt.Errorf("error parsing config file '%s': %s", path, err)
	}

	return c, nil
}
