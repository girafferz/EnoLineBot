package main

import (
	"github.com/line/line-bot-sdk-go/linebot"
	"time"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func onReceiveMessage(bot *linebot.Client, context context.Context, event *linebot.Event) {
	var reply linebot.Message

	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		reply = getTextMessageResponse(message.Text)
		break
	case *linebot.ImageMessage:
		reply = linebot.NewTextMessage("画像だにゃん")
		break
	case *linebot.StickerMessage:
		reply = linebot.NewTextMessage("スタンプだにゃん")
		break
	case *linebot.LocationMessage:
		log.Infof(context, "location:%s, %f , %f", message.Address, message.Latitude, message.Longitude)
		break
	default:
		break
	}

	if (reply == nil) {
		return
	}

	if _, err := bot.ReplyMessage(event.ReplyToken, reply).WithContext(context).Do(); err != nil {
		log.Errorf(context, "ReplayMessage: %v", err)
		return
	}
}

func getTextMessageResponse(text string) linebot.Message {
	switch text {
	case "はらへ":
		return linebot.NewTemplateMessage("search_meal", buildPostbackMealSearchTemplate())
	case "今の時間をおしえて！":
		t := time.Now().In(time.FixedZone("Asia/Tokyo", 9 * 60 * 60))
		return linebot.NewTextMessage(t.Format(time.Kitchen))
	case "あ":
		return linebot.NewTemplateMessage("ていれいほうこく", buildPostbackTodayReflectionTemplate())
	default:
		return linebot.NewTextMessage("理解できない言葉だにゃん＞＜")
	}
}

