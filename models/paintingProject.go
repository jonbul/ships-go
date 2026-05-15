package models

type PaintingProject struct {
	UserId       string `json:"userId"`
	Name         string `json:"name"`
	DateCreated  string `json:"dateCreated"`
	DateModified string `json:"dateModified"`
	Canvas       struct {
		Width  float32 `json:"width"`
		Height float32 `json:"height"`
	} `json:"canvas"`
	Layers []Layer `json:"layers"`
}
