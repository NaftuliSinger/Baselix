package utils

import (
	"baselix/internal/models"
	"time"

	"github.com/google/uuid"
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

func InferTypeFromValue(value any) string {
	// In our case we treat all numbers as float for simplicity
	switch value.(type) {
	case string:
		return "string"
	case float32, float64, int, int8, int16, int32, int64:
		return "float"
	case bool:
		return "bool"
	case time.Time:
		return "time"
	case map[string]any:
		return "json"
	case uuid.UUID:
		return "uuid"
	default:
		return "unknown"
	}
}

// InferSchemaFromRecords infers a unified schema from multiple records by merging all fields.
func InferSchemaFromRecords(records []map[string]any) map[string]interface{} {
	schema := make(map[string]interface{})
	for _, data := range records {
		for key, value := range data {
			if _, exists := schema[key]; !exists {
				valueType := InferTypeFromValue(value)
				if valueType == "unknown" {
					continue // skip fields with unknown types
				}
				schema[key] = valueType
			}
		}
	}
	return schema
}
