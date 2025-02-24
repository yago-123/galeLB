package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	nodeConfig "github.com/yago-123/galelb/config/node"
)

var rootCmd = &cobra.Command{
	Use: "gale-node",
	Run: func(cmd *cobra.Command, _ []string) {
		cfg = nodeConfig.InitConfig(cmd)
	},
}

func Execute(logger *logrus.Logger) {
	nodeConfig.AddConfigFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error executing command: %v", err)
	}
}
