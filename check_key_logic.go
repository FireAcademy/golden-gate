package main

import (
	"log"
	"time"
	"errors"
	"context"
	"strconv"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	. "github.com/fireacademy/golden-gate/redis"
)

func IsAPIKeyOK(ctx context.Context, apiKey string, info DataDudeResponse, creditsToProcess int64) (int64 /* creditsToBill */, bool /* ok */, string /* newOrigin */, bool /* purchaseCreditsPackage */) {
	ctx, span := GetSpan(ctx, "IsAPIKeyOK")
	defer span.End()
	span.SetAttributes(
		attribute.String("api_key", apiKey),
		attribute.Int64("credits_to_process", creditsToProcess),
		attribute.Bool("info.api_key.disabled", info.APIKey.Disabled),
		attribute.Int64("info.api_key.monthly_credit_limit", info.APIKey.MonthlyCreditLimit),
		attribute.String("info.api_key.origin", info.APIKey.Origin),
		attribute.Int64("info.user.remaining_credits", info.User.RemainingCredits),
		attribute.String("info.user.stripe_customer_id", info.User.StripeCustomerID.String),
		attribute.Bool("info.user.auto_purchase_credits_packages", info.User.AutoPurchaseCreditsPackages),
		attribute.Int64("info.usage.credits", info.Usage.Credits),
	)
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

	if info.APIKey.Disabled {
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

func RefreshAPIKey(ctx context.Context, apiKey string) (bool /* canBeUsed */, string /* origin */, error /* err */) {
	ctx, span := GetSpan(ctx, "RefreshAPIKey")
	defer span.End()
	span.SetAttributes(
		attribute.String("api_key", apiKey),
	)

	state, err := RDB.Get(ctx, API_KEY_PREFIX + apiKey).Result()
	if err == nil && state == API_KEY_PENDING_CHECK_VALUE {
		log.Print("api key is already being checked; calling CheckAPIKeyQuickly...")
		return CheckAPIKeyQuickly(ctx, apiKey)
	}

	err = RDB.Set(ctx, API_KEY_PREFIX + apiKey, API_KEY_PENDING_CHECK_VALUE, 2 * time.Second).Err()
	if err != nil {
		LogError(ctx, err, "could not set API key status in redis")
		return false, "", err
	}

	info, err := GetAPIKeyInfoFromDataDude(ctx, apiKey)
	if err != nil {
		LogError(ctx, err, "error with data-dude thingy")
		return false, "", err
	}

	creditsToProcessStr, err := RDB.Get(ctx, USAGE_PREFIX + apiKey).Result()
	if err != nil {
		if err != redis.Nil {
			LogError(ctx, err, "could not get credits to process")
		}
		creditsToProcessStr = "0"
	}
	creditsToProcess, err := strconv.ParseInt(creditsToProcessStr, 10, 64)
	if err != nil {
		LogError(ctx, err, "error when converting credits to process to int: " + creditsToProcessStr)
		creditsToProcess = 0
	}

	creditsToBill, ok, newOrigin, purchaseCreditsPackage := IsAPIKeyOK(ctx, apiKey, info, creditsToProcess)

	if purchaseCreditsPackage {
		success, err := TellDataDudeToBillCreditsPackage(ctx, info.User.StripeCustomerID.String)

		if err != nil || !success {
			if err != nil {
				LogError(ctx, err, "could not bill extra credits package due to error")
			}

			msg := "Could not bill extra package for " + info.User.StripeCustomerID.String
			LogError(ctx, errors.New(msg), msg)
		}
	}

	if creditsToBill > 0 {
		success, err := TellDataDudeToRecordUsage(ctx, apiKey, creditsToBill)
		if err != nil || !success {
			msg := "Could not bill credits for " + apiKey
			if err != nil {
				err = errors.New(msg)
			}

			LogError(ctx, err, msg)
		}
	}

	// update redis
	valueToSet := "not-ok"
	if ok {
		valueToSet = API_KEY_OK_VALUE
	}
	err = RDB.Set(ctx, API_KEY_PREFIX + apiKey, valueToSet, 0).Err()
	if err != nil {
		LogError(ctx, err, "could not set API key status in redis")
		return true, newOrigin, err
	}

	if ok {
		err = RDB.Set(ctx, ORIGIN_PREFIX + apiKey, newOrigin, 0).Err()
		if err != nil {
			LogError(ctx, err, "could not set API key origin in redis")
			return true, newOrigin, err
		}
		err = RDB.DecrBy(ctx, USAGE_PREFIX + apiKey, creditsToProcess).Err()
		if err != nil {
			LogError(ctx, err, "could not decrease API key credits (to bill) in redis")
			return true, newOrigin, err
		}
	} else {
		err = RDB.Del(ctx, ORIGIN_PREFIX + apiKey).Err()
		if err != nil && err != redis.Nil {
			LogError(ctx, err, "could not delete API key origin in redis")
			return false, newOrigin, err
		}
		err = RDB.Del(ctx, USAGE_PREFIX + apiKey).Err()
		if err != nil && err != redis.Nil {
			LogError(ctx, err, "could not delete API key usage in redis")
			return false, newOrigin, err
		}
	}
	RDB.SRem(ctx, PROCESS_QUEUE_SET_NAME, apiKey)

	return ok, newOrigin, nil
}