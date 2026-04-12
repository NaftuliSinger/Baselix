package db

import (
	"baselix/internal/models"
	"baselix/internal/types"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func FloatNoDecmimalToInt(f float64) (int, error) {
	if f == float64(int(f)) {
		return int(f), nil
	} else {
		return 0, fmt.Errorf("float value %v has decimal part, cannot convert to int", f)
	}
}

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
				return nil, &types.WrongFieldTypeError{FieldName: "string unique field", ExpectedType: "string", ActualType: fmt.Sprintf("%T", val)}
			}
			v.UniqueValueString = s
		case "int":
			i, ok := val.(int)
			if !ok {
				// also allow float values that are whole numbers for convenience, e.g. 1.0 -> 1
				if f, ok := val.(float64); ok {
					if i, err := FloatNoDecmimalToInt(f); err == nil {
						v.UniqueValueInt = i
						break
					}
				}
				return nil, &types.WrongFieldTypeError{FieldName: "int unique field", ExpectedType: "int", ActualType: fmt.Sprintf("%T", val)}
			}
			v.UniqueValueInt = i
		case "float":
			f, ok := val.(float64)
			if !ok {
				return nil, &types.WrongFieldTypeError{FieldName: "float unique field", ExpectedType: "float64", ActualType: fmt.Sprintf("%T", val)}
			}
			v.UniqueValueFloat = f
		case "uuid":
			switch uv := val.(type) {
			case uuid.UUID:
				v.UniqueValueUUID = uv
			case string:
				u, err := uuid.Parse(uv)
				if err != nil {
					return nil, &types.WrongFieldTypeError{FieldName: "uuid unique field", ExpectedType: "uuid", ActualType: fmt.Sprintf("%T", uv)}
				}
				v.UniqueValueUUID = u
			default:
				return nil, &types.WrongFieldTypeError{FieldName: "uuid unique field", ExpectedType: "uuid", ActualType: fmt.Sprintf("%T", val)}
			}
		default:
			return nil, &types.WrongFieldTypeError{FieldName: "unknown unique field", ExpectedType: "known type", ActualType: fmt.Sprintf("%T", val)}
		}
	} else {
		// non-unique fields use the regular value fields (all types)
		switch strings.ToLower(fieldType) {
		case "string":
			s, ok := val.(string)
			if !ok {
				return nil, &types.WrongFieldTypeError{FieldName: "string field", ExpectedType: "string", ActualType: fmt.Sprintf("%T", val)}
			}
			v.ValueString = s
		case "int":
			i, ok := val.(int)
			if !ok {
				// also allow float values that are whole numbers for convenience, e.g. 1.0 -> 1
				if f, ok := val.(float64); ok {
					if i, err := FloatNoDecmimalToInt(f); err == nil {
						v.ValueInt = i
						break
					}
				}
				return nil, &types.WrongFieldTypeError{FieldName: "int field", ExpectedType: "int", ActualType: fmt.Sprintf("%T", val)}
			}
			v.ValueInt = i
		case "float":
			f, ok := val.(float64)
			if !ok {
				return nil, &types.WrongFieldTypeError{FieldName: "float field", ExpectedType: "float64", ActualType: fmt.Sprintf("%T", val)}
			}
			v.ValueFloat = f
		case "bool":
			b, ok := val.(bool)
			if !ok {
				return nil, &types.WrongFieldTypeError{FieldName: "bool field", ExpectedType: "bool", ActualType: fmt.Sprintf("%T", val)}
			}
			v.ValueBool = b
		case "time":
			switch tv := val.(type) {
			case time.Time:
				v.ValueTime = tv
			case string:
				t, err := time.Parse(time.RFC3339, tv)
				if err != nil {
					return nil, &types.WrongFieldTypeError{FieldName: "time field", ExpectedType: "ISO 8601 timestamp", ActualType: fmt.Sprintf("%T", val)}
				}
				v.ValueTime = t
			default:
				return nil, &types.WrongFieldTypeError{FieldName: "time field", ExpectedType: "time.Time or ISO 8601 timestamp", ActualType: fmt.Sprintf("%T", val)}
			}
		case "json":
			var s string
			switch jv := val.(type) {
			case string:
				s = jv
			default:
				b, err := json.Marshal(jv)
				if err != nil {
					return nil, &types.WrongFieldTypeError{FieldName: "json field", ExpectedType: "JSON serializable value", ActualType: fmt.Sprintf("%T", val)}
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
					return nil, &types.WrongFieldTypeError{FieldName: "uuid field", ExpectedType: "uuid", ActualType: fmt.Sprintf("%T", uv)}
				}
				v.ValueUUID = u
			default:
				return nil, &types.WrongFieldTypeError{FieldName: "uuid field", ExpectedType: "uuid", ActualType: fmt.Sprintf("%T", val)}
			}
		default:
			return nil, &types.WrongFieldTypeError{FieldName: "unknown field", ExpectedType: "known type", ActualType: fmt.Sprintf("%T", val)}
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
