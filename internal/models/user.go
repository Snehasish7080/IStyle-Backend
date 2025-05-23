package models

type User struct {
	ID               string `json:"id"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	UserName         string `json:"userName"`
	ProfilePic       string `json:"profilePic"`
	Mobile           string `json:"mobile"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	Created_at       string `json:"created_at"`
	Updated_at       string `json:"updated_at"`
	EmailOtp         string `json:"emailOtp"`
	MobileOtp        string `json:"mobileOtp"`
	IsEmailVerified  bool   `json:"isEmailVerified"`
	IsMobileVerified bool   `json:"isMobileVerified"`
	IsComplete       bool   `json:"isComplete"`
	Bio              string `json:"bio"`
	IsFollowing      bool   `json:"isFollowing"`
}
