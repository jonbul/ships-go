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

		err = cursor.All(context.TODO(), result)

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

func (dataAccess PaintingBoardDataAccessType) GetPublicShips() (*[]models.PaintingProject, error) {
	var ships = &[]models.PaintingProject{}
	err := dataAccess.ExecuteSecurely(CollectionNames.ships(), func(collection mongo.Collection) error {
		cursor, err := collection.Find(context.TODO(), bson.D{{}})
		if nil != err {
			return err
		}

		err = cursor.All(context.TODO(), ships)

		for i := range *ships {
			ship := &(*ships)[i]

			if ship.Canvas.Width == 0 {
				ship.Canvas.Width = ship.Width
			}
			if ship.Canvas.Height == 0 {
				ship.Canvas.Height = ship.Height
			}
		}
		return nil
	})
	return ships, err
}
