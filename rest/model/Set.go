package model

// SetBody define the struct of the body
type SetBody struct {
	Id string `json:"id"`
}

// UpdateBody define the body struct of patch route
type UpdateBody struct {
	Public *bool `json:"public,omitempty"`
}
