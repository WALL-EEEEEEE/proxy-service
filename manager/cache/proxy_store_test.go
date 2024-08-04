package cache

import (
	"context"
	"testing"

	"github.com/WALL-EEEEEEE/proxy-service/manager/model"

	"github.com/WALL-EEEEEEE/Axiom/test"
	"github.com/redis/go-redis/v9"
)

func TestProxyStore(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0, // use default DB
	})
	ctx := context.Background()
	cases := []test.TestCase[any, any]{
		{
			Name:     "ProxyStore.Filter.HappyPath",
			Input:    nil,
			Error:    nil,
			Expected: nil,
			Check: func(tc test.TestCase[any, any]) {
				var proxy model.Proxy
				store := NewProxyStore(client)
				t.Log(store.GetByKey(ctx, "127.0.0.1", &proxy))
			},
		},
		{
			Name:     "ProxyStore.Set.HappyPath",
			Input:    nil,
			Error:    nil,
			Expected: nil,
			Check: func(tc test.TestCase[any, any]) {
			},
		},
		{
			Name:     "ProxyStore.Get.HappyPath",
			Input:    nil,
			Error:    nil,
			Expected: nil,
			Check: func(tc test.TestCase[any, any]) {
				var proxy model.Proxy
				store := NewProxyStore(client)
				t.Log(store.GetByKey(ctx, "127.0.0.1", &proxy))
			},
		},
	}
	test.Run(cases, t)
}
