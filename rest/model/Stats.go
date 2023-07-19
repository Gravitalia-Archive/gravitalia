package model

// Profile struct defines user's data architecture
type Profile struct {
	Followers uint32 `json:"followers"`
	Following uint32 `json:"following"`
	Public    bool   `json:"public"`
	Suspended bool   `json:"suspended"`
	PostCount uint16 `json:"post_count"`
}
