package utils

import (
	"baselix/internal/models"
)

func RemoveIDFromSchemaMap(m map[string]interface{}) map[string]interface{} {
	delete(m, "id")
	return m
}

func ConvertSchemaMapToFields(m map[string]interface{}) []models.Field {
	fields := make([]models.Field, 0, len(m))
	for name, typ := range m {
		attrType, ok := typ.(string)
		if !ok {
			continue
		}
		fields = append(fields, models.Field{
			Name: name,
			Type: attrType,
		})
	}
	return fields
}
