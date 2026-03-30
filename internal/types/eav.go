package types

import "encoding/json"

type SchemaRequestBody struct {
	Schema map[string]any `json:"schema"`
}

type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TableResponse struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Fields map[string]string `json:"fields"`
}

type RecordResponse struct {
	ID     string         `json:"id"`
	Values map[string]any `json:"values"`
}

func (r RecordResponse) MarshalJSON() ([]byte, error) {
	// start with ID
	m := map[string]any{"id": r.ID}

	// merge all other fields
	for k, v := range r.Values {
		m[k] = v
	}

	return json.Marshal(m)
}
