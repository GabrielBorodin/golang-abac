package models

import "github.com/dgrijalva/jwt-go"

// todo: 'Интерфейс' пользователя для авторизации
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}
