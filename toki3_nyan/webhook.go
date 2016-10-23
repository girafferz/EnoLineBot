package main

import (
	"github.com/line/line-bot-sdk-go/linebot"
	"time"
	"google.golang.org/appengine"
	"net/http"
	"google.golang.org/appengine/taskqueue"
	"encoding/json"
	"encoding/base64"
	"net/url"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
	"os"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"fmt"
)

var botHandler *httphandler.WebhookHandler

func InitWebHook() {
	var err error
	botHandler, err = httphandler.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)

	if err != nil {
		panic(err)
		return
	}

	botHandler.HandleEvents(HandleCallback)
}

func GetBotHandler() *httphandler.WebhookHandler {
	return botHandler
}

// Webhook を受け取って TaskQueueに詰める関数
func HandleCallback(evs []*linebot.Event, r *http.Request) {
	c := appengine.NewContext(r)
	ts := make([]*taskqueue.Task, len(evs))
	for i, e := range evs {
		j, err := json.Marshal(e)
		if err != nil {
			log.Errorf(c, "json.Marshal: %v", err)
			return
		}
		data := base64.StdEncoding.EncodeToString(j)
		t := taskqueue.NewPOSTTask("/task", url.Values{"data": {data}})
		ts[i] = t
	}
	taskqueue.AddMulti(c, ts, "")
}

// 受け取ったメッセージを実際に処理する
func HandleTask(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	data := r.FormValue("data")
	if data == "" {
		log.Errorf(c, "No data")
		return
	}

	j, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Errorf(c, "base64 DecodeString: %v", err)
		return
	}

	event := new(linebot.Event)
	err = json.Unmarshal(j, event)
	if err != nil {
		log.Errorf(c, "json.Unmarshal: %v", err)
		return
	}

	bot, err := botHandler.NewClient(linebot.WithHTTPClient(urlfetch.Client(c)))
	if err != nil {
		log.Errorf(c, "newLINEBot: %v", err)
		return
	}

	log.Infof(c, "EventType: %s\nMessage: %#v", event.Type, event.Message)

	handleEvent(bot, c, event)

	w.WriteHeader(200)
}

func handleEvent(bot *linebot.Client, context context.Context, event *linebot.Event) {
	switch event.Type {
	case linebot.EventTypeFollow:
		onReceiveFollow(bot, context, event)
		break
	case linebot.EventTypeUnfollow:
		onReceiveUnFollow(bot, context, event)
		break
	case linebot.EventTypeJoin:
		// 部屋に参加した
		break
	case linebot.EventTypeLeave:
		// 部屋から離れた
		break
	case linebot.EventTypePostback:
		onReceivePostBack(event, bot, context)
		break
	case linebot.EventTypeMessage:
		onReceiveMessage(bot, context, event)
		break
	}
}

func onReceiveFollow(bot *linebot.Client, context context.Context, event *linebot.Event) {
	profile, err := bot.GetProfile(getId(event.Source)).Do()
	if err != nil {
		log.Errorf(context, "Error occurred at get sender profile. err: %v", err)
		return
	}

	if err := updateSubscriber(context, profile, action_follow); err != nil {
		return
	}

	text := fmt.Sprintf("よろしくだにゃん、げぼく「%s」", profile.DisplayName)
	message := linebot.NewTextMessage(text)

	if _, err := bot.ReplyMessage(event.ReplyToken, message).WithContext(context).Do(); err != nil {
		log.Errorf(context, "ReplayMessage: %v", err)
		return
	}
}

func onReceiveUnFollow(bot *linebot.Client, context context.Context, event *linebot.Event) {
	senderProfile, err := bot.GetProfile(getId(event.Source)).Do()
	if err != nil {
		log.Errorf(context, "Error occurred at get sender profile. err: %v", err)
		return
	}

	//購読者を削除
	key := datastore.NewKey(context, "Subscriber", senderProfile.UserID, 0, nil)
	if err := datastore.Delete(context, key); err != nil {
		log.Errorf(context, "Error occurred at delete subcriber to datastore. mid:%v, err: %v", err)
		return
	}
}

func getId(source *linebot.EventSource) string {
	if source.Type == linebot.EventSourceTypeRoom {
		return source.RoomID;
	} else if source.Type == linebot.EventSourceTypeGroup {
		return source.GroupID;
	}

	return source.UserID;
}

func onReceiveMessage(bot *linebot.Client, context context.Context, event *linebot.Event) {
	message := getMessageResponse(event)
	if (message == nil) {
		return
	}

	if _, err := bot.ReplyMessage(event.ReplyToken, message).WithContext(context).Do(); err != nil {
		log.Errorf(context, "ReplayMessage: %v", err)
		return
	}
}

func getMessageResponse(event *linebot.Event) linebot.Message {
	switch message := event.Message.(type) {
	case *linebot.TextMessage:
		return getTextMessageResponse(message.Text)
	case *linebot.ImageMessage:
		return linebot.NewTextMessage("画像だにゃん")
	case *linebot.StickerMessage:
		return linebot.NewTextMessage("スタンプだにゃん")
	default:
		return nil
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

