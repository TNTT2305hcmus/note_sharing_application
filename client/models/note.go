package models

type Note struct {
	ID              string `json:"note_id"`
	CipherText      string `json:"cipher_text"`
	EncryptedAesKey string `json:"encrypted_aes_key"`
	OwnerID         string `json:"owner_id"`
}

type NoteData struct {
	EncryptedContent string `json:"cipher_text"` 
	EncryptedKey     string `json:"encrypted_aes_key"`
}
