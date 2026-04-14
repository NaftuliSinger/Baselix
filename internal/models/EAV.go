package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Table struct {
	bun.BaseModel `bun:"table:tables"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp"`

	ProjectID uuid.UUID `bun:",type:uuid,notnull,unique:idx_project_name"` // part of composite unique

	Name string `bun:"name,notnull,unique:idx_project_name"` // part of composite unique

	// Relationships
	Project *Project  `bun:"rel:has-one,join:project_id=id" json:"project,omitempty"`
	Fields  []*Field  `bun:"rel:has-many,join:id=table_id" json:"fields,omitempty"`
	Records []*Record `bun:"rel:has-many,join:id=table_id" json:"records,omitempty"`
	Values  []*Value  `bun:"rel:has-many,join:id=table_id" json:"values,omitempty"`
}

type Field struct {
	bun.BaseModel `bun:"table:fields"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	TableID uuid.UUID `bun:",type:uuid,notnull,unique:idx_table_name"` // part of composite unique
	Name    string    `bun:"name,notnull,unique:idx_table_name"`       // part of composite unique
	Type    string    `bun:"type,notnull"`
	Unique  bool      `bun:"unique,notnull,default:false"`

	// Relationships
	Table  *Table   `bun:"rel:has-one,join:table_id=id" json:",omitempty"`
	Values []*Value `bun:"rel:has-many,join:id=field_id" json:",omitempty"`
}

type Record struct {
	bun.BaseModel `bun:"table:records"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	ProjectID uuid.UUID `bun:",type:uuid,notnull"`

	TableID uuid.UUID `bun:",type:uuid,notnull"`

	// Relationships
	Project *Project `bun:"rel:has-one,join:project_id=id" json:",omitempty"`
	Table   *Table   `bun:"rel:has-one,join:table_id=id" json:",omitempty"`
	Values  []*Value `bun:"rel:has-many,join:id=record_id" json:",omitempty"`
}

type Value struct {
	bun.BaseModel `bun:"table:values"`

	ID uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	TableID     uuid.UUID `bun:"table_id,type:uuid,notnull"`
	RecordID    uuid.UUID `bun:",type:uuid,notnull,unique:idx_record_field"`                                                                                                                    // part of composite unique
	FieldID     uuid.UUID `bun:",type:uuid,notnull,unique:idx_record_field,unique:idx_field_unique_str,unique:idx_field_unique_int,unique:idx_field_unique_float,unique:idx_field_unique_uuid"` // part of multiple unique indexes
	ValueString string    `bun:"value_str,nullzero"`
	ValueInt    int       `bun:"value_int,nullzero"`
	ValueFloat  float64   `bun:"value_float,nullzero"`
	ValueBool   bool      `bun:"value_bool,nullzero"`
	ValueTime   time.Time `bun:"value_time,nullzero"`
	ValueJSON   string    `bun:"value_json,nullzero"`
	ValueUUID   uuid.UUID `bun:"value_uuid,type:uuid,nullzero"`

	// Populated instead of Value* when Field.Unique=true.
	// Each column pairs with FieldID to form a composite unique index per field.
	UniqueValueString string    `bun:"unique_value_str,nullzero,unique:idx_field_unique_str"`
	UniqueValueInt    int       `bun:"unique_value_int,nullzero,unique:idx_field_unique_int"`
	UniqueValueFloat  float64   `bun:"unique_value_float,nullzero,unique:idx_field_unique_float"`
	UniqueValueUUID   uuid.UUID `bun:"unique_value_uuid,type:uuid,nullzero,unique:idx_field_unique_uuid"`

	// Relationships
	Table  *Table  `bun:"rel:has-one,join:table_id=id" json:",omitempty"`
	Record *Record `bun:"rel:has-one,join:record_id=id" json:",omitempty"`
	Field  *Field  `bun:"rel:has-one,join:field_id=id" json:",omitempty"`
}
