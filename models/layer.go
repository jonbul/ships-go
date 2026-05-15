package models

type Layer struct {
	Name    string  `json:"name"`
	Visible bool    `json:"visible"`
	Shapes  []Shape `json:"shapes"`
}
