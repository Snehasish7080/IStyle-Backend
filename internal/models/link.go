package models

type Link struct {
	Uuid       string `json:"uuid"`
	Url        string `json:"url"`
	Image      string `json:"image"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}
