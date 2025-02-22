package main

import (
	lbConfig "github.com/yago-123/galelb/config/lb"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "gale-lb",
	Run: func(cmd *cobra.Command, _ []string) {
		cfg = lbConfig.InitConfig(cmd)
	},
}

func Execute(logger *logrus.Logger) {
	lbConfig.AddConfigFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("error executing command: %v", err)
	}
}
