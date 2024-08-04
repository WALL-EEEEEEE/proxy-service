package cmd

import "github.com/spf13/cobra"

var ManagerCmd = &cobra.Command{
	Use:   "manager",
	Short: "manager cli",
	Long:  `A crucial proxy service for spider, to maintain a  high quality ,high speed ip pool`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
