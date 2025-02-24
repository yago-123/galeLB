package lb

import (
	"time"

	"github.com/sirupsen/logrus"
	common "github.com/yago-123/galelb/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	KeyNodeHealthChecksBeforeRouting = "node_health.checks_before_routing"
	KeyNodeHealthChecksTimeout       = "node_health.checks_timeout"
)

const (
	DefaultNodeHealthChecksBeforeRouting = 3
	DefaultNodeHealthChecksTimeout       = 10 * time.Second

	DefaultConfigFile = "lb.toml"
)

type Config struct {
	NodeHealth NodeHealth `mapstructure:"node_health"`
	Logger     *logrus.Logger
}

type NodeHealth struct {
	// ChecksBeforeRouting number of continuous health checks passed before node is added to routing ring.
	ChecksBeforeRouting uint `mapstructure:"checks_before_routing"`
	// ChecksTimeout defines the maximum time allowed between health checks before a node is considered
	// unresponsive. Nodes must send health checks at least once every half of this duration. Minimum time allowed
	// is 1s.
	ChecksTimeout time.Duration `mapstructure:"checks_timeout"`
}

func New() *Config {
	return &Config{
		NodeHealth: NodeHealth{
			ChecksBeforeRouting: DefaultNodeHealthChecksBeforeRouting,
			ChecksTimeout:       DefaultNodeHealthChecksTimeout,
		},
		Logger: logrus.New(),
	}
}

// InitConfig initializes the configuration for the command
func InitConfig(cmd *cobra.Command) *Config {
	cfg, err := common.LoadConfig(common.GetConfigFilePath(cmd), New())
	if err != nil {
		cfg = New()

		cfg.Logger.Warnf("failed to load default config: %v", err)
		cfg.Logger.Infof("relying on default configuration")
	}

	ApplyFlagsToConfig(cmd, cfg)

	cfg.Logger = logrus.New()

	return cfg
}

// AddConfigFlags defines the configuration flags for the command
func AddConfigFlags(cmd *cobra.Command) {
	cmd.Flags().Uint(KeyNodeHealthChecksBeforeRouting, DefaultNodeHealthChecksBeforeRouting, "Continuous node health checks that must be received before starting routing traffic to the node")
	cmd.Flags().Duration(KeyNodeHealthChecksTimeout, DefaultNodeHealthChecksTimeout, "Maximum time between health checks before node is considered unresponsive and traffic is re-routed")
	cmd.Flags().String(common.KeyConfigFile, DefaultConfigFile, "config file (default is $PWD/config/lb.toml)")

	_ = viper.BindPFlag(KeyNodeHealthChecksBeforeRouting, cmd.Flags().Lookup(KeyNodeHealthChecksBeforeRouting))
	_ = viper.BindPFlag(KeyNodeHealthChecksTimeout, cmd.Flags().Lookup(KeyNodeHealthChecksTimeout))
	_ = viper.BindPFlag(common.KeyConfigFile, cmd.Flags().Lookup(common.KeyConfigFile))
}

func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed(KeyNodeHealthChecksBeforeRouting) {
		cfg.NodeHealth.ChecksBeforeRouting = viper.GetUint(KeyNodeHealthChecksBeforeRouting)
	}
	if cmd.Flags().Changed(KeyNodeHealthChecksTimeout) {
		cfg.NodeHealth.ChecksTimeout = viper.GetDuration(KeyNodeHealthChecksTimeout)
	}
}
