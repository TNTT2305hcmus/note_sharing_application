package models

type Metadata struct {
	ExpiresIn string `json:"expires_in"` // "1h", "30m"
	MaxAccess int    `json:"max_access"` // int (Server yêu cầu số)
}
