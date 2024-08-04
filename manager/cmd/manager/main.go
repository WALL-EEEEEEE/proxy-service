package main

import (
	"os"

	manager "github.com/WALL-EEEEEEE/proxy-service/manager"
	"github.com/WALL-EEEEEEE/proxy-service/manager/cmd"
)

func main() {
	manager.SetupLog()
	cli := cmd.ManagerCmd
	cli.AddCommand(cmd.ServerCmd)
	cli.AddCommand(cmd.ClientCmd)
	cli.AddCommand(cmd.AssessCmd)
	cli.AddCommand(cmd.CheckApiCmd)
	cli.AddCommand(cmd.CheckProxyCmd)
	if err := cli.Execute(); err != nil {
		cli.PrintErr(err)
		os.Exit(1)
	}
}
