package model

type Profile struct {
	Followers int64 `json:"followers"`
	Following int64 `json:"following"`
	Public    bool  `json:"public"`
	Suspended bool  `json:"suspended"`
}

type Post struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Like        int64  `json:"like"`
	Comments    []any  `json:"comments,omitempty"`
}
