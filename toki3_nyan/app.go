package main

import (
    "encoding/base64"
    "encoding/json"
    "net/http"
    "net/url"
    "os"
    "google.golang.org/appengine"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/taskqueue"
    "google.golang.org/appengine/urlfetch"
    "github.com/joho/godotenv"
    "github.com/line/line-bot-sdk-go/linebot"
    "github.com/line/line-bot-sdk-go/linebot/httphandler"
    "time"
)

var botHandler *httphandler.WebhookHandler

func init() {
    err := godotenv.Load("line.env")
    if err != nil {
        panic(err)
    }

    botHandler, err = httphandler.New(
        os.Getenv("LINE_BOT_CHANNEL_SECRET"),
        os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
    )
    botHandler.HandleEvents(handleCallback)

    http.Handle("/callback", botHandler)
    http.HandleFunc("/task", handleTask)
}

// Webhook を受け取って TaskQueueに詰める関数
func handleCallback(evs []*linebot.Event, r *http.Request) {
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
func handleTask(w http.ResponseWriter, r *http.Request) {
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

    switch event.Type {
    case "postback":
    case "message":

    }

    log.Infof(c, "EventType: %s\nMessage: %#v", event.Type, event.Message)

    m := getResponseMessage(event)
    if _, err = bot.ReplyMessage(event.ReplyToken, m).WithContext(c).Do(); err != nil {
        log.Errorf(c, "ReplayMessage: %v", err)
        return
    }

    w.WriteHeader(200)
}

func getResponseMessage(event *linebot.Event) linebot.Message {
    switch event.Type {
    case "postback":
        return onReceivePostBack(event.Postback.Data)
    case "message":
        switch message := event.Message.(type) {
        case *linebot.TextMessage:
            return onReceiveTextMessage(message.Text)
        case *linebot.ImageMessage:
            return linebot.NewTextMessage("画像だにゃん")
        case *linebot.StickerMessage:
            return linebot.NewTextMessage("スタンプだにゃん")
        default:
            return linebot.NewTextMessage("良くわかんないにゃん")
        }
    }
    return linebot.NewTextMessage("良くわかんないにゃん")
}

func onReceiveTextMessage(text string) linebot.Message {
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

func onReceivePostBack(data string) linebot.Message {
    switch data {
    case "ramen":
        return linebot.NewTextMessage("らーめん探すにゃん")
    case "beer":
        return linebot.NewTextMessage("ビール探すにゃん")
    default:
        return linebot.NewTextMessage("良くわかんないにゃん")
    }
}