package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Table struct {
	bun.BaseModel `bun:"table:tables"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	ProjectID uuid.UUID `bun:",type:uuid,notnull,unique:idx_project_name"` // part of composite unique
	ProjectScopedModel

	Name string `bun:"name,notnull,unique:idx_project_name"` // part of composite unique

	// Relationships
	Project *Project  `bun:"rel:has-one,join:project_id=id" json:"project,omitempty"`
	Fields  []*Field  `bun:"rel:has-many,join:id=table_id" json:"fields,omitempty"`
	Records []*Record `bun:"rel:has-many,join:id=table_id" json:"records,omitempty"`
}

type Field struct {
	bun.BaseModel `bun:"table:fields"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	TableID uuid.UUID `bun:",type:uuid,notnull,unique:idx_table_name"` // part of composite unique
	Name    string    `bun:"name,notnull,unique:idx_table_name"`       // part of composite unique
	Type    string    `bun:"type,notnull"`

	// Relationships
	Table  *Table   `bun:"rel:has-one,join:table_id=id" json:",omitempty"`
	Values []*Value `bun:"rel:has-many,join:id=field_id" json:",omitempty"`
}

type Record struct {
	bun.BaseModel `bun:"table:records"`
	ID            uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	ProjectID uuid.UUID `bun:",type:uuid,notnull"`
	ProjectScopedModel

	TableID uuid.UUID `bun:",type:uuid,notnull"`

	// Relationships
	Project *Project `bun:"rel:has-one,join:project_id=id" json:",omitempty"`
	Table   *Table   `bun:"rel:has-one,join:table_id=id" json:",omitempty"`
	Values  []*Value `bun:"rel:has-many,join:id=record_id" json:",omitempty"`
}

type Value struct {
	bun.BaseModel `bun:"table:values"`

	ID uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`

	RecordID    uuid.UUID `bun:",type:uuid,notnull,unique:idx_record_field"` // part of composite unique
	FieldID     uuid.UUID `bun:",type:uuid,notnull,unique:idx_record_field"` // part of composite unique
	ValueString string    `bun:"value,nullzero"`
	ValueInt    int       `bun:"value_int,nullzero"`
	ValueFloat  float64   `bun:"value_float,nullzero"`
	ValueBool   bool      `bun:"value_bool,nullzero"`
	ValueTime   time.Time `bun:"value_time,nullzero"`
	ValueJSON   string    `bun:"value_json,nullzero"`
	ValueUUID   uuid.UUID `bun:"value_uuid,type:uuid,nullzero"`

	// Relationships
	Record *Record `bun:"rel:has-one,join:record_id=id" json:",omitempty"`
	Field  *Field  `bun:"rel:has-one,join:field_id=id" json:",omitempty"`
}
