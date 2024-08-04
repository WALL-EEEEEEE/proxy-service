package common

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

func parseYML(fname string, cfg interface{}) (err error) {
	f, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("failed to open file %s", fname)
	}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("file %s is not an valid yaml file", fname)
	}
	return nil
}

func ParseConfig(f string, cfg interface{}) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return err
	}
	err = parseYML(f, cfg)
	return err
}
