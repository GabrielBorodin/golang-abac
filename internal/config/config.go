package config

import (
	"github.com/joho/godotenv" //todo: подключение библиотеки для обработки .env-файлов
	"log"
)

func LoadConfig() {
	//todo: загрузка содержимого файла .env
	err := godotenv.Load()

	log.Println("попытка получения .env", err)
	//todo: обращение к переменным можно реализовать через jwtSecret := os.Getenv("JWT_SECRET")
	if err != nil {
		log.Println("успех получения .env")
		//todo: можно прописать обработку ошибок на случай отсутствия данных в файле
	} else {
		log.Println("ошибка получения .env")
	}
}
