package main

import (
	"log"
	"time"
	"bytes"
	"errors"
	"net/http"
	"io/ioutil"
	"encoding/json"
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
	req.Header.Set("X-Management-Token", ManagementToken) // f4ee5c1eb39d14517e90b9cejustkidding

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
		log.Print(str(resBody))
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
	req.Header.Set("X-Management-Token", ManagementToken) // f4ee5c1eb39d14517e90b9cejustkidding

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
		log.Print(str(resBody))
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
	req.Header.Set("X-Management-Token", ManagementToken) // f4ee5c1eb39d14517e90b9cejustkidding

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
		log.Print(str(resBody))
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