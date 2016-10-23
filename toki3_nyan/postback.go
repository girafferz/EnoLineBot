package main

import (
	"github.com/line/line-bot-sdk-go/linebot"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

const postback_ramen_search = "postback_ramen_search"
const postback_beer_search = "postback_beer_search"

const postback_today_good = "postback_today_good"
const postback_today_normal = "postback_today_normal"
const postback_today_bad = "postback_today_bad"

func buildPostbackMealSearchTemplate() (*linebot.ButtonsTemplate) {
	return linebot.NewButtonsTemplate(
		"",
		"ごはんさがすにゃん",
		"なにが食べたいにゃん？",
		linebot.NewPostbackTemplateAction("らーめん", postback_ramen_search, ""),
		linebot.NewPostbackTemplateAction("びーる", postback_beer_search, ""))
}

func buildPostbackTodayReflectionTemplate() (*linebot.ButtonsTemplate) {
	return linebot.NewButtonsTemplate(
		"",
		"",
		"もうすぐきょうがおわるにゃん。きょうはいい日だったかにゃん？",
		linebot.NewPostbackTemplateAction("さいこう！", "today_good", ""),
		linebot.NewPostbackTemplateAction("ぼちぼち", "today_normal", ""),
		linebot.NewPostbackTemplateAction("だめだめ", "today_bad", ""))
}

func onReceivePostBack(event *linebot.Event, bot *linebot.Client, context context.Context) {
	profile, err := bot.GetProfile(getId(event.Source)).Do()
	if err != nil {
		log.Errorf(context, "Error occurred at get sender profile. err: %v", err)
		return
	}

	var message linebot.Message

	switch event.Postback.Data {
	case postback_ramen_search:
		if err := updateSubscriber(context, profile, action_search_ramen); err != nil {
			return
		}
		message = linebot.NewTextMessage("探したい場所を10分以内に教えてにゃん")
		break
	case postback_beer_search:
		message = linebot.NewTextMessage("探したい場所を10分以内に教えてにゃん")
		break
	case postback_today_good:
		message = linebot.NewTextMessage("それはよかったにゃん。このちょうしだにゃん")
		break
	case postback_today_normal:
		message = linebot.NewTextMessage("いつもどおりの日常がじつは大切なんだにゃん")
		break
	case postback_today_bad:
		message = linebot.NewTextMessage("きょうはビール飲んで早く寝るんだにゃん")
		break
	default:
		message = linebot.NewTextMessage("良くわかんないにゃん")
		break
	}

	if _, err := bot.ReplyMessage(event.ReplyToken, message).WithContext(context).Do(); err != nil {
		log.Errorf(context, "ReplayMessage: %v", err)
		return
	}
}
