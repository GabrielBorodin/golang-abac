package middlewares

import (
	"context"
	"log"
	"net/http"

	"golang-abac-demo/internal/config"
	"golang-abac-demo/internal/controllers"
	"golang-abac-demo/internal/models"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/structpb"
)

// todo: функция проверки прав пользователя
func ABACMiddleware(permission string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//todo: получение информации о пользователе
			claims := r.Context().Value(controllers.UserKey).(*models.Claims)
			username := claims.Username
			user, err := models.GetUserByUsername(username)

			if err != nil {
				log.Printf("Failed to fetch user: %v", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			//todo: получение id документа из параметров запроса
			vars := mux.Vars(r)
			documentID := vars["id"]

			// check if document exists
			//todo: проверка документа на существование
			_, docErr := models.GetDocumentByID(documentID)

			if docErr != nil {
				log.Printf("Failed to fetch document: %v", docErr)
				http.Error(w, "Document not found", http.StatusNotFound)
				return
			}

			//todo: подготовка контекста для проверки прав доступа с передачей отдела пользователя
			data := map[string]interface{}{
				"dept": user.Department,
			}

			structData, err := structpb.NewStruct(data)

			if err != nil {
				log.Fatalf("Failed to create protobuf struct: %v", err)
			}

			cr, err := config.PermifyClient.Permission.Check(context.Background(), &v1.PermissionCheckRequest{
				TenantId: "t1", //todo: идентификатор арендатора
				Metadata: &v1.PermissionCheckRequestMetadata{
					SnapToken: config.SnapToken, //todo: токен синхронизации данных
					Depth:     50,
				},
				Entity: &v1.Entity{ //todo: сущность, для которой проверяются права (в этом случае document)
					Type: "document",
					Id:   documentID,
				},
				Permission: permission, //todo: права, которые необходимо проверить
				Subject: &v1.Subject{
					Type: "user",
					Id:   user.ID,
				},
				Context: &v1.Context{ //todo: контекст (пример: отдел пользователя)
					Data: structData,
				},
			})

			//todo: обработка результатов запроса
			if err != nil {
				log.Printf("Failed to check permission: %v", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			//todo: при разрешении доступа к ресурсу передаёт управление следующему обработчику
			if cr.Can == v1.CheckResult_CHECK_RESULT_ALLOWED {
				next.ServeHTTP(w, r)
				return
			}

			log.Printf("Permission denied")
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}
