package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tfstoys",
	Short: "Useful tools for remote terraform state on S3",
	Long: `Useful tools for remote terraform state on S3
	Currently only diff'ing is supported. Stay tuned!
	View more at: github.com/kim0/tfstoys`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
