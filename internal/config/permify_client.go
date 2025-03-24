package config

import (
	"context"
	"golang-abac-demo/internal/models"
	"golang-abac-demo/internal/utils"
	"log"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//todo: permify - система управления доступом на основе атрибутов
//todo: файл отвечает за работу с permify, его инициализацию запись схемы доступа и синхронизации

var PermifyClient *permify.Client //todo: клиент для взаимодействия с permify
var SchemaVersion string          //todo:версия схемы для записи в permify
var SnapToken string              //todo: токен синхронизации данных в permify

func InitPermifyClient() {
	client, err := permify.NewClient(
		permify.Config{
			Endpoint: "localhost:3478", //todo: указывает адрес и порт сервера permify
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()), //отключение шифрования
	)
	if err != nil {
		log.Fatalf("Failed to initialize Permify client: %v", err)
	} else {
		PermifyClient = client
		log.Println("Permify client initialized successfully")
	}
}

func WritePermifySchema() {
	// Write schema
	//todo: определение схем сущностей, определение между ними отношений и определяет к ним разрешения
	schema := `
    entity user {}

    entity department {
        relation member @user
    }

    entity classification {
        relation public @user
        relation internal @user
        relation confidential @user
    }

    entity document {
        relation owner @user
        relation department @department
        relation classification @classification
             
        permission view = owner or (classification.internal and department.member) or classification.public
        permission edit = owner or (classification.internal and department.member)
        permission delete = owner
    }
`

	sr, err := PermifyClient.Schema.Write(context.Background(), &v1.SchemaWriteRequest{
		TenantId: "t1",
		Schema:   schema,
	})

	if err != nil {
		log.Fatalf("Failed to write schema: %v", err)
	}

	SchemaVersion = sr.SchemaVersion
	log.Printf("Schema version %s written successfully", SchemaVersion)
}

// todo; синхронизация локальной БД и permify
func SyncPermify() {
	// Read current relationships from Permify
	rr, err := PermifyClient.Data.ReadRelationships(context.Background(), &v1.RelationshipReadRequest{
		TenantId: "t1",
		Metadata: &v1.RelationshipReadRequestMetadata{
			SnapToken: SnapToken, //токен синхронизации состояния данных
		},
		//todo: ограничивает выборку сущностями типов document и user
		Filter: &v1.TupleFilter{
			Entity: &v1.EntityFilter{
				Type: "document",
			},
			Subject: &v1.SubjectFilter{
				Type: "user",
			},
		},
	})

	if err != nil {
		log.Fatalf("Failed to read relationships from Permify: %v", err)
	}

	// Map of existing document IDs in Permify
	existingDocumentIDs := make([]string, 0)
	nonExistingDocumentIDs := make([]string, 0)
	//todo: проверка сущестсвования документов в локальной БД
	for _, tuple := range rr.Tuples {
		if tuple.Entity.Type == "document" {
			_, err := models.GetDocumentByID(tuple.Entity.Id)

			if err != nil {
				nonExistingDocumentIDs = append(nonExistingDocumentIDs, tuple.Entity.Id)
			} else {
				existingDocumentIDs = append(existingDocumentIDs, tuple.Entity.Id)
			}
		}
	}

	// Delete documents that don't exist in the database
	//todo: удаление документов, которые отсутствуют в БД
	if len(nonExistingDocumentIDs) > 0 {
		rr, err := PermifyClient.Data.Delete(context.Background(), &v1.DataDeleteRequest{
			TenantId: "t1",
			TupleFilter: &v1.TupleFilter{
				Entity: &v1.EntityFilter{
					Type: "document",
					Ids:  nonExistingDocumentIDs,
				},
			},
			AttributeFilter: &v1.AttributeFilter{
				Entity: &v1.EntityFilter{
					Type: "document",
					Ids:  nonExistingDocumentIDs,
				},
				Attributes: []string{"classification", "department"},
			},
		})

		if err != nil {
			log.Fatalf("Failed to delete orphaned documents from Permify: %v", err)
		}

		//todo: обновляет токен после удаления
		SnapToken = rr.SnapToken
		log.Printf("Orphaned documents deleted from Permify successfully\nSnap token: %s", SnapToken)

	} else {
		log.Println("No orphaned documents to delete from Permify")
	}

	// Add missing documents to Permify
	//todo: добавляет отсутствующие документы в Permify, создавая им атрибуты и отношения
	var tuples []*v1.Tuple
	var attributes []*v1.Attribute

	for _, doc := range models.Documents {
		if !utils.ContainsString(existingDocumentIDs, doc.ID) {
			tuples = append(tuples, &v1.Tuple{
				Entity: &v1.Entity{
					Type: "document",
					Id:   doc.ID,
				},
				Relation: "owner",
				Subject: &v1.Subject{
					Type: "user",
					Id:   doc.OwnerID,
				},
			})

			user, err := models.GetUserByID(doc.OwnerID)

			if err != nil {
				log.Fatalf("Failed to fetch user by ID: %v", err)
			}

			attributes = append(attributes, []*v1.Attribute{
				{
					Entity: &v1.Entity{
						Type: "document",
						Id:   doc.ID,
					},
					Attribute: "classification",
					Value:     utils.ConvertStringToAny(doc.Classification),
				},
				{
					Entity: &v1.Entity{
						Type: "document",
						Id:   doc.ID,
					},
					Attribute: "department",
					Value:     utils.ConvertStringToAny(user.Department),
				},
			}...)
		}
	}

	// Write missing documents to Permify
	if len(tuples) > 0 {
		rr, err := PermifyClient.Data.Write(context.Background(), &v1.DataWriteRequest{
			TenantId: "t1",
			Metadata: &v1.DataWriteRequestMetadata{
				SchemaVersion: SchemaVersion,
			},
			Tuples:     tuples,
			Attributes: attributes,
		})

		if err != nil {
			log.Fatalf("Failed to write missing documents to Permify: %v", err)
		}

		SnapToken = rr.SnapToken
		log.Printf("Missing documents written successfully\nSnap token: %s", SnapToken)
	} else {
		log.Println("No missing documents to write to Permify")
	}

	log.Println("Permify synced successfully")
}

//todo: резюмируя, файл инициализирует permify, записывает схему доступа, осуществляет синхронизацию данных
