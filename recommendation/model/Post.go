package model

// Post defines how a post should be sent and generated
type Post struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Author      string `json:"author,"`
	Tag         string `json:"tag,omitempty"`
	Hash        []any  `json:"hash,"`
	Like        int64  `json:"like,"`
	MeLiked     bool   `json:"me_liked,"`
}
