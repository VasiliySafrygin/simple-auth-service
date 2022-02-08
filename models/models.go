package models

type Group struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Id       string `json:"id" field:"id"`
	UserName     string `json:"username" field:"username"`
	FirstName     string `json:"first_name" field:"first_name"`
	LastName     string `json:"last_name" field:"last_name"`
	MiddleName     string `json:"middle_name" field:"middle_name"`
	Password string `json:"password" field:"password"`
}
