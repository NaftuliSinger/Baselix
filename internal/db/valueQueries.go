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
func NewValue(recordID uuid.UUID, fieldID uuid.UUID, fieldType string, val any) (*models.Value, error) {
	v := &models.Value{
		RecordID: recordID,
		FieldID:  fieldID,
	}
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
		t, ok := val.(time.Time)
		if !ok {
			return nil, fmt.Errorf("expected time.Time, got %T", val)
		}
		v.ValueTime = t
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
		u, ok := val.(uuid.UUID)
		if !ok {
			return nil, fmt.Errorf("expected uuid.UUID, got %T", val)
		}
		v.ValueUUID = u
	default:
		return nil, fmt.Errorf("unknown field type: %q", fieldType)
	}
	return v, nil
}

// GetValue returns the value from the Value struct based on the field type.
func GetValue(v *models.Value, fieldType string) any {
	switch fieldType {
	case "string":
		return v.ValueString
	case "int":
		return v.ValueInt
	case "float":
		return v.ValueFloat
	case "bool":
		return v.ValueBool
	case "time":
		return v.ValueTime
	case "json":
		return v.ValueJSON
	case "uuid":
		return v.ValueUUID
	default:
		return nil
	}
}
