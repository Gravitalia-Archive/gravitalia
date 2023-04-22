package model

// Post defines how a post should be sent and generated
type Post struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Tag         string `json:"tag,omitempty"`
	Like        int64  `json:"like,"`
}
