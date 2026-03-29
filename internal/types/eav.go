package types

type SchemaRequestBody struct {
	Schema map[string]any `json:"schema"`
}

type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TableResponse struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Project ProjectResponse `json:"project"`
	Fields  []FieldResponse `json:"fields"`
}

type FieldResponse struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
