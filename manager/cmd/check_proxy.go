package cmd

import (
	"context"

	common "github.com/WALL-EEEEEEE/proxy-service/common"

	manager "github.com/WALL-EEEEEEE/proxy-service/manager"
	conf "github.com/WALL-EEEEEEE/proxy-service/manager/config"
	"github.com/WALL-EEEEEEE/proxy-service/manager/job"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
)

var (
	CheckProxyCmd = &cobra.Command{
		Use:   "check_proxy",
		Short: "start a job for checking proxy",
		Run: func(cmd *cobra.Command, args []string) {
			var conf conf.Config
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
			cmd.Printf("Config: %+v\n", conf)
			redis, err := manager.SetupRedis(&conf)
			if err != nil {
				cmd.PrintErr(err)
				return
			}
			ctx := context.Background()
			ret := redis.Ping(ctx)
			if ret.Err() != nil {
				cmd.PrintErrf("can't connect to redis server: %s", ret.Err())
				return
			}
			uid := uuid.New().String()
			proxy_check_job, err := job.NewProxyCheckJob(ctx, uid, manager_api, redis)
			if err != nil {
				cmd.PrintErr(err)
			}
			proxy_check_job.Start()
			proxy_check_job.Join()
		},
	}
)

func init() {
	CheckProxyCmd.Flags().StringVarP(&config_file, "conf", "c", "", "config file")
	CheckProxyCmd.Flags().StringVarP(&manager_api, "manager-api", "s", "", "manager api address, grpc")
	CheckProxyCmd.MarkFlagRequired("manager-api")
}
