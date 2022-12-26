package model

type Stats struct {
	Followers int64
	Following int64
}

type Post struct {
	Id          string
	Description string
	Text        string
	Tags        []string
}
