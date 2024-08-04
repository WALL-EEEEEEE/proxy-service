package cmd

import (
	"github.com/spf13/cobra"
)

var AssessCmd = &cobra.Command{
	Use:   "assess",
	Short: "Assess all ip provides' quality and give the detail report",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
