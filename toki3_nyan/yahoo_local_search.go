package main

import (
	"net/http"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"strconv"
	"time"
	"os"
	"google.golang.org/appengine/urlfetch"
)

type LocalSearchResponse struct {
	id       string    `json:"Id"`
	gid      string    `json:"Gid"`
	name     string    `json:"Name"`
	property *Property `json:"Property"`
}

type Property struct {
	uid     string `json:"Uid"`
	address string `json:"Address"`
}

func requestLocalSearchRamen(context context.Context, lat float64, lon float64) *[]LocalSearchResponse {
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

	log.Errorf(context, "success do request: %v", response)

	defer response.Body.Close()

	return nil
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
