package models

type Shape struct {
	Desc   string `json:"desc"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Points []struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"points"`
	Width           float32 `json:"width"`
	Height          float32 `json:"height"`
	Radius          float32 `json:"radius"`
	RadiusX         float32 `json:"radiusX"`
	RadiusY         float32 `json:"radiusY"`
	StartAngle      float32 `json:"startAngle"`
	EndAngle        float32 `json:"endAngle"`
	BackgroundColor string  `json:"backgroundColor"`
	BorderColor     string  `json:"borderColor"`
	Rotation        float32 `json:"rotation"`
	Src             string  `json:"src"`
	Name            string  `json:"name"`
	Mirror          bool    `json:"mirror"`
	ProjectId       string  `json:"projectId"`
}
