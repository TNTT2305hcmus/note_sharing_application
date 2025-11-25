package models

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	PublicKey string `json:"public_key"`
}

type RegisterRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	PublicKey string `json: "public_key"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
