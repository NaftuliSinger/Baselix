package utils

import (
	"baselix/internal/db"
	"baselix/internal/models"
	"baselix/internal/types"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func LowercaseSchemaMapValues(m map[string]interface{}) (map[string]interface{}, error) {
	lowercased := make(map[string]interface{}, len(m))
	for key, value := range m {
		strValue, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("value for key %q is not a string", key)
		}
		lowercased[key] = strings.ToLower(strValue)
	}
	return lowercased, nil
}

func CheckForReservedFields(m map[string]interface{}) error {
	/* if the map has any of the following, return an error
	- "id"
	- "created_at"
	- "updated_at"
	*/

	switch {
	case m["id"] != nil:
		// return custom error
		return &types.ResrvedFieldError{FieldName: "id"}
	case m["created_at"] != nil:
		return &types.ResrvedFieldError{FieldName: "created_at"}
	case m["updated_at"] != nil:
		return &types.ResrvedFieldError{FieldName: "updated_at"}
	default:
		return nil
	}

}

func IsValidFieldType(fieldType string) bool {
	switch fieldType {
	case "string", "int", "float", "bool", "time", "json", "uuid":
		return true
	case "string_u", "int_u", "float_u", "uuid_u":
		return true
	default:
		return false
	}
}

func isUniqueFieldType(fieldName string) bool {
	// if has suffix "_u"
	return strings.HasSuffix(fieldName, "_u")
}

func StripUniqueSuffix(fieldName string) string {
	return strings.TrimSuffix(fieldName, "_u")
}

func ConvertSchemaMapToFields(m map[string]interface{}) ([]models.Field, error) {
	// lowercase the map
	lowercased, err := LowercaseSchemaMapValues(m)
	if err != nil {
		return nil, err
	}
	m = lowercased

	// check for reserved field names
	err = CheckForReservedFields(m)
	if err != nil {
		return nil, err
	}

	fields := make([]models.Field, 0, len(m))
	for name, typ := range m {
		// validate the type is one of the allowed types, if not return an error
		if !IsValidFieldType(typ.(string)) {
			return nil, fmt.Errorf("invalid field type for field %q: %q", name, typ)
		}

		// set unique to true if the type has a "_u" suffix, and remove the suffix from the type
		unique := isUniqueFieldType(typ.(string))
		if unique {
			typ = StripUniqueSuffix(typ.(string))
		}

		attrType, ok := typ.(string)
		if !ok {
			continue
		}
		fields = append(fields, models.Field{
			Name:   name,
			Type:   attrType,
			Unique: unique,
		})
	}
	return fields, nil
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

	case int, int8, int16, int32, int64, float32, float64:
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

func MapRecordModelToRecordResponse(record *models.Record) types.RecordResponse {
	values := make([]types.RecordField, 0, len(record.Values))

	for _, v := range record.Values {
		raw := db.GetValue(v, v.Field.Type, v.Field.Unique)

		// If field type is JSON, unmarshal string into map
		if v.Field.Type == "json" {
			if s, ok := raw.(string); ok && s != "" {
				var parsed map[string]any
				if err := json.Unmarshal([]byte(s), &parsed); err == nil {
					raw = parsed
				}
			}
		}

		values = append(values, types.RecordField{Key: v.Field.Name, Value: raw})
	}

	return types.RecordResponse{
		ID:        record.ID.String(),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		Values:    values,
	}
}
