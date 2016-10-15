package main

import (
	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
	"os"
	"net/http"
)

func init() {
	err := godotenv.Load("line.env")
	if err != nil {
		panic(err)
	}

	botHandler, err := httphandler.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)
	botHandler.HandleEvents(HandleCallback)

	SetBotHandler(botHandler)

	http.Handle("/callback", botHandler)
	http.HandleFunc("/task", HandleTask)
}
