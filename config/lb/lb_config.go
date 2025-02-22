package lb

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	KeyNodeHealthChecksBeforeRouting = "node_health_checks_before_routing"
	KeyNodeHealthChecksTimeout       = "node_health_checks_timeout"

	KeyConfigFile = "config"
)

const (
	DefaultNodeHealthChecksBeforeRouting = 3
	DefaultNodeHealthChecksTimeout       = 10 * time.Second

	DefaultConfigFile = "lb_config.toml"
)

type Config struct {
	// NodeHealthChecksBeforeRouting number of continuous health checks passed before node is added to routing ring.
	NodeHealthChecksBeforeRouting uint `mapstructure:"node_health_checks_before_routing"`
	// NodeHealthChecksTimeout defines the maximum time allowed between health checks before a node is considered
	// unresponsive. Nodes must send health checks at least once every half of this duration. Minimum time allowed
	// is 1s.
	NodeHealthChecksTimeout time.Duration `mapstructure:"node_health_checks_timeout"`

	Logger *logrus.Logger
}

func NewDefaultConfig() *Config {
	return &Config{
		NodeHealthChecksBeforeRouting: DefaultNodeHealthChecksBeforeRouting,
		NodeHealthChecksTimeout:       DefaultNodeHealthChecksTimeout,
	}
}

// LoadConfig loads the configuration from the specified path
func LoadConfig(path string) (*Config, error) {
	var cfg Config

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

	cfg.Logger = logrus.New()

	return &cfg, nil
}

// InitConfig initializes the configuration for the command
func InitConfig(cmd *cobra.Command) *Config {
	cfg, err := LoadConfig(GetConfigFilePath(cmd))
	if err != nil {
		cfg = NewDefaultConfig()

		cfg.Logger.Warnf("failed to load default config: %v", err)
		cfg.Logger.Infof("relying on default configuration")
	}

	AddConfigFlags(cmd)
	ApplyFlagsToConfig(cmd, cfg)

	return cfg
}

// AddConfigFlags defines the configuration flags for the command
func AddConfigFlags(cmd *cobra.Command) {
	cmd.Flags().Uint(KeyNodeHealthChecksBeforeRouting, DefaultNodeHealthChecksBeforeRouting, "Continuous node health checks that must be received before starting routing traffic to the node")
	cmd.Flags().Duration(KeyNodeHealthChecksTimeout, DefaultNodeHealthChecksTimeout, "Maximum time between health checks before node is considered unresponsive and traffic is re-routed")
	cmd.Flags().String(KeyConfigFile, DefaultConfigFile, "config file (default is $PWD/lb_config.toml)")

	_ = viper.BindPFlag(KeyNodeHealthChecksBeforeRouting, cmd.Flags().Lookup(KeyNodeHealthChecksBeforeRouting))
	_ = viper.BindPFlag(KeyNodeHealthChecksTimeout, cmd.Flags().Lookup(KeyNodeHealthChecksTimeout))
	_ = viper.BindPFlag(KeyConfigFile, cmd.Flags().Lookup(KeyConfigFile))
}

// GetConfigFilePath retrieves the configuration file path from command flags
func GetConfigFilePath(cmd *cobra.Command) string {
	if cmd.Flags().Changed(KeyConfigFile) {
		return viper.GetString(KeyConfigFile)
	}
	return ""
}

func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed(KeyNodeHealthChecksBeforeRouting) {
		cfg.NodeHealthChecksBeforeRouting = viper.GetUint(KeyNodeHealthChecksBeforeRouting)
	}
	if cmd.Flags().Changed(KeyNodeHealthChecksTimeout) {
		cfg.NodeHealthChecksTimeout = viper.GetDuration(KeyNodeHealthChecksTimeout)
	}
}
