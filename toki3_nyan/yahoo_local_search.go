package main

import (
	"net/http"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"strconv"
	"time"
	"os"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"encoding/json"
)

type LocalSearchResponse struct {
	id        string
	name      string
	linkUrl   string
	address   string
	leadImage string
}

func requestLocalSearchRamen(context context.Context, lat float64, lon float64) []LocalSearchResponse {
	request := buildRequest(context, "0106", lat, lon)
	if request == nil {
		return nil
	}

	client := urlfetch.Client(context)
	client.Timeout = time.Duration(5) * time.Second

	response, err := client.Do(request)
	if err != nil {
		log.Errorf(context, "failed do request: %v", err)
		return nil
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Errorf(context, "failed do request status code invalid: %v", response)
		return nil
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorf(context, "failed read response body: %v", response.Body)
		return nil
	}

	return parseResponseBody(context, responseBody)
}

func buildRequest(context context.Context, gc string, lat float64, lon float64) (*http.Request) {
	url := "http://search.olp.yahooapis.jp/OpenLocalPlatform/V1/localSearch"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Errorf(context, "failed new request: %v", err)
		return nil
	}

	q := request.URL.Query()
	q.Add("appid", os.Getenv("YAHOO_APPID"))
	q.Add("gc", gc)
	q.Add("output", "json")
	q.Add("lat", strconv.FormatFloat(lat, 'f', 12, 64))
	q.Add("lon", strconv.FormatFloat(lon, 'f', 12, 64))
	q.Add("dist", "1")
	q.Add("sort", "hybrid")

	request.URL.RawQuery = q.Encode()

	return request
}

func parseResponseBody(context context.Context, responseBody []byte) []LocalSearchResponse {
	var localSearchResponse interface{}
	if err := json.Unmarshal(responseBody, &localSearchResponse); err != nil {
		log.Errorf(context, "failed response json unmarshal: %v", err)
		return nil
	}

	if _, ok := localSearchResponse.(map[string]interface{}); !ok {
		log.Errorf(context, "response is invalid %v", localSearchResponse)
		return nil
	}

	if _, ok := localSearchResponse.(map[string]interface{})["Feature"]; !ok {
		log.Errorf(context, "response is invalid %v", localSearchResponse)
		return nil
	}

	if _, ok := localSearchResponse.(map[string]interface{})["Feature"].([]interface{}); !ok {
		log.Errorf(context, "response is invalid %v", localSearchResponse)
		return nil
	}

	features := localSearchResponse.(map[string]interface{})["Feature"].([]interface{})
	featuresSize := len(features)
	if (featuresSize < 1) {
		// 検索結果が1件もない
		return nil
	}

	responses := make([]LocalSearchResponse, featuresSize)

	for i := 0; i < featuresSize; i++ {
		feature := features[i].(map[string]interface{})
		property := feature["Property"].(map[string]interface{})

		response := LocalSearchResponse{}
		response.id = feature["Id"].(string)
		response.name = feature["Name"].(string)
		response.linkUrl = "http://loco.yahoo.co.jp/place/g-" + feature["Gid"].(string)
		response.address = property["Address"].(string)

		if _, ok := property["LeadImage"]; ok {
			response.leadImage = property["LeadImage"].(string)
		}

		responses[i] = response
	}

	return responses
}
