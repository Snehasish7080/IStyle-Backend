package models

type Link struct {
	Url   string `json:"url"`
	Image string `json:"image"`
}
type Style struct {
	Image      string `json:"image"`
	Links      []Link `json:"links"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}
