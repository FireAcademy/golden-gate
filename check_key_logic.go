package main

import (
	"bytes"
	"errors"
	"strconv"
	"net/http"
	"encoding/json"
)

type GetAPIKeyInfoArgs struct {
    APIKey string `json:"api_key"`
}

func GetInfoFromDataDude(apiKey string) (DataDudeResponse, error) {
	errObj := DataDudeResponse{
		Success: false,
	}
	args := GetAPIKeyInfoArgs{
		APIKey: apiKey,
	}

	args_JSON, err = json.Marshal(args)
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
	err := json.Unmarshal(resBody, &resp)
	if err != nil {
		log.Print(err)
		return errObj, err
	}

	return resp, nil
}

func IsAPIKeyOK(apiKey String, info DataDudeResponse, creditsToProcess int64) (int64 /* creditsToBill */, bool /* ok */, string /* newOrigin */, bool /* purchaseCreditsPackage */) {
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
		if info.User.StripeCustomerID.Valid && info.user.AutoPurchaseCreditsPackages {
			return creditsToBill, true, origin, true
		} else {
			return 0, false, origin, false
		}
	}

	return creditsToBill, true, origin, false
}

func RefreshAPIKey(apiKey string) (bool /* canBeUsed */, error /* err */) {
	info, err := GetInfoFromDataDude(apiKey)
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
		todo
	}

	if creditsToBill > 0 {
		err := BillCredits(apiKey, creditsToBill)
		if err != nil {
			log.Print("Could not bill credits for " + apiKey)
			log.Print(err)
		}
	}

	// update redis
	valueToSet := "not-ok"
	if ok {
		valueToSet := API_KEY_OK_VALUE
	}
	err := RDB.Set(context.Background(), apiKey, valueToSet).Error()
	if err != nil {
		log.Print(err)
		return true, err
	}

	if ok {
		err := RDB.Set(context.Background(), ORIGIN_PREFIX + apiKey, newOrigin).Error()
		if err != nil {
			log.Print(err)
			return true, err
		}
		err := RDB.DecrBy(context.Background(), USAGE_PREFIX + apiKey, creditsToProcess).Error()
		if err != nil {
			log.Print(err)
			return true, err
		}
	} else {
		err := RDB.Del(context.Background(), ORIGIN_PREFIX + apiKey).Error()
		if err != nil && err != redis.Nil {
			log.Print(err)
			return false, err
		}
		err := RDB.Del(context.Background(), USAGE_PREFIX + apiKey, creditsToProcess).Error()
		if err != nil && err != redis.Nil {
			log.Print(err)
			return false, err
		}
	}
	RDB.SRem(context.Background(), PROCESS_QUEUE_SET_NAME, apiKey)

	return ok, nil
}