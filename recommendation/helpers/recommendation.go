package helpers

import "github.com/Gravitalia/recommendation/model"

// RemoveDuplicates allows for the removal of duplicate content,
// allowing for faster analysis
func RemoveDuplicates(list []model.Post) []model.Post {
	var newList []model.Post
	seen := make(map[string]bool)

	for _, post := range list {
		if !seen[post.Id] {
			seen[post.Id] = true
			newList = append(newList, post)
		}
	}

	return newList
}
