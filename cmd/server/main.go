package main

// todo: Файл, который является точкой входа в приложение
// Пакет для загрузки конфигурации и инициализации Permify
// Пакет, содержащий обработчики HTTP-запросов (например, для аутентификации и управления документами)
//  Пакет, содержащий middleware для аутентификации, авторизации и логирования.
// Стандартный пакет для логирования.
// Стандартный пакет для работы с HTTP
// Внешний пакет для маршрутизации HTTP-запросов.
import (
	"github.com/gorilla/mux"
	"golang-abac-demo/internal/config"
	"golang-abac-demo/internal/controllers"
	"golang-abac-demo/internal/middlewares"
	"log"
	"net/http"
)

// todo: выполняется при входе в программу
func main() {
	// Load configuration
	config.LoadConfig() //todo: вызов ф-ии LoadConfig из файла config.go

	// Initialize Permify client
	config.InitPermifyClient() //todo: функция инициализирует клиент для взаимодействия с Permify

	// Write Permify schema
	//todo: отправляет схему ABAC (Attribute-Based Access Control) в Permify
	//Схема определяет сущности, атрибуты и правила доступа
	config.WritePermifySchema() //

	// Sync Permify with the database
	//todo: синхронизирует данные между вашим приложением и Permify
	// удаляет устаревшие записи и добавляет новые записи
	config.SyncPermify()

	// Initialize the router
	//todo: Создает новый маршрутизатор с помощью пакета gorilla/mux (обработка HTTP-запросов)
	r := mux.NewRouter()

	// Log all requests
	//todo: Регистрирует middleware для логирования всех входящих HTTP-запросов
	//Middleware — это функция, которая выполняется перед основным обработчиком запроса
	r.Use(middlewares.LoggingMiddleware)

	// Public Routes
	//todo: определение маршрута для авторизации пользователей
	//Обработчик controllers.Login вызывается при POST-запросе на /login
	r.HandleFunc("/login", controllers.Login).Methods("POST")

	// Private Routes
	//todo: создание подмаршрутизатора для всех запросов, начинающихся с 'api'
	api := r.PathPrefix("/api").Subrouter()
	//todo: Регистрирует middleware для аутентификации (с проверкой JWT-токена)
	api.Use(middlewares.AuthMiddleware)
	//todo: загрузка документа, обработчик: UploadDocument
	api.HandleFunc("/documents", controllers.UploadDocument).Methods("POST")
	//todo: просмотр документа, использование ABACMiddleware("view") для просмотра прав доступа
	api.Handle("/documents/{id}", middlewares.ABACMiddleware("view")(http.HandlerFunc(controllers.ViewDocument))).Methods("GET")
	//todo: редактирование документа, используется ABACMiddleware("edit")
	api.Handle("/documents/{id}", middlewares.ABACMiddleware("edit")(http.HandlerFunc(controllers.EditDocument))).Methods("PUT")
	//todo: удаление документа, используется ABACMiddleware("delete")
	api.Handle("/documents/{id}", middlewares.ABACMiddleware("delete")(http.HandlerFunc(controllers.DeleteDocument))).Methods("DELETE")

	// Start the server
	//todo: запуск сервера на порту 8080 с маршрутизатором r для обработки запросов
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
