package types

import (
	"bytes"
	"encoding/json"
	"time"
)

type SchemaRequestBody struct {
	Schema map[string]any `json:"schema"`
}

type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TableField struct {
	Key   string
	Value string
}

type TableResponse struct {
	Name   string       `json:"name"`
	Fields []TableField // ordered; marshaled as object
}

func (t TableResponse) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`{"name":`)
	name, err := json.Marshal(t.Name)
	if err != nil {
		return nil, err
	}
	buf.Write(name)
	buf.WriteString(`,"fields":{`)
	for i, f := range t.Fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		k, err := json.Marshal(f.Key)
		if err != nil {
			return nil, err
		}
		v, err := json.Marshal(f.Value)
		if err != nil {
			return nil, err
		}
		buf.Write(k)
		buf.WriteByte(':')
		buf.Write(v)
	}
	buf.WriteString(`}}`)
	return buf.Bytes(), nil
}

type RecordField struct {
	Key   string
	Value any
}

type RecordResponse struct {
	ID        string        `json:"id"`
	CreatedAt time.Time     `json:"created_at,omitempty"`
	UpdatedAt time.Time     `json:"updated_at,omitempty"`
	Values    []RecordField // ordered; marshaled as flat keys
}

func (r RecordResponse) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	writeField := func(comma bool, key string, val any) error {
		if comma {
			buf.WriteByte(',')
		}
		k, err := json.Marshal(key)
		if err != nil {
			return err
		}
		buf.Write(k)
		buf.WriteByte(':')
		v, err := json.Marshal(val)
		if err != nil {
			return err
		}
		buf.Write(v)
		return nil
	}

	buf.WriteByte('{')

	if err := writeField(false, "id", r.ID); err != nil {
		return nil, err
	}
	if !r.CreatedAt.IsZero() {
		if err := writeField(true, "created_at", r.CreatedAt); err != nil {
			return nil, err
		}
	}
	if !r.UpdatedAt.IsZero() {
		if err := writeField(true, "updated_at", r.UpdatedAt); err != nil {
			return nil, err
		}
	}
	for _, f := range r.Values {
		if err := writeField(true, f.Key, f.Value); err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}
