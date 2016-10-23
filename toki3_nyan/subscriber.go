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
const action_search_beer = "action_search_beer"

// 購読者エンティティ
type subscriber struct {
	DisplayName string
	ID          string
	LastAction  string
	Updated     int64
}

func getSubscriberKey(context context.Context, userId string) *datastore.Key {
	return datastore.NewKey(context, subscriber_key, userId, 0, nil)
}

func updateSubscriber(context context.Context, profile *linebot.UserProfileResponse, action string) (error) {
	entity := subscriber{
		DisplayName: profile.DisplayName,
		ID: profile.UserID,
		LastAction: action,
		Updated: time.Now().Unix(),
	}

	key := getSubscriberKey(context, profile.UserID)

	if _, err := datastore.Put(context, key, &entity); err != nil {
		log.Errorf(context, "Error occurred at put subcriber to datastore. err: %v", err)
		return err
	}

	return nil;
}

func getSubscriber(context context.Context, profile *linebot.UserProfileResponse) *subscriber {
	entities := new(subscriber)
	key := getSubscriberKey(context, profile.UserID)

	if err := datastore.Get(context, key, entities); err != nil {
		log.Errorf(context, "Error occurred at get subcriber from datastore. err: %v", err)
	}

	return entities
}
