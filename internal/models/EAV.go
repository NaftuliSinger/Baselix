package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Entity struct {
	bun.BaseModel `bun:"table:entities"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	ProjectID string `bun:",type:uuid,notnull, unique:idx_project_name"` // part of composite unique
	ProjectScopedModel

	Name string `bun:"name,notnull, unique:idx_project_name"` // part of composite unique

	// Relationships
	Project    *Project     `bun:"rel:has-one,join:project_id=id"`
	Attributes []*Attribute `bun:"rel:has-many,join:id=entity_id"`
	Records    []*Record    `bun:"rel:has-many,join:id=entity_id"`
}

type Attribute struct {
	bun.BaseModel `bun:"table:attributes"`

	ID uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	EntityID uuid.UUID `bun:",type:uuid,notnull, unique:idx_entity_name"` // part of composite unique

	Name string `bun:"name,notnull, unique:idx_entity_name"` // part of composite unique
	Type string `bun:"type,notnull"`

	// Relationships
	Entity  *Entity   `bun:"rel:has-one,join:entity_id=id"`
	Records []*Record `bun:"rel:has-many,join:id=attribute_id"`
}

type Record struct {
	bun.BaseModel `bun:"table:records"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	ProjectID string `bun:",type:uuid,notnull, unique:idx_project_name"` // part of composite unique
	ProjectScopedModel
	EntityID uuid.UUID `bun:",type:uuid,notnull"`

	// Relationships
	Project *Project `bun:"rel:has-one,join:project_id=id"`
	Entity  *Entity  `bun:"rel:has-one,join:entity_id=id"`
	Values  []*Value `bun:"rel:has-many,join:id=record_id"`
}

type Value struct {
	bun.BaseModel `bun:"table:values"`

	ID uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	RecordID    uuid.UUID `bun:",type:uuid,notnull, unique:idx_record_attribute"` // part of composite unique
	AttributeID uuid.UUID `bun:",type:uuid,notnull, unique:idx_record_attribute"` // part of composite unique
	ValueString string    `bun:"value,nullzero"`
	ValueInt    int       `bun:"value_int,nullzero"`
	ValueFloat  float64   `bun:"value_float,nullzero"`
	ValueBool   bool      `bun:"value_bool,nullzero"`
	ValueTime   time.Time `bun:"value_time,nullzero"`
	ValueJSON   string    `bun:"value_json,nullzero"`
	ValueUUID   uuid.UUID `bun:"value_uuid,type:uuid,nullzero"`

	// Relationships
	Record    *Record    `bun:"rel:has-one,join:record_id=id"`
	Attribute *Attribute `bun:"rel:has-one,join:attribute_id=id"`
}

// NewValue creates a Value with only the relevant typed field set.
// attributeType must match Attribute.Type ("string", "int", "float", "bool", "time", "json").
func NewValue(recordID, attributeID uuid.UUID, attributeType string, val any) (*Value, error) {
	v := &Value{
		RecordID:    recordID,
		AttributeID: attributeID,
	}
	switch attributeType {
	case "string":
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", val)
		}
		v.ValueString = s
	case "int":
		i, ok := val.(int)
		if !ok {
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
		s, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("expected json string, got %T", val)
		}
		v.ValueJSON = s
	case "uuid":
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
