package manager

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/WALL-EEEEEEE/proxy-service/manager/cache"
	"github.com/WALL-EEEEEEE/proxy-service/manager/config"
	"github.com/WALL-EEEEEEE/proxy-service/manager/repository"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/redis/go-redis/v9"
	redis_search "github.com/redis/rueidis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"

	"github.com/zput/zxcTool/ztLog/zt_formatter"
)

func SetupLog(fields_order ...string) {
	log.SetReportCaller(true)
	log.SetFormatter(&zt_formatter.ZtFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			var filename string = path.Base(f.File)
			dirs := strings.Split(path.Dir(f.File), string(os.PathSeparator))
			if len(dirs) > 0 {
				base_dir := dirs[len(dirs)-1]
				filename = strings.Join([]string{base_dir, filename}, string(os.PathSeparator))
			}
			return "", fmt.Sprintf("%s:%d", filename, f.Line)
		},
		Formatter: nested.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			ShowFullLevel:   true,
			FieldsOrder:     fields_order,
		},
	})
}

func SetupDb(conf *config.Config) (*gorm.DB, error) {
	//init db
	port := conf.Mysql.Port
	host := conf.Mysql.Host
	user := conf.Mysql.User
	password := conf.Mysql.Password
	database := conf.Mysql.Database
	if host == "" {
		return nil, fmt.Errorf("invalid mysql config (mysql.host required)")
	}
	if port == "" {
		return nil, fmt.Errorf("invalid mysql config (mysql.port required)")
	}
	if user == "" {
		return nil, fmt.Errorf("invalid mysql config (mysql.user required)")
	}
	if password == "" {
		return nil, fmt.Errorf("invalid mysql config (mysql.password required)")
	}
	if database == "" {
		return nil, fmt.Errorf("invalid mysql config (mysql.database required)")
	}
	create_db_dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	conn, err := gorm.Open(mysql.Open(create_db_dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	result := conn.Exec("CREATE DATABASE IF NOT EXISTS " + database + ";")
	if result.Error != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", user, password, host, port, database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&repository.ProxyProvider{}, &repository.ProxyApi{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func SetupRedis(conf *config.Config) (*redis.Client, error) {
	//init redis client for cache
	redis_addr := conf.Redis.Address
	if len(redis_addr) < 1 {
		return nil, fmt.Errorf("invalid redis config (redis.address required)")
	}
	redis_cli := cache.NewRedis(conf)
	return redis_cli, nil

}

func SetupRedisSearch(conf *config.Config) (redis_search.Client, error) {
	//init redis client for cache
	redis_addr := conf.Redis.Address
	if len(redis_addr) < 1 {
		return nil, fmt.Errorf("invalid redis config (redis.address required)")
	}
	redis_cli, err := cache.NewRedisSearch(conf)
	return redis_cli, err

}
