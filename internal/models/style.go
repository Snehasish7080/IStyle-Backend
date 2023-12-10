package models

type Style struct {
	ID         string `json:"id"`
	Uuid       string `json:"uuid"`
	Image      string `json:"image"`
	Links      []Link `json:"links"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}
