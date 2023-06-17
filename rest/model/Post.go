package model

// Post struct defines how post must be
type Post struct {
	Id          string `json:"id"`
	Hash        []any  `json:"hash"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Like        int64  `json:"like"`
	Author      string `json:"author"`
	Comments    []any  `json:"comments,omitempty"`
}

// PostBody defines how body when posting
// new image must be
type PostBody struct {
	Description string   `json:"description"`
	Images      [][]byte `json:"images"`
}
