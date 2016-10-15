package main

import (
	"github.com/joho/godotenv"
	"net/http"
)

func init() {
	err := godotenv.Load("line.env")
	if err != nil {
		panic(err)
	}

	InitWebHook()

	http.Handle("/callback", GetBotHandler())
	http.HandleFunc("/task", HandleTask)
}
