package dataaccess

import (
	"context"
	"ships/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SessionDataAccessType struct {
	*baseDataAccess
}

var SessionDataAccess = SessionDataAccessType{
	baseDataAccess: &BaseDataAccess,
}

func (dataAccess SessionDataAccessType) InsertSession(session models.Session) error {
	return dataAccess.ExecuteSecurely(CollectionNames.sessions(), func(collection mongo.Collection) error {
		_, err := collection.InsertOne(context.Background(), session)
		return err
	})
}

func (dataAccess SessionDataAccessType) GetSessionByToken(token string) (*models.Session, error) {
	var session models.Session
	err := dataAccess.ExecuteSecurely(CollectionNames.sessions(), func(collection mongo.Collection) error {
		return collection.FindOne(context.Background(), bson.M{"token": token}).Decode(&session)
	})
	return &session, err
}

func (dataAccess SessionDataAccessType) UpdateSession(session *models.Session) error {
	return dataAccess.ExecuteSecurely(CollectionNames.sessions(), func(collection mongo.Collection) error {
		_, err := collection.UpdateByID(context.Background(), session.Id, bson.M{"$set": session})
		return err
	})

}
