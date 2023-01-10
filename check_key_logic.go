package main

import (
	"log"
	"time"
	"bytes"
	"errors"
	"context"
	"strconv"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/go-redis/redis/v8"
)

type GetAPIKeyInfoArgs struct {
    APIKey string `json:"api_key"`
}

func GetAPIKeyInfoFromDataDude(apiKey string) (DataDudeResponse, error) {
	errObj := DataDudeResponse{
		Success: false,
	}
	args := GetAPIKeyInfoArgs{
		APIKey: apiKey,
	}

	args_JSON, err := json.Marshal(args)
	if err != nil {
		log.Print(err)
		return errObj, err
	}

	bodyReader := bytes.NewReader(args_JSON)

	req, err := http.NewRequest(http.MethodPost, DataDudeAPIKeyInfoURL, bodyReader)
	if err != nil {
		log.Print(err)
		return errObj, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Management-Token", "ManagementToken") // f4ee5c1eb39d14517e90b9cejustkidding

	client := http.Client{
		Timeout: 30 * time.Second,
  	}

  	res, err := client.Do(req)
  	if err != nil {
		log.Print(err)
		return errObj, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return DataDudeResponse{
			Success: false,
		}, err
	}
	if res.StatusCode != 200 {
		log.Print("data-dude error")
		log.Print(resBody)
		return errObj, errors.New("")
	}

	var resp DataDudeResponse
	err = json.Unmarshal(resBody, &resp)
	if err != nil {
		log.Print(err)
		return errObj, err
	}

	return resp, nil
}

type BillCreditsPackageArgs struct {
    StripeCustomerID string `json:"stripe_customer_id"`
}

type BillCreditsPackageResponse struct {
    Success bool `json:"success"`
}

func TellDataDudeToBillCreditsPackage(custId string) (bool /* success */, error) {
	args := BillCreditsPackageArgs{
		StripeCustomerID: custId,
	}

	args_JSON, err := json.Marshal(args)
	if err != nil {
		log.Print(err)
		return false, err
	}

	bodyReader := bytes.NewReader(args_JSON)

	req, err := http.NewRequest(http.MethodPost, DataDudeBillCreditsPackageURL, bodyReader)
	if err != nil {
		log.Print(err)
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Management-Token", "ManagementToken") // f4ee5c1eb39d14517e90b9cejustkidding

	client := http.Client{
		Timeout: 30 * time.Second,
  	}

  	res, err := client.Do(req)
  	if err != nil {
		log.Print(err)
		return false, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return false, err
	}
	if res.StatusCode != 200 {
		log.Print("data-dude error")
		log.Print(resBody)
		return false, errors.New("")
	}

	var resp BillCreditsPackageResponse
	err = json.Unmarshal(resBody, &resp)
	if err != nil {
		log.Print(err)
		return false, err
	}

	return resp.Success, nil
}

/* formerly RecordUsageWithoutAnyChecksWhatsoeverArgs */
type RecordUsageArgs struct {
    APIKey string `json:"api_key"`
    Credits int64 `json:"credits"`
}

type RecordUsageResponse struct {
    Success bool `json:"success"`
}

func TellDataDudeToRecordUsage(apiKey string, credits int64) (bool /* success */, error) {
	args := RecordUsageArgs{
		APIKey: apiKey,
		Credits: credits,
	}

	args_JSON, err := json.Marshal(args)
	if err != nil {
		log.Print(err)
		return false, err
	}

	bodyReader := bytes.NewReader(args_JSON)

	req, err := http.NewRequest(http.MethodPost, DataDudeRecordUsageURL, bodyReader)
	if err != nil {
		log.Print(err)
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Management-Token", "ManagementToken") // f4ee5c1eb39d14517e90b9cejustkidding

	client := http.Client{
		Timeout: 30 * time.Second,
  	}

  	res, err := client.Do(req)
  	if err != nil {
		log.Print(err)
		return false, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
		return false, err
	}
	if res.StatusCode != 200 {
		log.Print("data-dude error")
		log.Print(resBody)
		return false, errors.New("")
	}

	var resp RecordUsageResponse
	err = json.Unmarshal(resBody, &resp)
	if err != nil {
		log.Print(err)
		return false, err
	}

	return resp.Success, nil
}

func IsAPIKeyOK(apiKey string, info DataDudeResponse, creditsToProcess int64) (int64 /* creditsToBill */, bool /* ok */, string /* newOrigin */, bool /* purchaseCreditsPackage */) {
	origin := info.APIKey.Origin
	creditsToBill := creditsToProcess

	if info.APIKey.MonthlyCreditLimit < info.Usage.Credits + creditsToBill {
		log.Print("Missed a few credits on API key " + apiKey)
		creditsToBill = info.APIKey.MonthlyCreditLimit - info.Usage.Credits
	}

	if creditsToBill == 0 || info.APIKey.Disabled {
		return 0, false, origin, false
	}

	if creditsToBill > info.User.RemainingCredits {
		if info.User.StripeCustomerID.Valid && info.User.AutoPurchaseCreditsPackages {
			return creditsToBill, true, origin, true
		} else {
			return 0, false, origin, false
		}
	}

	return creditsToBill, true, origin, false
}

func RefreshAPIKey(apiKey string) (bool /* canBeUsed */, error /* err */) {
	info, err := GetAPIKeyInfoFromDataDude(apiKey)
	if err != nil {
		return false, err
	}

	creditsToProcessStr, err := RDB.Get(context.Background(), apiKey).Result()
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
		return true, err
	}

	if ok {
		err = RDB.Set(context.Background(), ORIGIN_PREFIX + apiKey, newOrigin, 0).Err()
		if err != nil {
			log.Print(err)
			return true, err
		}
		err = RDB.DecrBy(context.Background(), USAGE_PREFIX + apiKey, creditsToProcess).Err()
		if err != nil {
			log.Print(err)
			return true, err
		}
	} else {
		err = RDB.Del(context.Background(), ORIGIN_PREFIX + apiKey).Err()
		if err != nil && err != redis.Nil {
			log.Print(err)
			return false, err
		}
		err = RDB.Del(context.Background(), USAGE_PREFIX + apiKey).Err()
		if err != nil && err != redis.Nil {
			log.Print(err)
			return false, err
		}
	}
	RDB.SRem(context.Background(), PROCESS_QUEUE_SET_NAME, apiKey)

	return ok, nil
}