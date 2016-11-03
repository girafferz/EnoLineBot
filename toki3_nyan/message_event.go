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
		onReceiveLocationMessage(bot, context, event)
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
		return linebot.NewTemplateMessage("ごはんさがすにゃん", buildPostbackMealSearchTemplate())
	case "今の時間をおしえて！":
		t := time.Now().In(time.FixedZone("Asia/Tokyo", 9 * 60 * 60))
		return linebot.NewTextMessage(t.Format(time.Kitchen))
	case "あ":
		return linebot.NewTemplateMessage("ていれいほうこく", buildPostbackTodayReflectionTemplate())
	default:
		return linebot.NewTextMessage("理解できない言葉だにゃん＞＜")
	}
}

func onReceiveLocationMessage(bot *linebot.Client, context context.Context, event *linebot.Event) {
	profile, err := bot.GetProfile(getId(event.Source)).Do()
	if err != nil {
		log.Errorf(context, "Error occurred at get sender profile. err: %v", err)
		return
	}

	subscriber := getSubscriber(context, profile)
	durationFromLastAction := time.Now().Unix() - subscriber.Updated
	if (durationFromLastAction > (60 * 10)) {
		// 10分以上経過していたら前回のアクションは見ない
		reply := linebot.NewTextMessage("にゃん？")
		if _, err := bot.ReplyMessage(event.ReplyToken, reply).WithContext(context).Do(); err != nil {
			log.Errorf(context, "ReplayMessage: %v", err)
			return
		}
	} else {
		onReceiveLocationAction(bot, context, event, subscriber.LastAction)
	}
}

func onReceiveLocationAction(bot *linebot.Client, context context.Context, event *linebot.Event, action string) {
	message := getLocationMessage(event.Message)
	if (message == nil) {
		return
	}

	var reply linebot.Message

	switch action {
	case action_search_ramen:
		requestLocalSearchRamen(context, message.Latitude, message.Longitude)
		reply = linebot.NewTextMessage("「" + message.Address + "」でらーめんを探すにゃん")
		break
	case action_search_beer:
		reply = linebot.NewTextMessage("「" + message.Address + "」でビールを探すにゃん")
		break
	}

	if _, err := bot.ReplyMessage(event.ReplyToken, reply).WithContext(context).Do(); err != nil {
		log.Errorf(context, "ReplayMessage: %v", err)
		return
	}
}

func getLocationMessage(message linebot.Message) (*linebot.LocationMessage) {
	switch locMessage := message.(type) {
	case *linebot.LocationMessage:
		return locMessage
	default:
		return nil
	}
}
