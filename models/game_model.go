package models

type Game struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Players []Player
}
