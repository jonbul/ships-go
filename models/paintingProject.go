package models

import "go.mongodb.org/mongo-driver/v2/bson"

type PaintingProject struct {
	Id           bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	UserId       string        `json:"userId" bson:"userId"`
	Name         string        `json:"name" bson:"name"`
	DateCreated  int64         `json:"dateCreated" bson:"dateCreated"`
	DateModified int64         `json:"dateModified" bson:"dateModified"`
	Canvas       struct {
		Width  string `json:"width"`
		Height string `json:"height"`
	} `json:"canvas"`
	Layers []Layer `json:"layers"`
}
