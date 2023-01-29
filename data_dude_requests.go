package main

import (
	"time"
	"bytes"
	"errors"
	"context"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"
	. "github.com/fireacademy/golden-gate/redis"
)

type GetAPIKeyInfoArgs struct {
    APIKey string `json:"api_key"`
}

func GetAPIKeyInfoFromDataDude(ctx context.Context, apiKey string) (DataDudeResponse, error) {
	ctx, span := GetSpan(ctx, "GetAPIKeyInfoFromDataDude")
	defer span.End()
	span.SetAttributes(
		attribute.String("api_key", apiKey),
	)

	errObj := DataDudeResponse{
		Success: false,
	}
	args := GetAPIKeyInfoArgs{
		APIKey: apiKey,
	}

	args_JSON, err := json.Marshal(args)
	if err != nil {
		LogError(ctx, err, "error while encoding JSOn arguments")
		return errObj, err
	}

	bodyReader := bytes.NewReader(args_JSON)

	req, err := http.NewRequest(http.MethodPost, DataDudeAPIKeyInfoURL, bodyReader)
	if err != nil {
		LogError(ctx, err, "error while creating request")
		return errObj, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Management-Token", ManagementToken) // f4ee5c1eb39d14517e90b9cejustkidding

	client := http.Client{
		Timeout: 5 * time.Second,
  	}

  	res, err := client.Do(req)
  	if err != nil {
		LogError(ctx, err, "error while making request")
		return errObj, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		LogError(ctx, err, "error while reading response")
		return DataDudeResponse{
			Success: false,
		}, err
	}
	if res.StatusCode != 200 {
		LogError(ctx, errors.New("data-dude error"), "data-dude error; resp body: " + string(resBody))
		return errObj, errors.New("")
	}

	var resp DataDudeResponse
	err = json.Unmarshal(resBody, &resp)
	if err != nil {
		LogError(ctx, err, "error while decoding JSON response" + string(resBody))
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

func TellDataDudeToBillCreditsPackage(ctx context.Context, custId string) (bool /* success */, error) {
	ctx, span := GetSpan(ctx, "TellDataDudeToBillCreditsPackage")
	defer span.End()
	span.SetAttributes(
		attribute.String("customer_id", custId),
	)

	args := BillCreditsPackageArgs{
		StripeCustomerID: custId,
	}

	args_JSON, err := json.Marshal(args)
	if err != nil {
		LogError(ctx, err, "error while encoding JSOn arguments")
		return false, err
	}

	bodyReader := bytes.NewReader(args_JSON)

	req, err := http.NewRequest(http.MethodPost, DataDudeBillCreditsPackageURL, bodyReader)
	if err != nil {
		LogError(ctx, err, "error while creating request")
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Management-Token", ManagementToken) // f4ee5c1eb39d14517e90b9cejustkidding

	client := http.Client{
		Timeout: 5 * time.Second,
  	}

  	res, err := client.Do(req)
  	if err != nil {
		LogError(ctx, err, "error while making request")
		return false, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		LogError(ctx, err, "error while reading response")
		return false, err
	}
	if res.StatusCode != 200 {
		LogError(ctx, errors.New("data-dude error"), "data-dude error; resp body: " + string(resBody) + "; arguments: " + string(args_JSON))
		return false, errors.New("")
	}

	var resp BillCreditsPackageResponse
	err = json.Unmarshal(resBody, &resp)
	if err != nil {
		LogError(ctx, err, "error while decoding JSON response: " + string(resBody))
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

func TellDataDudeToRecordUsage(ctx context.Context, apiKey string, credits int64) (bool /* success */, error) {
	ctx, span := GetSpan(ctx, "TellDataDudeToRecordUsage")
	defer span.End()
	span.SetAttributes(
		attribute.String("api_key", apiKey),
		attribute.Int64("credits", credits),
	)

	args := RecordUsageArgs{
		APIKey: apiKey,
		Credits: credits,
	}

	args_JSON, err := json.Marshal(args)
	if err != nil {
		LogError(ctx, err, "error while encoding JSOn arguments")
		return false, err
	}

	bodyReader := bytes.NewReader(args_JSON)

	req, err := http.NewRequest(http.MethodPost, DataDudeRecordUsageURL, bodyReader)
	if err != nil {
		LogError(ctx, err, "error while creating request")
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Management-Token", ManagementToken) // f4ee5c1eb39d14517e90b9cejustkidding

	client := http.Client{
		Timeout: 10 * time.Second,
  	}

  	res, err := client.Do(req)
  	if err != nil {
		LogError(ctx, err, "error while making request")
		return false, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		LogError(ctx, err, "error while reading response")
		return false, err
	}
	if res.StatusCode != 200 {
		LogError(ctx, errors.New("data-dude error"), "data-dude error; resp body: " + string(resBody) + "; arguments: " + string(args_JSON))
		return false, errors.New("")
	}

	var resp RecordUsageResponse
	err = json.Unmarshal(resBody, &resp)
	if err != nil {
		LogError(ctx, err, "error while decoding JSON response: " + string(resBody))
		return false, err
	}

	return resp.Success, nil
}