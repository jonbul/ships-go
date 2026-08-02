package models

type Layer struct {
	Name    string  `json:"name" bson:"name"`
	Visible bool    `json:"visible" bson:"visible"`
	Shapes  []Shape `json:"shapes" bson:"shapes"`
}
