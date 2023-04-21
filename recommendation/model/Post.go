package model

type Post struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Tag         string `json:"tag"`
}
