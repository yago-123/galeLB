package config

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	KeyConfigFile = "config"
)

// LoadConfig loads the configuration from the specified path
func LoadConfig[V any](path string, cfg *V) (*V, error) {
	if path == "" {
		return nil, errors.New("config path not specified")
	}

	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// GetConfigFilePath retrieves the configuration file path from command flags
func GetConfigFilePath(cmd *cobra.Command) string {
	if cmd.Flags().Changed(KeyConfigFile) {
		return viper.GetString(KeyConfigFile)
	}
	return ""
}
