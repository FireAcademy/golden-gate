package main

import (
	"log"
	"context"
	"github.com/go-redis/redis/v8"
)

const PROCESS_QUEUE_SET_NAME = "api-keys-to-update"
const USAGE_PREFIX = "usage-"
const ORIGIN_PREFIX = "origin-"

var RDB *redis.Client

func BillCreditsQuickly(apiKey string, credits int64) error {
	err := RDB.IncrBy(context.Background(), USAGE_PREFIX + apiKey, credits).Err()
	if err != nil {
		log.Print(err)
		return err
	}

	err = RDB.SAdd(context.Background(), PROCESS_QUEUE_SET_NAME, apiKey).Err()
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

func checkAPIKeyQuickly(apiKey string) (bool /* ok */, string /* origin */, error /* errored */) {
	ok, err := rdb.Get(context.Background(), apiKey).Result()
	if err != nil {
		if err != redis.Nil {
			log.Print(err)
		}
		return false, "", err
	}

	origin, err := rdb.Get(context.Background(), ORIGIN_PREFIX + apiKey).Result()
	if err != nil {
		if err != redis.Nil {
			log.Print(err)
		}
		return false, "", err
	}

	return ok, origin, nil
}

func getRedisConnectionString() string {
	url := os.Getenv("REDIS_CONNECTION_STRING")
    if url == "" {
       panic("REDIS_CONNECTION_STRING not set.")
    }

    return url
}

func SetupRedis() {
	ctx := context.Background()

	opt, err := redis.ParseURL(getRedisConnectionString())
	if err != nil {
		panic(err)
	}

	rdb = redis.NewClient(opt)
}