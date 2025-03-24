package utils

import (
	"log"

	v1 "github.com/Permify/permify-go/generated/base/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

// todo: преобразует строку в protobuf-тип Any
func ConvertStringToAny(s string) *anypb.Any {
	anyValue, err := anypb.New(&v1.StringValue{Data: s})
	if err != nil {
		log.Fatalf("Failed to create Any from string: %v", err)
	}
	return anyValue
}

// todo: проверка, содержится ли строка в массиве строк
func ContainsString(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}
