package models

type Metadata struct {
	SharedEncryptedAESKey string `json:"shared_encrypted_aes_key"`
	ExpiresIn             string `json:"expires_in"` // "1h", "30m"
	MaxAccess             int    `json:"max_access"` // int (Server yêu cầu số)
	Receiver              string `json:"receiver_id"`
}
