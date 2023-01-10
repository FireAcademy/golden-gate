package main

import (
	"os"
	"github.com/go-redis/redis/v8"
)

var DataDudeAPIKeyInfoURL string
var DataDudeBillCreditsPackageURL string
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
				log.Print(err)
				panic(err)
			}
		}

		_, err := RefreshAPIKey(apiKey)
		if err != nil {
			log.Print(err)
			err = RDB.SAdd(context.Background(), PROCESS_QUEUE_SET_NAME, apiKey).Err()
			if err != nil {
				log.Print(err)
			}
		}
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
	ManagementToken = getManagementToken()
	go loop()
}