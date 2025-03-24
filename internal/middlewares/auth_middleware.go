package middlewares

import (
	"context"
	"log"
	"net/http"
	"strings"

	"golang-abac-demo/internal/controllers"
	"golang-abac-demo/internal/models"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("my_secret_key")

// todo: проверка аутентификации пользователя и JWT-токена
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		//todo: без заголовка Authorization всегда будет сыпать ошибки
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		//извлекает токен из заголовка
		tokenString := strings.Split(authHeader, "Bearer ")[1]
		claims := &models.Claims{}

		//todo: парсинг токена по параметрам Claims
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			log.Printf("Error parsing token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		//todo: добавляет информацию о пользователе в контекст запроса
		// передаёт управление следующему обработчику
		ctx := context.WithValue(r.Context(), controllers.UserKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// todo: логирование информации о запросе
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Implementation of logging
		next.ServeHTTP(w, r)
	})
}
