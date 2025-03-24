package models

import "errors"

// todo: 'Интерфейс' документов
type Document struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	Classification string `json:"classification"`
	OwnerID        string `json:"owner_id"`
}

// todo: имитация БД документов
var Documents = []Document{
	{ID: "1", Title: "Document 1", Content: "This is document 1", Classification: "public", OwnerID: "1"},
	{ID: "2", Title: "Document 2", Content: "This is document 2", Classification: "internal", OwnerID: "2"},
	{ID: "3", Title: "Document 3", Content: "This is document 3", Classification: "confidential", OwnerID: "3"},
	{ID: "4", Title: "Document 4", Content: "This is document 4", Classification: "internal", OwnerID: "4"},
}

// todo: добавление документа
func AddDocument(doc Document) {
	Documents = append(Documents, doc)
}

// todo: получение документа по ID
func GetDocumentByID(id string) (Document, error) {
	for _, doc := range Documents {
		if doc.ID == id {
			return doc, nil
		}
	}
	return Document{}, errors.New("Document not found")
}

// todo: Обновление данных о документе
func UpdateDocument(updatedDoc Document) {
	for i, doc := range Documents {
		if doc.ID == updatedDoc.ID {
			Documents[i] = updatedDoc
			return
		}
	}
}

// todo: удаление документа
func DeleteDocument(id string) {
	for i, doc := range Documents {
		if doc.ID == id {
			Documents = append(Documents[:i], Documents[i+1:]...)
			return
		}
	}
}
