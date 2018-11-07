package cmd

import (
	"fmt"

	"git.incubator.sh/sighup/furyctl/pkg/component"
	"git.incubator.sh/sighup/furyctl/pkg/storage"

	"github.com/spf13/viper"
)

// AgentConfig is the structure of the furyagent.yml
type AgentConfig struct {
	Storage          storage.Config          `yml:"storage"`
	ClusterComponent component.ClusterConfig `yml:"clusterComponent"`
}

// InitAgent reads the configuration file
func InitAgent(configFile string) (*AgentConfig, error) {

	viper.SetConfigFile(configFile)

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("Error reading config using config file: %v", viper.ConfigFileUsed())
	}
	// Populate the conf with configuration data
	conf := new(AgentConfig)
	err = viper.Unmarshal(conf)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %v", err)
	}
	return conf, nil
}
