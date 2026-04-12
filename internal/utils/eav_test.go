package utils

import (
	"baselix/internal/models"
	"baselix/internal/types"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test ParseRecordsBody

// Success case: single JSON object should be returned as a one-element slice
func TestParseRecordsBody_SingleObject(t *testing.T) {
	body := []byte(`{"name":"Alice","age":30}`)
	result, err := ParseRecordsBody(body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result))
	}
	if result[0]["name"] != "Alice" {
		t.Errorf("Expected name=Alice, got %v", result[0]["name"])
	}
}

// Success case: JSON array should be returned as-is
func TestParseRecordsBody_ArrayOfObjects(t *testing.T) {
	body := []byte(`[{"name":"Alice"},{"name":"Bob"}]`)
	result, err := ParseRecordsBody(body)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result))
	}
}

// Failure case: empty body should return an error
func TestParseRecordsBody_EmptyBody(t *testing.T) {
	_, err := ParseRecordsBody([]byte{})
	if err == nil {
		t.Error("Expected error for empty body, got nil")
	}
}

// Failure case: empty JSON array should return an error
func TestParseRecordsBody_EmptyArray(t *testing.T) {
	_, err := ParseRecordsBody([]byte(`[]`))
	if err == nil {
		t.Error("Expected error for empty array, got nil")
	}
}

// Failure case: invalid JSON should return an error
func TestParseRecordsBody_InvalidJSON(t *testing.T) {
	_, err := ParseRecordsBody([]byte(`not json`))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// Test MapToUUIDMap

func TestMapToUUIDMap_Success(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	input := map[string]any{
		id1.String(): map[string]any{"name": "Alice"},
		id2.String(): map[string]any{"age": 30},
	}
	result, err := MapToUUIDMap(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(result))
	}
	if result[id1]["name"] != "Alice" {
		t.Errorf("Expected name=Alice for id1, got %v", result[id1]["name"])
	}
	if result[id2]["age"] != 30 {
		t.Errorf("Expected age=30 for id2, got %v", result[id2]["age"])
	}
}

func TestMapToUUIDMap_EmptyFields(t *testing.T) {
	id1 := uuid.New()
	input := map[string]any{
		id1.String(): map[string]any{},
	}
	result, err := MapToUUIDMap(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result[id1]) != 0 {
		t.Errorf("Expected empty fields map, got %v", result[id1])
	}
}

func TestMapToUUIDMap_InvalidID(t *testing.T) {
	input := map[string]any{
		"not-a-uuid": map[string]any{"name": "Alice"},
	}
	_, err := MapToUUIDMap(input)
	if err == nil {
		t.Error("Expected error for invalid UUID key, got nil")
	}
}

func TestMapToUUIDMap_ValueNotObject(t *testing.T) {
	id1 := uuid.New()
	input := map[string]any{
		id1.String(): "not-an-object",
	}
	_, err := MapToUUIDMap(input)
	if err == nil {
		t.Error("Expected error when value is not an object, got nil")
	}
}

// Test RecordsToUUIDMap
// Success case: valid records with "id" field should be converted to map

func TestRecordsToUUIDMap_Success(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	input := []map[string]any{
		{"id": id1.String(), "name": "Alice"},
		{"id": id2.String(), "age": 30},
	}
	result, err := RecordsToUUIDMap(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(result))
	}
	if result[id1]["name"] != "Alice" {
		t.Errorf("Expected name=Alice for id1, got %v", result[id1]["name"])
	}
	if _, hasID := result[id1]["id"]; hasID {
		t.Error("Expected 'id' key to be stripped from fields")
	}
}

// Success case: record with only "id" field should produce an entry with empty fields map
func TestRecordsToUUIDMap_EmptyFields(t *testing.T) {
	id1 := uuid.New()
	input := []map[string]any{
		{"id": id1.String()},
	}
	result, err := RecordsToUUIDMap(input)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result[id1]) != 0 {
		t.Errorf("Expected empty fields map, got %v", result[id1])
	}
}

// Failure case: record missing "id" field should return an error

func TestRecordsToUUIDMap_MissingID(t *testing.T) {
	input := []map[string]any{
		{"name": "Alice"},
	}
	_, err := RecordsToUUIDMap(input)
	if err == nil {
		t.Error("Expected error for missing id field, got nil")
	}
}

// Failure case: record with non-string "id" field should return an error
func TestRecordsToUUIDMap_InvalidID(t *testing.T) {
	input := []map[string]any{
		{"id": "not-a-uuid", "name": "Alice"},
	}
	_, err := RecordsToUUIDMap(input)
	if err == nil {
		t.Error("Expected error for invalid UUID id, got nil")
	}
}

// Failure case: record with "id" field that is not a string should return an error
func TestRecordsToUUIDMap_NonStringID(t *testing.T) {
	input := []map[string]any{
		{"id": 12345, "name": "Alice"},
	}
	_, err := RecordsToUUIDMap(input)
	if err == nil {
		t.Error("Expected error when id is not a string, got nil")
	}
}

// Test LowercaseSchemaMapValues
// Success case: map with mixed case keys should be converted to lowercase
func TestLowercaseSchemaMapValues_MixedCase(t *testing.T) {
	input := map[string]interface{}{
		"Name": "String",
		"Age":  "Int",
	}
	expected := map[string]interface{}{
		"Name": "string",
		"Age":  "int",
	}
	result, err := LowercaseSchemaMapValues(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Failure case: map with already lowercase keys should remain unchanged
func TestLowercaseSchemaMapValues_Lowercase(t *testing.T) {
	input := map[string]interface{}{
		"name": "string",
		"age":  "Int",
	}
	expected := map[string]interface{}{
		"name": "string",
		"age":  "int",
	}
	result, err := LowercaseSchemaMapValues(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Test CheckForReservedFields
// Success case: map with "id" key should have it removed
func TestCheckForReservedFields_Success(t *testing.T) {
	// multiple examples (id, created_at, updated_at)
	inputs := []map[string]interface{}{
		{"id": "123", "name": "Alice"},
		{"created_at": "2026-03-30T10:00:00Z", "name": "Bob"},
		{"updated_at": "2026-03-30T10:00:00Z", "name": "Charlie"},
	}

	// expected list of errors for each case
	expected := []error{
		&types.ReservedFieldError{FieldName: "id"},
		&types.ReservedFieldError{FieldName: "created_at"},
		&types.ReservedFieldError{FieldName: "updated_at"},
	}

	for i, input := range inputs {
		result := CheckForReservedFields(input)
		if result == nil || result.Error() != expected[i].Error() {
			t.Errorf("Expected %v, got %v", expected[i], result)
		}
	}
}

// Failure case: map without "id" key should remain unchanged
func TestCheckForReservedFields_Failure(t *testing.T) {
	input := map[string]interface{}{
		"name": "Bob",
	}
	// expect no error
	result := CheckForReservedFields(input)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

// Test isUniqueFieldType
// Success case: field name with "_u" suffix should be identified as unique
func TestIsUniqueFieldType_Unique(t *testing.T) {
	fieldName := "string_u"
	if !isUniqueFieldType(fieldName) {
		t.Errorf("Expected %q to be identified as unique", fieldName)
	}
}

// Failure case: field name without "_u" suffix should not be identified as unique
func TestIsUniqueFieldType_NotUnique(t *testing.T) {
	fieldName := "string"
	if isUniqueFieldType(fieldName) {
		t.Errorf("Expected %q to not be identified as unique", fieldName)
	}
}

// Test StripUniqueSuffix
// Success case: field name with "_u" suffix should have it stripped
func TestStripUniqueSuffix(t *testing.T) {
	fieldName := "string_u"
	expected := "string"
	result := StripUniqueSuffix(fieldName)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Failure case: field name without "_u" suffix should remain unchanged
func TestStripUniqueSuffix_NoSuffix(t *testing.T) {
	fieldName := "string"
	expected := "string"
	result := StripUniqueSuffix(fieldName)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Test IsValidFieldType
// Success case: valid field types should be recognized
func TestIsValidFieldType_Valid(t *testing.T) {
	validTypes := []string{"string", "int", "float", "bool", "time", "json", "uuid"}
	for _, fieldType := range validTypes {
		if !IsValidFieldType(fieldType) {
			t.Errorf("Expected %q to be recognized as a valid field type", fieldType)
		}
	}
}

// Failure case: invalid field types should not be recognized
func TestIsValidFieldType_Invalid(t *testing.T) {
	invalidTypes := []string{"str", "integer", "float64", "boolean", "datetime", "jsonb", "uuid4"}
	for _, fieldType := range invalidTypes {
		if IsValidFieldType(fieldType) {
			t.Errorf("Expected %q to not be recognized as a valid field type", fieldType)
		}
	}
}

// Test CleanAndConvertPayloadToFieldModels
// Success case: valid schema map should be converted to fields
func TestCleanAndConvertPayloadToFieldModels_WithValidSchema(t *testing.T) {
	input := map[string]interface{}{
		"name":   "string",
		"age":    "Int",      // Int is capitalized to test lowercasing
		"name_u": "string_u", // unique field with "_u" suffix
	}
	expected := []models.Field{
		{
			Name:   "name",
			Type:   "string",
			Unique: false,
		},
		{
			Name:   "age",
			Type:   "int",
			Unique: false,
		},
		{
			Name:   "name_u",
			Type:   "string",
			Unique: true,
		},
	}
	result, err := CleanAndConvertPayloadToFieldModels(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// sort both slices by field name before comparing to avoid order issues
	sortFields := func(fields []models.Field) {
		sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	}
	sortFields(expected)
	sortFields(result)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Failure case: using reserved field names should return an error
func TestCleanAndConvertPayloadToFieldModels_ReservedFieldNames(t *testing.T) {
	input := map[string]interface{}{
		"id":         "string",
		"created_at": "time",
		"updated_at": "time",
	}
	// expect an error due to reserved field names
	_, err := CleanAndConvertPayloadToFieldModels(input)
	if err == nil {
		t.Errorf("Expected error due to reserved field names, got nil")
	}
}

// Failure case: invalid type in schema map should be skipped
func TestCleanAndConvertPayloadToFieldModels_InvalidType(t *testing.T) {
	input := map[string]interface{}{
		"name": "bool_u", // invalid type with extra "u"
		"age":  "int",
	}
	// expect an error due to invalid type
	_, err := CleanAndConvertPayloadToFieldModels(input)
	if err == nil {
		t.Errorf("Expected error due to invalid field type, got nil")
	}
}

// Test InferTypeFromValue
// Success cases for different types
func TestInferTypeFromValue_Success(t *testing.T) {
	tests := []struct {
		value    any
		expected string
	}{
		{"hello", "string"},
		{123, "float"},
		{3.14, "float"},
		{true, "bool"},
	}
	for _, test := range tests {
		result := InferTypeFromValue(test.value)
		if result != test.expected {
			t.Errorf("Expected %s, got %s for value %v", test.expected, result, test.value)
		}
	}
}

// Failure case: unknown type should return "unknown"
func TestInferTypeFromValue_UnknownType(t *testing.T) {
	var unknownType struct{}
	result := InferTypeFromValue(unknownType)
	if result != "unknown" {
		t.Errorf("Expected %s, got %s for value %v", "unknown", result, unknownType)
	}
}

// Test InferSchemaFromRecords
// Success case: multiple records with different fields should produce a merged schema
func TestInferSchemaFromRecords_Success(t *testing.T) {
	records := []map[string]any{
		{"name": "Alice", "age": 30, "active": true, "score": 3.14, "created_at": "2026-03-30T10:00:00Z", "user_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "meta": map[string]any{"key": "value"}},
	}
	expected := map[string]interface{}{
		"name":       "string",
		"age":        "float",
		"active":     "bool",
		"score":      "float",
		"created_at": "time",
		"user_id":    "uuid",
		"meta":       "json",
	}
	result := InferSchemaFromRecords(records)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Failure case: schema that includes unknown types should not be included in the result
func TestInferSchemaFromRecords_UnknownType(t *testing.T) {
	records := []map[string]any{
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25, "data": struct{}{}}, // unknown type
	}
	expected := map[string]interface{}{
		"name": "string",
		"age":  "float",
	}
	result := InferSchemaFromRecords(records)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Test MapRecordToResponse
// Success case: valid Record model should be mapped to RecordResponse
func TestMapRecordToResponse_Success(t *testing.T) {
	record := &models.Record{
		// use a random UUID and a fixed timestamp for testing
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Values: []*models.Value{
			{
				Field:       &models.Field{Name: "name", Type: "string"},
				ValueString: "Alice",
			},
			{
				Field:    &models.Field{Name: "age", Type: "int"},
				ValueInt: 30,
			},
		},
	}
	expected := types.RecordResponse{
		ID:        record.ID.String(),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		Values: []types.RecordField{
			{Key: "name", Value: "Alice"},
			{Key: "age", Value: 30},
		},
	}
	result := MapRecordModelToRecordResponse(record)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Failure case: Record with unknown field types should return nil values for those fields
func TestMapRecordToResponse_UnknownFieldType(t *testing.T) {
	record := &models.Record{
		// use a random UUID and a fixed timestamp for testing
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Values: []*models.Value{
			{
				Field:       &models.Field{Name: "name", Type: "string"},
				ValueString: "Alice",
			},
			{
				Field:       &models.Field{Name: "data", Type: "unknown"},
				ValueString: "unknown",
			},
		},
	}
	expected := types.RecordResponse{
		ID:        record.ID.String(),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
		Values: []types.RecordField{
			{Key: "name", Value: "Alice"},
			{Key: "data", Value: nil}, // unknown type should result in nil value
		},
	}
	result := MapRecordModelToRecordResponse(record)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
