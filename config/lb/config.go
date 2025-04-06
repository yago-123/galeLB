package lb

import (
	"time"

	"github.com/sirupsen/logrus"
	common "github.com/yago-123/galelb/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// Private network options
	KeyPrivateNodePort        = "private_interface.node_port"
	KeyPrivateAPIPort         = "private_interface.api_port"
	KeyPrivateNetIfacePrivate = "private_interface.net_interface_private"

	// Public network options
	KeyPublicClientsPort    = "public_interface.clients_port"
	KeyPublicNetIfacePublic = "public_interface.net_interface_public"

	// Node health options
	KeyNodeHealthChecksBeforeRouting = "node_health.checks_before_routing"
	KeyNodeHealthChecksTimeout       = "node_health.checks_timeout"
	KeyNodeHealthBlackListAfterFails = "node_health.black_list_after_fails"
	KeyNodeHealthBlackListExpiry     = "node_health.black_list_expiry"
)

const (
	// Private network options
	DefaultPrivateNodePort        = 7070
	DefaultPrivateAPIPort         = 5555
	DefaultPrivateNetIfacePrivate = ""

	// Public network options
	DefaultPublicClientsPort    = 8080
	DefaultPublicNetIfacePublic = ""

	DefaultNodeHealthChecksBeforeRouting = 3
	DefaultNodeHealthChecksTimeout       = 10 * time.Second
	DefaultNodeHealthBlackListAfterFails = -1
	DefaultNodeHealthBlackListExpiry     = 60 * time.Second

	DefaultConfigFile = "lb.toml"
)

type Config struct {
	PrivateInterface PrivateInterface `mapstructure:"private_interface"`
	PublicInterface  PublicInterface  `mapstructure:"public_interface"`
	NodeHealth       NodeHealth       `mapstructure:"node_health"`
	Logger           *logrus.Logger
}

type PrivateInterface struct {
	// NodePort is the port that will be used by nodes to communicate with LB
	NodePort int `mapstructure:"node_port"`
	// APIPort is the port that will receive and forward client requests to the nodes
	APIPort int `mapstructure:"api_port"`
	// NetIfacePrivate is the network interface that will be used to retrieve and route packets to nodes
	NetIfacePrivate string `mapstructure:"net_interface_private"`
}

type PublicInterface struct {
	// ClientsPort is the port that will receive and forward client requests to the nodes
	ClientsPort int `mapstructure:"clients_port"`
	// NetIfacePublic is the network interface that will be used to retrieve and route client packets. This variable
	// is required to be set in order to load the XDP program
	NetIfacePublic string `mapstructure:"net_interface_public"`
}

type NodeHealth struct {
	// ChecksBeforeRouting number of continuous health checks passed before node is added to routing ring.
	ChecksBeforeRouting uint `mapstructure:"checks_before_routing"`
	// ChecksTimeout defines the maximum time allowed between health checks before a node is considered
	// unresponsive. Nodes must send health checks at least once every half of this duration. Minimum time
	// allowed is 1s.
	ChecksTimeout time.Duration `mapstructure:"checks_timeout"`
	// BlackListAfterFails number of times a node can be a added and disabled from the routing table before it is
	// added into the ignore list. By default is disabled
	BlackListAfterFails int `mapstructure:"black_list_after_fails"`
	// BlackListExpiry represents duration of ban after which black listed nodes will be accepted again
	BlackListExpiry time.Duration `mapstructure:"black_list_expiry"`
}

func New() *Config {
	return &Config{
		PrivateInterface: PrivateInterface{
			NodePort:        DefaultPrivateNodePort,
			APIPort:         DefaultPrivateAPIPort,
			NetIfacePrivate: DefaultPrivateNetIfacePrivate,
		},
		PublicInterface: PublicInterface{
			ClientsPort:    DefaultPublicClientsPort,
			NetIfacePublic: DefaultPublicNetIfacePublic,
		},
		NodeHealth: NodeHealth{
			ChecksBeforeRouting: DefaultNodeHealthChecksBeforeRouting,
			ChecksTimeout:       DefaultNodeHealthChecksTimeout,
			BlackListAfterFails: DefaultNodeHealthBlackListAfterFails,
			BlackListExpiry:     DefaultNodeHealthBlackListExpiry,
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
	cmd.Flags().Int(KeyPrivateNodePort, DefaultPrivateNodePort, "Port that will be used by nodes to communicate with LB")
	cmd.Flags().Int(KeyPrivateAPIPort, DefaultPrivateAPIPort, "Port that will receive and forward client requests to the nodes")
	cmd.Flags().Int(KeyPublicClientsPort, DefaultPublicClientsPort, "Port that will receive and forward client requests to the nodes")
	cmd.Flags().String(KeyPrivateNetIfacePrivate, DefaultPrivateNetIfacePrivate, "Network interface that will be used to retrieve and route packets to nodes")
	cmd.Flags().String(KeyPublicNetIfacePublic, DefaultPublicNetIfacePublic, "Network interface that will be used to retrieve and route client packets")
	cmd.Flags().Uint(KeyNodeHealthChecksBeforeRouting, DefaultNodeHealthChecksBeforeRouting, "Continuous node health checks that must be received before starting routing traffic to the node")
	cmd.Flags().Duration(KeyNodeHealthChecksTimeout, DefaultNodeHealthChecksTimeout, "Maximum time between health checks before node is considered unresponsive and traffic is re-routed")
	cmd.Flags().Int(KeyNodeHealthBlackListAfterFails, DefaultNodeHealthBlackListAfterFails, "Number of times node can be added and disabled from routing table before is ignored by load balancer")
	cmd.Flags().Duration(KeyNodeHealthBlackListExpiry, DefaultNodeHealthBlackListExpiry, "Duration of the black list ban after which the node will be accepted again")
	cmd.Flags().String(common.KeyConfigFile, DefaultConfigFile, "config file (default is $PWD/config/lb.toml)")

	_ = viper.BindPFlag(KeyPrivateNodePort, cmd.Flags().Lookup(KeyPrivateNodePort))
	_ = viper.BindPFlag(KeyPrivateAPIPort, cmd.Flags().Lookup(KeyPrivateAPIPort))
	_ = viper.BindPFlag(KeyPublicClientsPort, cmd.Flags().Lookup(KeyPublicClientsPort))
	_ = viper.BindPFlag(KeyPrivateNetIfacePrivate, cmd.Flags().Lookup(KeyPrivateNetIfacePrivate))
	_ = viper.BindPFlag(KeyPublicNetIfacePublic, cmd.Flags().Lookup(KeyPublicNetIfacePublic))
	_ = viper.BindPFlag(KeyNodeHealthChecksBeforeRouting, cmd.Flags().Lookup(KeyNodeHealthChecksBeforeRouting))
	_ = viper.BindPFlag(KeyNodeHealthChecksTimeout, cmd.Flags().Lookup(KeyNodeHealthChecksTimeout))
	_ = viper.BindPFlag(KeyNodeHealthBlackListAfterFails, cmd.Flags().Lookup(KeyNodeHealthBlackListAfterFails))
	_ = viper.BindPFlag(KeyNodeHealthBlackListExpiry, cmd.Flags().Lookup(KeyNodeHealthBlackListExpiry))
	_ = viper.BindPFlag(common.KeyConfigFile, cmd.Flags().Lookup(common.KeyConfigFile))
}

func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed(KeyPrivateNodePort) {
		cfg.PrivateInterface.NodePort = viper.GetInt(KeyPrivateNodePort)
	}
	if cmd.Flags().Changed(KeyPrivateAPIPort) {
		cfg.PrivateInterface.APIPort = viper.GetInt(KeyPrivateAPIPort)
	}
	if cmd.Flags().Changed(KeyPrivateNetIfacePrivate) {
		cfg.PrivateInterface.NetIfacePrivate = viper.GetString(KeyPrivateNetIfacePrivate)
	}
	if cmd.Flags().Changed(KeyPublicClientsPort) {
		cfg.PublicInterface.ClientsPort = viper.GetInt(KeyPublicClientsPort)
	}
	if cmd.Flags().Changed(KeyPublicNetIfacePublic) {
		cfg.PublicInterface.NetIfacePublic = viper.GetString(KeyPublicNetIfacePublic)
	}
	if cmd.Flags().Changed(KeyNodeHealthChecksBeforeRouting) {
		cfg.NodeHealth.ChecksBeforeRouting = viper.GetUint(KeyNodeHealthChecksBeforeRouting)
	}
	if cmd.Flags().Changed(KeyNodeHealthChecksTimeout) {
		cfg.NodeHealth.ChecksTimeout = viper.GetDuration(KeyNodeHealthChecksTimeout)
	}
	if cmd.Flags().Changed(KeyNodeHealthBlackListAfterFails) {
		cfg.NodeHealth.BlackListAfterFails = viper.GetInt(KeyNodeHealthBlackListAfterFails)
	}
	if cmd.Flags().Changed(KeyNodeHealthBlackListExpiry) {
		cfg.NodeHealth.BlackListExpiry = viper.GetDuration(KeyNodeHealthBlackListExpiry)
	}
}
