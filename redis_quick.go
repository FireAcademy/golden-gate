package main

import (
	"os"
	"log"
	"context"
	"github.com/go-redis/redis/v8"
)

const PROCESS_QUEUE_SET_NAME = "api-keys-to-update"
const USAGE_PREFIX = "usage-"
const ORIGIN_PREFIX = "origin-"
const API_KEY_OK_VALUE = "ok"

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
	ok, err := RDB.Get(context.Background(), apiKey).Result()
	if err != nil {
		if err != redis.Nil {
			log.Print(err)
		}
		return false, "", err
	}

	origin, err := RDB.Get(context.Background(), ORIGIN_PREFIX + apiKey).Result()
	if err == redis.Nil {
		origin = "*"
	} else if err != nil {
		log.Print(err)
		return false, "", err
	}

	return ok == API_KEY_OK_VALUE, origin, nil
}

func getRedisConnectionString() string {
	url := os.Getenv("REDIS_CONNECTION_STRING")
    if url == "" {
       panic("REDIS_CONNECTION_STRING not set.")
    }

    return url
}

func SetupRedis() {
	opt, err := redis.ParseURL(getRedisConnectionString())
	if err != nil {
		panic(err)
	}

	RDB = redis.NewClient(opt)
}