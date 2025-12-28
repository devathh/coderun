package dto

type Token struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
