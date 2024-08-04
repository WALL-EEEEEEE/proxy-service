package main

import (
	"os"

	manager "github.com/WALL-EEEEEEE/proxy-service/manager"
	"github.com/WALL-EEEEEEE/proxy-service/manager/cmd"
)

func main() {
	manager.SetupLog()
	cli := cmd.ClientCmd
	if err := cli.Execute(); err != nil {
		cli.PrintErr(err)
		os.Exit(1)
	}
}
