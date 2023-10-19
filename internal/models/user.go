package models

type User struct {
	ID         string `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	UserName   string `json:"userName"`
	ProfilePic string `json:"profilePic"`
	Mobile     string `json:"mobile"`
	Email      string `json:"email"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
	Otp        string `json:"otp"`
	IsVerified bool   `json:"isVerified"`
	IsComplete bool   `json:"isComplete"`
}
