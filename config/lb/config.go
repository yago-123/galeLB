package lb

import (
	"time"

	"github.com/sirupsen/logrus"
	common "github.com/yago-123/galelb/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	KeyLocalNodePort                 = "local.node_port"
	KeyLocalClientsPort              = "local.clients_port"
	KeyLocalNetIfaceNodes            = "local.net_interface_nodes"
	KeyLocalNetIfaceClients          = "local.net_interface_clients"
	KeyNodeHealthChecksBeforeRouting = "node_health.checks_before_routing"
	KeyNodeHealthChecksTimeout       = "node_health.checks_timeout"
	KeyNodeHealthBlackListAfterFails = "node_health.black_list_after_fails"
	KeyNodeHealthBlackListExpiry     = "node_health.black_list_expiry"
)

const (
	DefaultLocalNodePort                 = 7070
	DefaultLocalClientsPort              = 8080
	DefaultLocalNetIfaceNodes            = ""
	DefaultLocalNetIfaceClients          = ""
	DefaultNodeHealthChecksBeforeRouting = 3
	DefaultNodeHealthChecksTimeout       = 10 * time.Second
	DefaultNodeHealthBlackListAfterFails = -1
	DefaultNodeHealthBlackListExpiry     = 60 * time.Second

	DefaultConfigFile = "lb.toml"
)

type Config struct {
	Local      Local      `mapstructure:"local"`
	NodeHealth NodeHealth `mapstructure:"node_health"`
	Logger     *logrus.Logger
}

// Local contains the configuration used to determine characteristics that will be used by the local entity such
// as ports that will be available for other components
type Local struct {
	// NodePort is the port that will be used by nodes to communicate with LB
	NodePort int `mapstructure:"node_port"`
	// ClientsPort is the port that will receive and forward client requests to the nodes
	ClientsPort int `mapstructure:"clients_port"`
	// NetIfaceNodes is the network interface that will be used to retrieve and route packets to nodes
	NetIfaceNodes string `mapstructure:"net_interface_nodes"`
	// NetIfaceClients is the network interface that will be used to retrieve and route client packets. This variable
	// is required to be set in order to load the XDP program
	NetIfaceClients string `mapstructure:"net_interface_clients"`
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
		Local: Local{
			NodePort:        DefaultLocalNodePort,
			ClientsPort:     DefaultLocalClientsPort,
			NetIfaceNodes:   DefaultLocalNetIfaceNodes,
			NetIfaceClients: DefaultLocalNetIfaceClients,
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
	cmd.Flags().Int(KeyLocalNodePort, DefaultLocalNodePort, "Port that will be used by nodes to communicate with LB")
	cmd.Flags().Int(KeyLocalClientsPort, DefaultLocalClientsPort, "Port that will receive and forward client requests to the nodes")
	cmd.Flags().String(KeyLocalNetIfaceNodes, DefaultLocalNetIfaceNodes, "Network interface that will be used to retrieve and route packets to nodes")
	cmd.Flags().String(KeyLocalNetIfaceClients, DefaultLocalNetIfaceClients, "Network interface that will be used to retrieve and route client packets")
	cmd.Flags().Uint(KeyNodeHealthChecksBeforeRouting, DefaultNodeHealthChecksBeforeRouting, "Continuous node health checks that must be received before starting routing traffic to the node")
	cmd.Flags().Duration(KeyNodeHealthChecksTimeout, DefaultNodeHealthChecksTimeout, "Maximum time between health checks before node is considered unresponsive and traffic is re-routed")
	cmd.Flags().Int(KeyNodeHealthBlackListAfterFails, DefaultNodeHealthBlackListAfterFails, "Number of times node can be added and disabled from routing table before is ignored by load balancer")
	cmd.Flags().Duration(KeyNodeHealthBlackListExpiry, DefaultNodeHealthBlackListExpiry, "Duration of the black list ban after which the node will be accepted again")
	cmd.Flags().String(common.KeyConfigFile, DefaultConfigFile, "config file (default is $PWD/config/lb.toml)")

	_ = viper.BindPFlag(KeyLocalNodePort, cmd.Flags().Lookup(KeyLocalNodePort))
	_ = viper.BindPFlag(KeyLocalClientsPort, cmd.Flags().Lookup(KeyLocalClientsPort))
	_ = viper.BindPFlag(KeyLocalNetIfaceNodes, cmd.Flags().Lookup(KeyLocalNetIfaceNodes))
	_ = viper.BindPFlag(KeyLocalNetIfaceClients, cmd.Flags().Lookup(KeyLocalNetIfaceClients))
	_ = viper.BindPFlag(KeyNodeHealthChecksBeforeRouting, cmd.Flags().Lookup(KeyNodeHealthChecksBeforeRouting))
	_ = viper.BindPFlag(KeyNodeHealthChecksTimeout, cmd.Flags().Lookup(KeyNodeHealthChecksTimeout))
	_ = viper.BindPFlag(KeyNodeHealthBlackListAfterFails, cmd.Flags().Lookup(KeyNodeHealthBlackListAfterFails))
	_ = viper.BindPFlag(KeyNodeHealthBlackListExpiry, cmd.Flags().Lookup(KeyNodeHealthBlackListExpiry))
	_ = viper.BindPFlag(common.KeyConfigFile, cmd.Flags().Lookup(common.KeyConfigFile))
}

func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed(KeyLocalNodePort) {
		cfg.Local.NodePort = viper.GetInt(KeyLocalNodePort)
	}
	if cmd.Flags().Changed(KeyLocalClientsPort) {
		cfg.Local.ClientsPort = viper.GetInt(KeyLocalClientsPort)
	}
	if cmd.Flags().Changed(KeyLocalNetIfaceNodes) {
		cfg.Local.NetIfaceNodes = viper.GetString(KeyLocalNetIfaceNodes)
	}
	if cmd.Flags().Changed(KeyLocalNetIfaceClients) {
		cfg.Local.NetIfaceClients = viper.GetString(KeyLocalNetIfaceClients)
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
