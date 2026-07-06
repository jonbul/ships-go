package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PaintingProject struct {
	Id           bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	UserId       string        `json:"userId" bson:"userId"`
	Name         string        `json:"name" bson:"name"`
	DateCreated  int64         `json:"dateCreated" bson:"dateCreated"`
	DateModified int64         `json:"dateModified" bson:"dateModified"`
	Width        int           `json:"width" bson:"width"`
	Height       int           `json:"height" bson:"height"`
	Canvas       struct {
		Width  int `json:"width" bson:"width"`
		Height int `json:"height" bson:"height"`
	} `json:"canvas" bson:"canvas"`
	Layers []Layer `json:"layers" bson:"layers"`
}

func (project *PaintingProject) init() {
	if project.Width != 0 && project.Canvas.Width == 0 {
		project.Canvas.Width = project.Width
	}
	if project.Height != 0 && project.Canvas.Height == 0 {
		project.Canvas.Height = project.Height
	}
}
