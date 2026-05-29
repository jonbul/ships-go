package models

import (
	"crypto/rand"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const SESSIONDURATION = 30 * 24 * 3600000

type Session struct {
	Admin            bool   `json:"admin"`
	UserId           string `json:"userId"`
	SessionTimeStamp int64  `json:"sessionTimestamp"`
	Persistent       bool   `json:"persistent"`
	Token            string `json:"token"`
	LoggedOut        bool   `json:"loggedOut"`
	ExpirationTime   int64  `json:"expirationTime"`
}

func (session *Session) UserIdAsBsonObject() bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(session.UserId)
	return id
}

func (s Session) IsExpired() bool {
	if s.Persistent {
		return false
	}
	return s.ExpirationTime < time.Now().UnixMilli()
}

func NewSession(userId string, admin bool, persistent bool) Session {
	session := Session{
		Admin:            admin,
		UserId:           userId,
		SessionTimeStamp: time.Now().UnixMilli(),
		Persistent:       persistent,
		Token:            getNewBearerToken(),
		LoggedOut:        false,
	}
	if persistent {
		session.ExpirationTime = time.Now().Add(365 * 24 * time.Hour).UnixMilli()
	} else {
		session.ExpirationTime = time.Now().UnixMilli() + SESSIONDURATION
	}
	return session
}

func refeshToken(session *Session) string {
	if !session.Persistent && !session.IsExpired() {
		session.Token = getNewBearerToken()
		session.ExpirationTime = time.Now().UnixMilli() + SESSIONDURATION
		// TODO save in mongo
	}
	return session.Token
}

func getNewBearerToken() string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, 60)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}
