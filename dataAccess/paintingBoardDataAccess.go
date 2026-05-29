package dataaccess

import (
	"context"
	"ships/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PaintingBoardDataAccessType struct {
	*baseDataAccess
}

var PaintingBoardDataAccess = PaintingBoardDataAccessType{
	baseDataAccess: &BaseDataAccess,
}

func (dataAccess PaintingBoardDataAccessType) GetProjectsByUserId(userId string) (*[]models.PaintingProject, error) {
	var result = &[]models.PaintingProject{}
	err := dataAccess.ExecuteSecurely(CollectionNames.paintingProjects(), func(collection mongo.Collection) error {
		cursor, err := collection.Find(context.TODO(), bson.D{{Key: "userId", Value: userId}})
		if nil != err {
			return err
		}

		for cursor.Next(context.TODO()) {
			var project models.PaintingProject
			err := cursor.Decode(&project)
			if nil != err {
				return err
			}
			*result = append(*result, project)
		}

		return err
	})
	return result, err
}

func (dataAccess PaintingBoardDataAccessType) SaveProject(project *models.PaintingProject) error {
	return dataAccess.ExecuteSecurely(CollectionNames.paintingProjects(), func(collection mongo.Collection) error {
		res, err := collection.InsertOne(context.TODO(), project)
		if nil != res {
			project.Id, err = bson.ObjectIDFromHex(res.InsertedID.(bson.ObjectID).Hex())
		}
		return err
	})
}

func (dataAccess PaintingBoardDataAccessType) UpdateProject(project *models.PaintingProject) error {
	return dataAccess.ExecuteSecurely(CollectionNames.paintingProjects(), func(collection mongo.Collection) error {
		_, err := collection.UpdateByID(context.TODO(), project.Id, bson.M{"$set": project})
		return err
	})
}

func (dataAccess PaintingBoardDataAccessType) GetProjectById(id bson.ObjectID) (*models.PaintingProject, error) {
	var result *models.PaintingProject
	err := dataAccess.ExecuteSecurely(CollectionNames.paintingProjects(), func(collection mongo.Collection) error {
		return collection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: id}}).Decode(&result)
	})
	return result, err
}

func (dataAccess PaintingBoardDataAccessType) DeleteProjectById(id bson.ObjectID) error {
	err := dataAccess.ExecuteSecurely(CollectionNames.paintingProjects(), func(collection mongo.Collection) error {
		_, err := collection.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: id}})
		return err
	})
	return err
}
