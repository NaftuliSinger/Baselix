package utils

import (
	"baselix/internal/db"
	"baselix/internal/models"
	"baselix/internal/types"
	"encoding/json"
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
	switch v := value.(type) {
	case string:
		// Try UUID first
		if _, err := uuid.Parse(v); err == nil {
			return "uuid"
		}

		// Try time in RFC3339 (ISO 8601) format with timezone
		if _, err := time.Parse(time.RFC3339, v); err == nil {
			return "time"
		}

		// Otherwise treat as string
		return "string"

	case float32, float64, int, int8, int16, int32, int64:
		return "float" // your existing choice to treat all numbers as float

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
	// clean the schema by removing any "id" fields, since we don't want to create a field for "id" in the table
	cleanedRecords := make([]map[string]any, len(records))
	for i, record := range records {
		cleanedRecords[i] = RemoveIDFromSchemaMap(record)
	}

	schema := make(map[string]interface{})
	for _, data := range cleanedRecords {
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

func MapRecordModelToRecordResponse(record *models.Record) types.RecordResponse {
	values := make(map[string]any, len(record.Values))

	for _, v := range record.Values {
		raw := db.GetValue(v, v.Field.Type)

		// If field type is JSON, unmarshal string into map
		if v.Field.Type == "json" {
			if s, ok := raw.(string); ok && s != "" {
				var parsed map[string]any
				if err := json.Unmarshal([]byte(s), &parsed); err == nil {
					raw = parsed
				}
			}
		}

		values[v.Field.Name] = raw
	}

	return types.RecordResponse{
		ID:     record.ID.String(),
		Values: values, // will be flattened when marshaled
	}
}
