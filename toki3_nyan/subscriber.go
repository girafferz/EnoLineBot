package main

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"time"
	"github.com/line/line-bot-sdk-go/linebot"
	"google.golang.org/appengine/log"
)

const subscriber_key = "Subscriber"

const action_follow = "action_follow"
const action_search_ramen = "action_search_ramen"

// 購読者エンティティ
type subscriber struct {
	DisplayName string
	ID          string
	LastAction  string
	Updated     int64
}

func updateSubscriber(context context.Context, profile *linebot.UserProfileResponse, action string) (error) {
	entity := subscriber{
		DisplayName: profile.DisplayName,
		ID: profile.UserID,
		LastAction: action,
		Updated: time.Now().Unix(),
	}

	key := datastore.NewKey(context, subscriber_key, profile.UserID, 0, nil)

	if _, err := datastore.Put(context, key, &entity); err != nil {
		log.Errorf(context, "Error occurred at put subcriber to datastore. err: %v", err)
		return err
	}

	return nil;
}
