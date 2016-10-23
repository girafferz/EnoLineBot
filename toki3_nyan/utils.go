package main

import "github.com/line/line-bot-sdk-go/linebot"

func getId(source *linebot.EventSource) string {
	if source.Type == linebot.EventSourceTypeRoom {
		return source.RoomID;
	} else if source.Type == linebot.EventSourceTypeGroup {
		return source.GroupID;
	}

	return source.UserID;
}
