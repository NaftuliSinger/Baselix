package db

import (
	"baselix/internal/models"
	"baselix/internal/types"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NewValue creates a Value with only the relevant typed field set.
// attributeType must match Field.Type ("string", "int", "float", "bool", "time", "json", "uuid").
func NewValue(recordID uuid.UUID, fieldID uuid.UUID, attributeType types.AttributeType, val any) (*models.Value, error) {
	v := &models.Value{
		RecordID: recordID,
		FieldID:  fieldID,
	}
	switch attributeType {
	case types.AttributeTypeString:
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", val)
		}
		v.ValueString = s
	case types.AttributeTypeInt:
		i, ok := val.(int)
		if !ok {
			return nil, fmt.Errorf("expected int, got %T", val)
		}
		v.ValueInt = i
	case types.AttributeTypeFloat:
		f, ok := val.(float64)
		if !ok {
			return nil, fmt.Errorf("expected float64, got %T", val)
		}
		v.ValueFloat = f
	case types.AttributeTypeBool:
		b, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool, got %T", val)
		}
		v.ValueBool = b
	case types.AttributeTypeTime:
		t, ok := val.(time.Time)
		if !ok {
			return nil, fmt.Errorf("expected time.Time, got %T", val)
		}
		v.ValueTime = t
	case types.AttributeTypeJSON:
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("expected json string, got %T", val)
		}
		v.ValueJSON = s
	case types.AttributeTypeUUID:
		u, ok := val.(uuid.UUID)
		if !ok {
			return nil, fmt.Errorf("expected uuid.UUID, got %T", val)
		}
		v.ValueUUID = u
	default:
		return nil, fmt.Errorf("unknown attribute type: %q", attributeType)
	}
	return v, nil
}
