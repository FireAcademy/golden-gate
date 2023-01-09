package main

import (
	"bytes"
	"errors"
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

func RefreshAPIKey(apiKey string) (bool /* canBeUsed */, error /* err */) {
	keyInfo, err := GetInfoFromDataDude(apiKey)
	if err != nil {
		return false, err
	}

	// do things...
	return true, nil
}