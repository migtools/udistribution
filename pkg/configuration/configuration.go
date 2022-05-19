package configuration

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/distribution/distribution/v3/configuration"
	"gopkg.in/yaml.v2"
)

// v0_1Configuration is a Version 0.1 Configuration struct
// This is currently aliased to Configuration, as it is the current version
// https://github.com/distribution/distribution/blob/32ccbf193d5016bd0908d2eb636333d3cca22534/configuration/configuration.go#L355-L357
type v0_1Configuration configuration.Configuration

// Get configuration given array of strings as environment variables and current configuration object
// a modification from Parse https://github.com/distribution/distribution/blob/32ccbf193d5016bd0908d2eb636333d3cca22534/configuration/configuration.go#L649-L695
func ParseEnvironment(config *configuration.Configuration, env []string) (*configuration.Configuration, error) {
	// copy current environment to restore later
	oldEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, e := range oldEnv {
			os.Setenv(e[:strings.Index(e, "=")], e[strings.Index(e, "=")+1:])
		}
	}()

	// set environment variables
	for _, e := range env {
		os.Setenv(e[:strings.Index(e, "=")], e[strings.Index(e, "=")+1:])
	}
	in, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}
	// parse configuration and enviroment variables from parameters
	p := GetParser()

	config = new(configuration.Configuration)
	err = p.Parse(in, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetParser() *configuration.Parser {
	return configuration.NewParser("registry", []configuration.VersionedParseInfo{
		{
			Version: configuration.MajorMinorVersion(0, 1),
			ParseAs: reflect.TypeOf(v0_1Configuration{}),
			ConversionFunc: func(c interface{}) (interface{}, error) {
				if v0_1, ok := c.(*v0_1Configuration); ok {
					if v0_1.Log.Level == configuration.Loglevel("") {
						if v0_1.Loglevel != configuration.Loglevel("") {
							v0_1.Log.Level = v0_1.Loglevel
						} else {
							v0_1.Log.Level = configuration.Loglevel("info")
						}
					}
					if v0_1.Loglevel != configuration.Loglevel("") {
						v0_1.Loglevel = configuration.Loglevel("")
					}
					if v0_1.Storage.Type() == "" {
						return nil, errors.New("no storage configuration provided")
					}
					return (*configuration.Configuration)(v0_1), nil
				}
				return nil, fmt.Errorf("expected *v0_1Configuration, received %#v", c)
			},
		},
	})
}