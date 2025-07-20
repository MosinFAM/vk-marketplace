package models

type User struct {
	ID           string `json:"id" example:"d276be1d-a6b7-4e76-9d93-2f5a9f49b6d3"`
	Username     string `json:"username" example:"johndoe"`
	PasswordHash string `json:"-"`
}
