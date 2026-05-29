package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Ship struct {
	Id     bson.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name   string        `json:"name"`
	Layers []Layer       `json:"layers"`
}
