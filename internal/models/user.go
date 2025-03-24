package models

import "errors"

// todo: интерфейс пользователя с ролями
type User struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"-"`
	Role       string `json:"role"`
	Department string `json:"department"`
}

// todo: имитация БД
var Users = []User{
	{ID: "1", Username: "admin", Password: "adminpassword", Role: "admin", Department: "IT"},
	{ID: "2", Username: "hrmanager", Password: "hrpassword", Role: "manager", Department: "HR"},
	{ID: "3", Username: "salesstaff", Password: "salespassword", Role: "staff", Department: "Sales"},
	{ID: "4", Username: "itstaff", Password: "itpassword", Role: "staff", Department: "IT"},
}

func GetUserByUsername(username string) (User, error) {
	for _, user := range Users {
		if user.Username == username {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}

func GetUserByID(id string) (User, error) {
	for _, user := range Users {
		if user.ID == id {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}
