package model

// Profile struct defines user's data architecture
type Profile struct {
	Followers int64 `json:"followers"`
	Following int64 `json:"following"`
	Public    bool  `json:"public"`
	Suspended bool  `json:"suspended"`
}
