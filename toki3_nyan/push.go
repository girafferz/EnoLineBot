package main

import (
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
	"os"
)

func PushCron(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)

	log.Infof(context, "====== start push cron ======")

	channelSecret := os.Getenv("LINE_BOT_CHANNEL_SECRET")
	channelToken := os.Getenv("LINE_BOT_CHANNEL_TOKEN")

	// Appengineのurlfetchを使用する
	bot, err := linebot.New(channelSecret, channelToken, linebot.WithHTTPClient(urlfetch.Client(context)))
	if err != nil {
		log.Errorf(context, "Error occurred at create linebot client: %v", err)
		w.WriteHeader(500)
		return
	}

	//データストアから購読者のMIDを取得
	q := datastore.NewQuery(subscriber_key)
	var subscribers []subscriber
	if _, err := q.GetAll(context, &subscribers); err != nil {
		log.Errorf(context, "Error occurred at get-all from datastore. err: %v", err)
		w.WriteHeader(500)
		return
	}

	//全員に送信
	message := linebot.NewTextMessage("あと30分できょうがおわるにゃん。きょうはいい日だったかにゃん？")

	for _, current := range subscribers {
		if _, err := bot.PushMessage(current.ID, message).WithContext(context).Do(); err != nil {
			log.Errorf(context, "Error occurred at send message: %v", err)
			continue
		}
	}

	w.WriteHeader(200)
}
