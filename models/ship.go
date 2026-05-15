package models

type Ship struct {
	Name   string  `json:"name"`
	Layers []Layer `json:"layers"`
}
