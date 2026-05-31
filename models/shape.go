package models

type Shape struct {
	Desc   string `json:"desc" bson:"desc"`
	X      int    `json:"x" bson:"x"`
	Y      int    `json:"y" bson:"y"`
	Points []struct {
		X int `json:"x" bson:"x"`
		Y int `json:"y" bson:"y"`
	} `json:"points" bson:"points"`
	Width           float32 `json:"width" bson:"width"`
	Height          float32 `json:"height" bson:"height"`
	Radius          float32 `json:"radius" bson:"radius"`
	RadiusX         float32 `json:"radiusX" bson:"radiusX"`
	RadiusY         float32 `json:"radiusY" bson:"radiusY"`
	StartAngle      float32 `json:"startAngle" bson:"startAngle"`
	EndAngle        float32 `json:"endAngle" bson:"endAngle"`
	BackgroundColor string  `json:"backgroundColor" bson:"backgroundColor"`
	BorderColor     string  `json:"borderColor" bson:"borderColor"`
	Rotation        float32 `json:"rotation" bson:"rotation"`
	Src             string  `json:"src" bson:"src"`
	Name            string  `json:"name" bson:"name"`
	Mirror          bool    `json:"mirror" bson:"mirror"`
	ProjectId       string  `json:"projectId" bson:"projectId"`
}
