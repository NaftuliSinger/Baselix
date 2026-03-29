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
	switch value.(type) {
	case string:
		return "string"
	case int, int8, int16, int32, int64:
		return "int"
	case float32, float64:
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

// Take in raw json from new records and infer the schema for the table, returning a map of field name to field type. This is used when inserting new records to automatically create fields for any new attributes that are not already defined in the table's schema.
func InferSchemaFromRecordData(data map[string]any) map[string]interface{} {
	schema := make(map[string]interface{}, len(data))
	for key, value := range data {
		schema[key] = InferTypeFromValue(value)
	}
	return schema
}
