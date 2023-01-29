package main

import (
	"os"
	"time"
	"context"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	. "github.com/fireacademy/golden-gate/redis"
)

var DataDudeAPIKeyInfoURL string
var DataDudeBillCreditsPackageURL string
var DataDudeRecordUsageURL string
var ManagementToken string

func loop() {
	time.Sleep(5 * time.Second)
	for true {
		apiKey, err := RDB.SPop(context.Background(), PROCESS_QUEUE_SET_NAME).Result()
		if err != nil {
			if err == redis.Nil {
				time.Sleep(time.Second)
				continue
			} else {
				ctx, span := GetSpan(context.Background(), "loop")
				defer span.End()
				LogError(ctx, err, "error when popping item from API kyes queue")
				panic(err)
			}
		}
		ctx, span := GetSpan(context.Background(), "loop")
		span.SetAttributes(
			attribute.String("api_key", apiKey),
		)

		_, _, err = RefreshAPIKey(ctx, apiKey)
		if err != nil {
			LogError(ctx, err, "error while refreshing API key")
			err = RDB.SAdd(ctx, PROCESS_QUEUE_SET_NAME, apiKey).Err()
			if err != nil {
				LogError(ctx, err, "error while adding API key back to queue")
			}
		}

		span.End()
	}
}

func getDataDudeApiKeyInfoURL() string {
   url := os.Getenv("DATA_DUDE_API_KEY_INFO_URL")
   if url == "" {
       panic("DATA_DUDE_API_KEY_INFO_URL not set.")
   }

   return url
}

func getDataDudeBillCreditsPackageURL() string {
   url := os.Getenv("DATA_DUDE_BILL_CREDITS_PACKAGE_URL")
   if url == "" {
       panic("DATA_DUDE_BILL_CREDITS_PACKAGE_URL not set.")
   }

   return url
}

func getDataDudeRecordUsageURL() string {
   url := os.Getenv("DATA_DUDE_RECORD_USAGE_URL")
   if url == "" {
       panic("DATA_DUDE_RECORD_USAGE_URL not set.")
   }

   return url
}

func getManagementToken() string {
   token := os.Getenv("DATA_DUDE_MANAGEMENT_TOKEN")
   if token == "" {
       panic("DATA_DUDE_MANAGEMENT_TOKEN not set.")
   }

   return token
}

func SetupCheck() {
	DataDudeAPIKeyInfoURL = getDataDudeApiKeyInfoURL()
	DataDudeBillCreditsPackageURL = getDataDudeBillCreditsPackageURL()
	DataDudeRecordUsageURL = getDataDudeRecordUsageURL()
	ManagementToken = getManagementToken()
	go loop()
}