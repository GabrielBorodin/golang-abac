package controllers

//todo: контроллер аутентификации пользователей
import (
	"encoding/json"
	"net/http"
	"time"

	"golang-abac-demo/internal/models"
	"golang-abac-demo/internal/utils"

	"github.com/dgrijalva/jwt-go"
)

type contextKey string

var JwtKey = []byte("my_secret_key") //todo: ключ из env

const UserKey contextKey = "user" //todo: ключ хранения информации о пользователе в контексте запроса
// todo: параметры из POST-запроса
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// todo: функция обработки входа в систему
func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	//todo: декодирование тела запроса и проверка его на соответствие интерфейсу Credentials
	if err != nil {
		http.Error(w, "Некорректные параметры запроса", http.StatusBadRequest)
		return
	}

	// This should be replaced with a call to your user repository
	//todo: если такой пользователь найден (логин) в БД, то возвращает ошибку 'Не авторизован'
	user, err := models.GetUserByUsername(creds.Username)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
		return
	}
	//todo: та же проверка на совпадение пароля
	if user.Password != creds.Password {
		http.Error(w, "Некорректный пароль", http.StatusUnauthorized)
		return
	}

	//todo: устанавливает время жизни токена в 60 минут
	expirationTime := time.Now().Add(60 * time.Minute)
	claims := &models.Claims{
		Username: user.Username,
		Role:     user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	//todo: генерирует ключ по входным параметрам и ключу JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		http.Error(w, "Возникла ошибка при авторизации", http.StatusInternalServerError)
		return
	}

	//todo: сохранение лога и отправка данных пользователю
	utils.InfoLogger.Printf("User '%s' logged in", user.Username)

	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful", "token": tokenString})
}

//todo: буквально происходит сопоставление логина и пароля пользователя с данными из БД
//при успешном сопоставлении данных на фронт кидается jwt-токен сроком жизни в 60 минут
//для работы в рамках сессии пользователя
