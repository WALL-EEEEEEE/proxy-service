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
	manager_api string
	CheckApiCmd = &cobra.Command{
		Use:   "check_api",
		Short: "start a job for retrieving proxy from api at  intervals",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
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
			db, err := manager.SetupDb(&conf)
			if err != nil {
				cmd.PrintErr(err)
				return
			}
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
			proxy_api_check_job, err := job.NewProxyApiCheckJob(ctx, uid, manager_api, redis, db)
			if err != nil {
				cmd.PrintErrf("failed to initialized api check job: %s", err.Error())
				return
			}
			proxy_api_check_job.Start()
			proxy_api_check_job.Join()
		},
	}
)

func init() {
	CheckApiCmd.Flags().StringVarP(&config_file, "conf", "c", "", "config file")
	CheckApiCmd.Flags().StringVarP(&manager_api, "manager-api", "s", "", "manager api address, grpc")
	CheckApiCmd.MarkFlagRequired("manager-api")
}
