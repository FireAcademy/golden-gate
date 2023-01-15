package main

import (
	"log"
	"time"
	"context"
	"strconv"
	"github.com/go-redis/redis/v8"
	. "github.com/fireacademy/golden-gate/redis"
)

func IsAPIKeyOK(apiKey string, info DataDudeResponse, creditsToProcess int64) (int64 /* creditsToBill */, bool /* ok */, string /* newOrigin */, bool /* purchaseCreditsPackage */) {
	origin := info.APIKey.Origin
	creditsToBill := creditsToProcess
	ok := true

	if info.APIKey.MonthlyCreditLimit != 0 {
		if info.APIKey.MonthlyCreditLimit < info.Usage.Credits + creditsToBill {
			log.Print("Missed a few credits on API key " + apiKey)
			creditsToBill = info.APIKey.MonthlyCreditLimit - info.Usage.Credits
			ok = false
		}
	}

	if creditsToBill == 0 || info.APIKey.Disabled {
		return 0, false, origin, false
	}

	if creditsToBill > info.User.RemainingCredits {
		if info.User.StripeCustomerID.Valid && info.User.AutoPurchaseCreditsPackages {
			return creditsToBill, ok, origin, true
		} else {
			return info.User.RemainingCredits, false, origin, false
		}
	}

	return creditsToBill, ok, origin, false
}

func RefreshAPIKey(apiKey string) (bool /* canBeUsed */, string /* origin */, error /* err */) {
	state, err := RDB.Get(context.Background(), apiKey).Result()
	if err == nil && state == API_KEY_PENDING_CHECK_VALUE {
		return CheckAPIKeyQuickly(apiKey)
	}

	err = RDB.Set(context.Background(), apiKey, API_KEY_PENDING_CHECK_VALUE, 2 * time.Second).Err()
	if err != nil {
		log.Print(err)
		return false, "", err
	}

	info, err := GetAPIKeyInfoFromDataDude(apiKey)
	if err != nil {
		return false, "", err
	}

	creditsToProcessStr, err := RDB.Get(context.Background(), USAGE_PREFIX + apiKey).Result()
	if err != nil {
		if err != redis.Nil {
			log.Print(err)
		}
		creditsToProcessStr = "0"
	}
	creditsToProcess, err := strconv.ParseInt(creditsToProcessStr, 10, 64)
	if err != nil {
		log.Print(err)
		creditsToProcess = 0
	}

	creditsToBill, ok, newOrigin, purchaseCreditsPackage := IsAPIKeyOK(apiKey, info, creditsToProcess)

	if purchaseCreditsPackage {
		success, err := TellDataDudeToBillCreditsPackage(info.User.StripeCustomerID.String)

		if err != nil || !success {
			if err != nil {
				log.Print(err)
			}
			log.Print("Could not bill extra package for " + info.User.StripeCustomerID.String)
		}
	}

	if creditsToBill > 0 {
		success, err := TellDataDudeToRecordUsage(apiKey, creditsToBill)
		if err != nil || !success {
			log.Print("Could not bill credits for " + apiKey)
			if err != nil {
				log.Print(err)
			}
		}
	}

	// update redis
	valueToSet := "not-ok"
	if ok {
		valueToSet = API_KEY_OK_VALUE
	}
	err = RDB.Set(context.Background(), apiKey, valueToSet, 0).Err()
	if err != nil {
		log.Print(err)
		return true, newOrigin, err
	}

	if ok {
		err = RDB.Set(context.Background(), ORIGIN_PREFIX + apiKey, newOrigin, 0).Err()
		if err != nil {
			log.Print(err)
			return true, newOrigin, err
		}
		err = RDB.DecrBy(context.Background(), USAGE_PREFIX + apiKey, creditsToProcess).Err()
		if err != nil {
			log.Print(err)
			return true, newOrigin, err
		}
	} else {
		err = RDB.Del(context.Background(), ORIGIN_PREFIX + apiKey).Err()
		if err != nil && err != redis.Nil {
			log.Print(err)
			return false, newOrigin, err
		}
		err = RDB.Del(context.Background(), USAGE_PREFIX + apiKey).Err()
		if err != nil && err != redis.Nil {
			log.Print(err)
			return false, newOrigin, err
		}
	}
	RDB.SRem(context.Background(), PROCESS_QUEUE_SET_NAME, apiKey)

	return ok, newOrigin, nil
}