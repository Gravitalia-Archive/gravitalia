package model

type Stats struct {
	Followers int64 `json:"followers"`
	Following int64 `json:"following"`
}

type Post struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Like        int64  `json:"like"`
	Comments    []any  `json:"comments,omitempty"`
}
