package models

// UserTagWithFrequency representa un tag de usuario con su frecuencia
type UserTagWithFrequency struct {
	Tag       string
	Frequency int
}

// UserDoc representa un usuario en MongoDB
type UserDoc struct {
	UserID       int    `json:"userId"`
	UIdx         *int   `json:"uIdx,omitempty"`
	Email        string `json:"email"`
	PasswordHash string `json:"passwordHash"`
	Role         string `json:"role"`
	CreatedAt    string `json:"createdAt"`
}
