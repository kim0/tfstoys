package main

import (
	"github.com/kim0/tfstoys/pkgs/app"
	"github.com/spf13/cobra"
)

func init() {
	diffCmd.Flags().UintVarP(&days, "days", "d", 0, `Number of days to consider as gap. Used to compare most recent versions
	before and after the gap. The default of zero shows interactive version picker`)
	diffCmd.Flags().StringVarP(&state_bucket, "bucket", "b", "brave-devops-remote-state-production", "Name of the S3 remote state bucket")
	diffCmd.Flags().StringVarP(&state_path, "path", "p", "", "ex: website/foo.com/production. If undefined, shows interactive picker")
	rootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Performs diff operation on two TF state versions",
	Long:  `Diff two TF state versions stored on S3`,
	Run: func(cmd *cobra.Command, args []string) {
		since_strategy := app.Since_Strategy{Days: days}
		app.Diff(state_bucket, state_path, since_strategy)
	},
}
