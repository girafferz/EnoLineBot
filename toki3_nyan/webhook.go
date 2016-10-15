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
		break
	case linebot.EventTypeJoin:
		// 部屋に参加した
		break
	case linebot.EventTypeLeave:
		// 部屋から離れた
		break
	case linebot.EventTypePostback:
		onReceivePostBack(event.Postback.Data, bot, event.ReplyToken, context)
		break
	case linebot.EventTypeMessage:
		onReceiveMessage(bot, context, event)
		break
	}
}

func onReceiveFollow(bot *linebot.Client, context context.Context, event *linebot.Event) {
	senderProfile, err := bot.GetProfile(getId(event.Source)).Do()
	if err != nil {
		log.Errorf(context, "Error occurred at get sender profile. err: %v", err)
		return
	}

	entity := subscriber{
		DisplayName: senderProfile.DisplayName,
		ID: senderProfile.UserID,
	}

	key := datastore.NewKey(context, subscriber_key, senderProfile.UserID, 0, nil)
	if _, err := datastore.Put(context, key, &entity); err != nil {
		log.Errorf(context, "Error occurred at put subcriber to datastore. err: %v", err)
		return
	}

	text := fmt.Sprintf("よろしくだにゃん、げぼく「%s」", senderProfile.DisplayName)
	message := linebot.NewTextMessage(text)

	if _, err := bot.ReplyMessage(event.ReplyToken, message).WithContext(context).Do(); err != nil {
		log.Errorf(context, "ReplayMessage: %v", err)
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
		return linebot.NewTextMessage("良くわかんないにゃん")
	}

}

func getTextMessageResponse(text string) linebot.Message {
	switch text {
	case "はらへ":
		template := linebot.NewConfirmTemplate(
			"なにがたべたいにゃん",
			linebot.NewPostbackTemplateAction("らーめん", "ramen", ""),
			linebot.NewPostbackTemplateAction("びーる", "beer", ""))
		return linebot.NewTemplateMessage("らーめん種類選択", template)
	case "今の時間をおしえて！":
		t := time.Now().In(time.FixedZone("Asia/Tokyo", 9 * 60 * 60))
		return linebot.NewTextMessage(t.Format(time.Kitchen))
	default:
		return linebot.NewTextMessage("理解できない言葉だにゃん＞＜")
	}
}

func onReceivePostBack(data string, bot *linebot.Client, replyToken string, context context.Context) {
	var message linebot.Message

	switch data {
	case "ramen":
		message = linebot.NewTextMessage("らーめん探すにゃん")
	case "beer":
		message = linebot.NewTextMessage("ビール探すにゃん")
	default:
		message = linebot.NewTextMessage("良くわかんないにゃん")
	}

	if _, err := bot.ReplyMessage(replyToken, message).WithContext(context).Do(); err != nil {
		log.Errorf(context, "ReplayMessage: %v", err)
		return
	}
}
