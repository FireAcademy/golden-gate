package redis

import (
	"os"
	"time"
	"context"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	telemetry "github.com/fireacademy/telemetry"
	"github.com/redis/go-redis/extra/redisotel/v9"
)

const PROCESS_QUEUE_SET_NAME = "api-keys-to-update"
const USAGE_PREFIX = "usage-"
const ORIGIN_PREFIX = "origin-"
const API_KEY_PREFIX = "key-"
const API_KEY_OK_VALUE = "ok"
const API_KEY_PENDING_CHECK_VALUE = "pending"

var RDB *redis.Client

func BillCreditsQuickly(ctx context.Context, apiKey string, credits int64) error {
	ctx, span := telemetry.GetSpan(ctx, "BillCreditsQuickly")
	defer span.End()
	span.SetAttributes(
		attribute.String("api_key", apiKey),
		attribute.Int64("credits", credits),
	)
	
	err := RDB.IncrBy(ctx, USAGE_PREFIX + apiKey, credits).Err()
	if err != nil {
		telemetry.LogError(ctx, err, "could not increment usage for API key " + apiKey)
		return err
	}

	err = RDB.SAdd(ctx, PROCESS_QUEUE_SET_NAME, apiKey).Err()
	if err != nil {
		telemetry.LogError(ctx, err, "could not add API key to the job queue: " + apiKey)
		return err
	}

	return nil
}

func CheckAPIKeyQuickly(ctx context.Context, apiKey string) (bool /* ok */, string /* origin */, error /* error */) {
	ctx, span := telemetry.GetSpan(ctx, "CheckAPIKeyQuickly")
	defer span.End()
	span.SetAttributes(
		attribute.String("api_key", apiKey),
	)

	ok, err := RDB.Get(ctx, API_KEY_PREFIX + apiKey).Result()
	if err != nil {
		if err != redis.Nil {
			telemetry.LogError(ctx, err, "strange error")
		}
		return false, "", err
	}
	
	for ok == API_KEY_PENDING_CHECK_VALUE {
		time.Sleep(100 * time.Millisecond)
		ok, err = RDB.Get(ctx, API_KEY_PREFIX + apiKey).Result()
		if err != nil {
			if err != redis.Nil {
				telemetry.LogError(ctx, err, "strange error in for")
			}
			return false, "", err
		}
	}

	origin, err := RDB.Get(ctx, ORIGIN_PREFIX + apiKey).Result()
	if err == redis.Nil {
		origin = "*"
	} else if err != nil {
		telemetry.LogError(ctx, err, "strange error when getting origin")
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

	// https://redis.uptrace.dev/guide/go-redis-monitoring.html#opentelemetry-instrumentation
	// Enable tracing instrumentation.
	if err := redisotel.InstrumentTracing(RDB); err != nil {
		panic(err)
	}

	// Enable metrics instrumentation.
	if err := redisotel.InstrumentMetrics(RDB); err != nil {
		panic(err)
	}
}