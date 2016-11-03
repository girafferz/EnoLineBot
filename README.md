# EnoLineBot
cat line bot "toki3_nyan"

# init
```
$ brew install go-app-engine-64
$ export GOPATH=`pwd`
```

# go get
```
$ go get -u github.com/line/line-bot-sdk-go/linebot
$ go get -u github.com/joho/godotenv
$ go get -u google.golang.org/appengine
```

# files

```
$ tree toki3_nyan/
toki3_nyan/
├── app.go
├── cron.yaml
├── message_event.go
├── postback_event.go
├── push.go
├── subscriber.go
├── utils.go
└── webhook.go
```

# set your own confing files

## toki3_nyan/app.yaml
```
$ cat toki3_nyan/app.yaml
application: (google cloud platform application id <projectID>)
version: 1
runtime: go
api_version: go1

handlers:
- url: /task.*
  script: _go_app
  login: admin
  secure: always
- url: /.*
  script: _go_app
  secure: always
```

## toki3_nyan/line.env

```
$ cat demo/line.env
LINE_BOT_CHANNEL_SECRET=(line bot channnel secret)
LINE_BOT_CHANNEL_TOKEN=(line bot channel token)
```

# deoloy
after modify code
```
$ cd toki3_nyan
$ goapp deploy
```

# line developer setup
set variable Webhook like "https://***.appspot.com/callback"


