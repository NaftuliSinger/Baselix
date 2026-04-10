package db

import (
	"baselix/internal/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewValue creates a Value with only the relevant typed field set.
// fieldType must match Field.Type ("string", "int", "float", "bool", "time", "json", "uuid").
func NewValue(tableID uuid.UUID, recordID uuid.UUID, fieldID uuid.UUID, fieldType string, unique bool, val any) (*models.Value, error) {
	v := &models.Value{
		TableID:  tableID,
		RecordID: recordID,
		FieldID:  fieldID,
	}
	// if the field is unique, set the appropriate unique value field based on the field type (string, int, float, uuid)
	if unique {
		switch strings.ToLower(fieldType) {
		case "string":
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("expected string, got %T", val)
			}
			v.UniqueValueString = s
		case "int":
			var i int
			switch n := val.(type) {
			case int:
				i = n
			case float64:
				i = int(n)
			default:
				return nil, fmt.Errorf("expected int, got %T", val)
			}
			v.UniqueValueInt = i
		case "float":
			f, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("expected float64, got %T", val)
			}
			v.UniqueValueFloat = f
		case "uuid":
			switch uv := val.(type) {
			case uuid.UUID:
				v.UniqueValueUUID = uv
			case string:
				u, err := uuid.Parse(uv)
				if err != nil {
					return nil, fmt.Errorf("expected UUID string, got %q: %w", uv, err)
				}
				v.UniqueValueUUID = u
			default:
				return nil, fmt.Errorf("expected uuid.UUID or string, got %T", val)
			}
		default:
			return nil, fmt.Errorf("unknown field type: %q", fieldType)
		}
	} else {
		// non-unique fields use the regular value fields (all types)
		switch strings.ToLower(fieldType) {
		case "string":
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("expected string, got %T", val)
			}
			v.ValueString = s
		case "int":
			var i int
			switch n := val.(type) {
			case int:
				i = n
			case float64:
				i = int(n)
			default:
				return nil, fmt.Errorf("expected int, got %T", val)
			}
			v.ValueInt = i
		case "float":
			f, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("expected float64, got %T", val)
			}
			v.ValueFloat = f
		case "bool":
			b, ok := val.(bool)
			if !ok {
				return nil, fmt.Errorf("expected bool, got %T", val)
			}
			v.ValueBool = b
		case "time":
			switch tv := val.(type) {
			case time.Time:
				v.ValueTime = tv
			case string:
				t, err := time.Parse(time.RFC3339, tv)
				if err != nil {
					return nil, fmt.Errorf("expected ISO 8601 timestamp, got %q: %w", tv, err)
				}
				v.ValueTime = t
			default:
				return nil, fmt.Errorf("expected time.Time or string, got %T", val)
			}
		case "json":
			var s string
			switch jv := val.(type) {
			case string:
				s = jv
			default:
				b, err := json.Marshal(jv)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal json value: %w", err)
				}
				s = string(b)
			}
			v.ValueJSON = s
		case "uuid":
			switch uv := val.(type) {
			case uuid.UUID:
				v.ValueUUID = uv
			case string:
				u, err := uuid.Parse(uv)
				if err != nil {
					return nil, fmt.Errorf("expected UUID string, got %q: %w", uv, err)
				}
				v.ValueUUID = u
			default:
				return nil, fmt.Errorf("expected uuid.UUID or string, got %T", val)
			}
		default:
			return nil, fmt.Errorf("unknown field type: %q", fieldType)
		}
	}
	return v, nil
}

// GetValue returns the value from the Value struct based on the field type.
func GetValue(v *models.Value, fieldType string, unique bool) any {
	switch fieldType {
	case "string":
		if unique {
			return v.UniqueValueString
		} else {
			return v.ValueString
		}
	case "int":
		if unique {
			return v.UniqueValueInt
		} else {
			return v.ValueInt
		}
	case "float":
		if unique {
			return v.UniqueValueFloat
		} else {
			return v.ValueFloat
		}
	case "bool":
		return v.ValueBool
	case "time":
		return v.ValueTime
	case "json":
		return v.ValueJSON
	case "uuid":
		if unique {
			return v.UniqueValueUUID
		} else {
			return v.ValueUUID
		}
	default:
		return nil
	}
}
