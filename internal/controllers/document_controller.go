package controllers

import (
	"context"
	"encoding/json"
	"golang-abac-demo/internal/config"
	"golang-abac-demo/internal/models"
	"golang-abac-demo/internal/utils"
	"log"
	"net/http"
	"strconv"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/gorilla/mux"
)

// todo: функция обработки запроса на загрузку документа
func UploadDocument(w http.ResponseWriter, r *http.Request) {
	var doc models.Document
	err := json.NewDecoder(r.Body).Decode(&doc)
	//todo: декодирование тела POST-запроса с проверкой на соответствие полей типу Document
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//todo: декодирование данных о пользователе через декодирование JWT-ключа
	claims := r.Context().Value(UserKey).(*models.Claims)
	//todo: получение пользователя по имени пользователя
	user, err := models.GetUserByUsername(claims.Username)
	if err != nil {
		log.Printf("Failed to fetch user: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	//todo: установка владельца файла
	doc.OwnerID = user.ID
	noOfDocs := len(models.Documents)
	//todo: генерация уникального идентификатора документа на основе существующих ранее
	if noOfDocs == 0 {
		doc.ID = "1"
	} else {
		lastDocIndex := noOfDocs - 1
		lastDocID, err := strconv.Atoi(models.Documents[lastDocIndex].ID)

		if err != nil {
			log.Printf("Failed to fetch last document ID: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		//todo: если документов нет, то установит '1', иначе - 'last+1'
		doc.ID = strconv.Itoa(lastDocID + 1)
	}

	// Add document to repository (this would be replaced with actual DB call)
	//todo: добавляет документ в хранилище (+1 объект в массив Document)
	models.AddDocument(doc)

	// Add document to Permify
	//todo: добавление документа в permify
	//создание отношений для документа
	tuples := []*v1.Tuple{{
		Entity: &v1.Entity{
			Type: "document",
			Id:   doc.ID,
		},
		Relation: "owner",
		Subject: &v1.Subject{
			Type: "user",
			Id:   user.ID,
		},
	}}
	//создание атрибутов для документа
	attributes := []*v1.Attribute{
		{
			Entity: &v1.Entity{
				Type: "document",
				Id:   doc.ID,
			},
			Attribute: "classification",
			Value:     utils.ConvertStringToAny(doc.Classification), //классификация документа
		},
		{
			Entity: &v1.Entity{
				Type: "document",
				Id:   doc.ID,
			},
			Attribute: "department",
			Value:     utils.ConvertStringToAny(user.Department), //отдел документа
		},
	}

	_, err = config.PermifyClient.Data.Write(context.Background(), &v1.DataWriteRequest{
		TenantId: "t1",
		Metadata: &v1.DataWriteRequestMetadata{
			SchemaVersion: config.SchemaVersion,
		},
		Tuples:     tuples,
		Attributes: attributes,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to write relationship to Permify"})
		return
	}

	utils.InfoLogger.Printf("User '%s' uploaded document %s", user.Username, doc.ID)
	//todo: фиксируем корректный ответ сервера на загрузку/поломку
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Document uploaded successfully"})
}

// todo: просмотр существующего документа
func ViewDocument(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	//todo: получение id загружаемого файла
	docID := params["id"]
	claims := r.Context().Value(UserKey).(*models.Claims)

	// Fetch document from repository (this would be replaced with actual DB call)
	//todo: обращение в БД для поиска документа по идентификатору
	doc, err := models.GetDocumentByID(docID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}
	//todo: логирование пользователя
	utils.InfoLogger.Printf("User '%s' viewed document %s", claims.Username, doc.ID)

	//todo: возвращает успешное состояние документа
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(doc)
}

// todo: изменение документа по идентификатору
func EditDocument(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	docID := params["id"]
	claims := r.Context().Value(UserKey).(*models.Claims)

	//todo: превращение тела запроса в структуру Document
	var doc models.Document
	err := json.NewDecoder(r.Body).Decode(&doc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Fetch document from repository (this would be replaced with actual DB call)
	//todo: получение документа по идентификатору
	existingDoc, err := models.GetDocumentByID(docID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Update document in repository (this would be replaced with actual DB call)
	//todo: обновление параметров документа, исходя из параметров запроса
	existingDoc.Title = doc.Title
	existingDoc.Content = doc.Content
	models.UpdateDocument(existingDoc)

	utils.InfoLogger.Printf("User '%s' edited document %s", claims.Username, doc.ID)
	//todo: логирование и отправка ответа на сервер
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Document edited successfully"})
}

// todo: удаление документа
func DeleteDocument(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	docID := params["id"]
	claims := r.Context().Value(UserKey).(*models.Claims)

	// Fetch document from repository (this would be replaced with actual DB call)
	//todo: получение документа по ID
	doc, err := models.GetDocumentByID(docID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	// Delete document from repository (this would be replaced with actual DB call)
	//todo: удаление документа по ID
	models.DeleteDocument(docID)

	// Delete document from Permify
	//todo: удаление документа из permify
	_, err = config.PermifyClient.Data.Delete(context.Background(), &v1.DataDeleteRequest{
		TenantId: "t1",
		TupleFilter: &v1.TupleFilter{
			Entity: &v1.EntityFilter{
				Type: "document",
				Ids:  []string{doc.ID},
			},
		},
		AttributeFilter: &v1.AttributeFilter{
			Entity: &v1.EntityFilter{
				Type: "document",
				Ids:  []string{doc.ID},
			},
			Attributes: []string{"classification", "department"},
		},
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete document from Permify"})
		return
	}

	utils.InfoLogger.Printf("User '%s' deleted document %s", claims.Username, doc.ID)
	//todo: логирование итога действий
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Document deleted successfully"})
}
