package dataaccess

import (
	"context"
	"ships/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InsertSession(session models.Session) error {
	return ExecuteSecurely(CollectionNames.sessions(), func(collection mongo.Collection) error {
		_, err := collection.InsertOne(context.Background(), session)
		return err
	})
}

func GetSessionByToken(token string) (*models.Session, error) {
	var session models.Session
	err := ExecuteSecurely(CollectionNames.sessions(), func(collection mongo.Collection) error {
		return collection.FindOne(context.Background(), bson.M{"token": token}).Decode(&session)
	})
	return &session, err
}
