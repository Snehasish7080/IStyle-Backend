package models

type Style struct {
	Uuid       string `json:"uuid"`
	Image      string `json:"image"`
	Links      []Link `json:"links"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}
