package cache

import (
	"github.com/WALL-EEEEEEE/proxy-service/manager/config"

	"github.com/redis/go-redis/v9"
	redis_search "github.com/redis/rueidis"
)

func NewRedis(config *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.Password,
		DB:       0, // use default DB
	})
	return client
}

func NewRedisSearch(config *config.Config) (redis_search.Client, error) {
	client, err := redis_search.NewClient(redis_search.ClientOption{
		InitAddress: []string{config.Redis.Address},
		Password:    config.Redis.Password,
		SelectDB:    0, // use default DB
	})
	return client, err

}
