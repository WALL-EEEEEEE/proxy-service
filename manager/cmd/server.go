package cmd

import (
	common "github.com/WALL-EEEEEEE/proxy-service/common"

	manager "github.com/WALL-EEEEEEE/proxy-service/manager"
	conf "github.com/WALL-EEEEEEE/proxy-service/manager/config"
	log "github.com/sirupsen/logrus"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var (
	logger      = log.StandardLogger()
	loglevel    string
	config_file string
	http_port   string
	grpc_port   string

	ServerCmd = &cobra.Command{
		Use:   "server",
		Short: "start server for manager service",
		Run: func(cmd *cobra.Command, args []string) {
			var conf conf.Config
			common.SetLevel(loglevel)

			var err error
			if config_file != "" {
				err = common.ParseConfig(config_file, &conf)

			} else {
				err = envconfig.Process("", &conf)
			}
			if err != nil {
				cmd.PrintErrln(err)
				return
			}
			manager.StartServer(&conf, grpc_port, http_port)
		},
	}
)

func init() {

	ServerCmd.Flags().StringVarP(&loglevel, "log", "l", "INFO", "log level")
	ServerCmd.Flags().StringVarP(&config_file, "conf", "c", "", "config file")
	ServerCmd.Flags().StringVarP(&grpc_port, "grpc", "", "8082", "grpc port")
	ServerCmd.Flags().StringVarP(&http_port, "http", "", "8083", "http port")
}
