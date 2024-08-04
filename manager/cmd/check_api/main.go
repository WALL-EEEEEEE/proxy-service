package main

import (
	"os"

	manager "github.com/WALL-EEEEEEE/proxy-service/manager"
	"github.com/WALL-EEEEEEE/proxy-service/manager/cmd"
)

func main() {
	manager.SetupLog([]string{"task"}...)
	cli := cmd.CheckApiCmd
	if err := cli.Execute(); err != nil {
		cli.PrintErr(err)
		os.Exit(1)
	}
}
