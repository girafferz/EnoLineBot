package main

import (
	"github.com/line/line-bot-sdk-go/linebot"
	"time"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"fmt"
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
	case "天気":
		return linebot.NewTextMessage("天気にゃん http://weather.yahoo.co.jp/weather/")
	case "ニュース":
		str := `ニュースにゃん
http://bit.ly/2fgqvkD
http://bit.ly/2es6HJP
http://tcrn.ch/2eiuIzK`
		return linebot.NewTextMessage(str)
	case "動画":
		return linebot.NewTextMessage("動画にゃん http://bit.ly/2flJcQ3")
	case "画像":
		return linebot.NewTextMessage("画像にゃん http://bit.ly/2fgCax9")
	case "なう":
		return linebot.NewTextMessage("なうにゃん http://bit.ly/2flKpab")
	case "ヘルプ":
		str := `はらへ
今の時間をおしえて！
天気
ニュース
動画
画像
なう
と聞くにゃん`
		return linebot.NewTextMessage(str)
	default:
		return linebot.NewTextMessage("理解できない言葉だにゃん＞＜ 「ヘルプ」と聞くにゃん")
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
		responses := requestLocalSearchRamen(context, message.Latitude, message.Longitude)
		if responses == nil || len(responses) < 1 {
			reply = linebot.NewTextMessage("らーめんが見つからなかったにゃん")
		} else {
			reply = linebot.NewTemplateMessage(
				fmt.Sprintf("お店が%d件みつかったにゃん", len(responses)),
				buildCarouselTemplate(responses),
			)
		}
		break
	case action_search_beer:
		responses := requestLocalSearchBeer(context, message.Latitude, message.Longitude)
		if responses == nil || len(responses) < 1 {
			reply = linebot.NewTextMessage("ビールが見つからなかったにゃん")
		} else {
			reply = linebot.NewTemplateMessage(
				fmt.Sprintf("お店が%d件みつかったにゃん", len(responses)),
				buildCarouselTemplate(responses),
			)
		}
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

func buildCarouselTemplate(responses []LocalSearchResponse) *linebot.CarouselTemplate {
	len := len(responses)

	var cc []*linebot.CarouselColumn
	for i := 0; i < len; i++ {
		if i > 4 {
			// カルーセルは5件上限
			break
		}

		cc = append(cc, buildCarouselColumn(responses[i]))
	}
	return linebot.NewCarouselTemplate(cc...);
}

func buildCarouselColumn(response LocalSearchResponse) *linebot.CarouselColumn {
	return linebot.NewCarouselColumn(
		"",
		response.name,
		response.address,
		linebot.NewURITemplateAction("ページを見る", response.linkUrl),
	)
}
