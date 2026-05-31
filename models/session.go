package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Session struct {
	Id               bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Admin            bool          `json:"admin" bson:"admin"`
	UserId           string        `json:"userId" bson:"userId"`
	SessionTimeStamp int64         `json:"sessionTimestamp" bson:"sessionTimestamp"`
	Persistent       bool          `json:"persistent" bson:"persistent"`
	Token            string        `json:"token" bson:"token"`
	LoggedOut        bool          `json:"loggedOut" bson:"loggedOut"`
	ExpirationTime   int64         `json:"expirationTime" bson:"expirationTime"`
}

func (session *Session) UserIdAsBsonObject() bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(session.UserId)
	return id
}

func (session *Session) IsExpired() bool {
	if session.Persistent {
		return false
	}
	return session.ExpirationTime < time.Now().UnixMilli()
}
