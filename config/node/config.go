package node

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	common "github.com/yago-123/galelb/config"
	"strings"
)

const (
	KeyLoadBalancerAddresses = "load_balancer.addresses"

	DefaultConfigFile = "node.toml"
)

const (
	// LoadBalancerAddressArguments is the number of arguments expected for the load balancer address configuration
	// so far we only use the ip and port
	LoadBalancerAddressArguments = 2
)

var (
	DefaultLoadBalancerAddresses = []Address{}
)

type Config struct {
	LoadBalancer LoadBalancer `mapstructure:"load_balancer"`

	Logger *logrus.Logger
}

// LoadBalancer contains the configuration for the remote lbs
type LoadBalancer struct {
	Addresses []Address `mapstructure:"addresses"`
}

// Address represents an individual address entry in the TOML
type Address struct {
	IP   string `mapstructure:"ip"`
	Port int    `mapstructure:"port"`
}

func New() *Config {
	return &Config{
		LoadBalancer: LoadBalancer{
			Addresses: DefaultLoadBalancerAddresses,
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

func AddConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringArray(KeyLoadBalancerAddresses, []string{}, "Node seeds used to synchronize during startup")
	cmd.Flags().String(common.KeyConfigFile, DefaultConfigFile, "config file (default is $PWD/config/lb.toml)")

	_ = viper.BindPFlag(KeyLoadBalancerAddresses, cmd.Flags().Lookup(KeyLoadBalancerAddresses))
	_ = viper.BindPFlag(common.KeyConfigFile, cmd.Flags().Lookup(common.KeyConfigFile))
}

func ApplyFlagsToConfig(cmd *cobra.Command, cfg *Config) {
	if cmd.Flags().Changed(KeyLoadBalancerAddresses) {
		addrs, err := parseLBAddresses(viper.GetStringSlice(KeyLoadBalancerAddresses))
		if err != nil {
			cfg.Logger.Fatalf("failed to parse load balancer addresses: %v", err)
		}

		cfg.LoadBalancer.Addresses = addrs
	}
}

func parseLBAddresses(addrsStr []string) ([]Address, error) {
	var addrs []Address

	for idx, addr := range addrsStr {
		parts := strings.SplitN(addr, ":", LoadBalancerAddressArguments)
		if len(parts) != LoadBalancerAddressArguments {
			return nil, fmt.Errorf("invalid load balancer address at index %d: %s", idx, addr)
		}

		addrs = append(addrs, Address{
			IP:   parts[0],
			Port: viper.GetInt(parts[1]),
		})
	}

	return addrs, nil
}
