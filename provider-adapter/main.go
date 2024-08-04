package main

import (
	"os"
	"path"

	common "github.com/WALL-EEEEEEE/proxy-service/common"

	conf "github.com/WALL-EEEEEEE/proxy-service/provider-adapter/config"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var logger = log.StandardLogger()

var (
	config_file string
	http_port   string
	grpc_port   string
	loglevel    string

	rootCmd = &cobra.Command{
		Use:   "provider-adapter",
		Short: "start adapter api server for proxy provider",
		Run: func(cmd *cobra.Command, args []string) {
			work_dir, err := os.Getwd()
			common.SetLevel(loglevel)
			if err != nil {
				cmd.PrintErrln(err)
				return
			}
			var conf conf.Config
			if config_file != "" {
				config_file := path.Join(work_dir, config_file)
				err = common.ParseConfig(config_file, &conf)

			} else {
				err = envconfig.Process("", &conf)
			}
			if err != nil {
				cmd.PrintErrln(err)
				return
			}
			StartServer(logger, &conf, grpc_port, http_port)
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&config_file, "conf", "c", "", "config file")
	rootCmd.Flags().StringVarP(&http_port, "http", "", "8083", "http port")
	rootCmd.Flags().StringVarP(&grpc_port, "grpc", "", "8082", "grpc port")
	rootCmd.Flags().StringVarP(&loglevel, "log", "l", "INFO", "log level")

}

func main() {
	common.SetupLog()
	if err := rootCmd.Execute(); err != nil {
		rootCmd.PrintErr(err)
		os.Exit(1)
	}
}
