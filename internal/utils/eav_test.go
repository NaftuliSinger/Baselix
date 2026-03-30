package utils

import (
	"baselix/internal/models"
	"baselix/internal/types"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// Test RemoveIDFromSchemaMap
// Success case: map with "id" key should have it removed
func TestRemoveIDFromSchemaMap_WithID(t *testing.T) {
	input := map[string]interface{}{
		"id":   "123",
		"name": "Alice",
	}
	expected := map[string]interface{}{
		"name": "Alice",
	}
	result := RemoveIDFromSchemaMap(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Failure case: map without "id" key should remain unchanged
func TestRemoveIDFromSchemaMap_NoID(t *testing.T) {
	input := map[string]interface{}{
		"name": "Bob",
	}
	expected := map[string]interface{}{
		"name": "Bob",
	}
	result := RemoveIDFromSchemaMap(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Test ConvertSchemaMapToFields
// Success case: valid schema map should be converted to fields
func TestConvertSchemaMapToFields_WithValidSchema(t *testing.T) {
	input := map[string]interface{}{
		"name": "string",
		"age":  "int",
	}
	expected := []models.Field{
		{
			Name: "name",
			Type: "string",
		},
		{
			Name: "age",
			Type: "int",
		},
	}
	result := ConvertSchemaMapToFields(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Failure case: invalid type in schema map should be skipped
func TestConvertSchemaMapToFields_InvalidType(t *testing.T) {
	input := map[string]interface{}{
		"name": "string",
		"age":  123, // invalid type, should be string ("int")
	}
	expected := []models.Field{
		{
			Name: "name",
			Type: "string",
		},
	}
	result := ConvertSchemaMapToFields(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
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

// Failure case: ID field should be removed from the schema
func TestInferSchemaFromRecords_IDFieldRemoved(t *testing.T) {
	records := []map[string]any{
		{"id": "123", "name": "Alice"},
		{"id": "456", "name": "Bob"},
	}
	expected := map[string]interface{}{
		"name": "string",
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
		ID: uuid.New(),
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
		ID: record.ID.String(),
		Values: map[string]any{
			"name": "Alice",
			"age":  30,
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
		ID: uuid.New(),
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
		ID: record.ID.String(),
		Values: map[string]any{
			"name": "Alice",
			"data": nil, // unknown type should result in nil value
		},
	}
	result := MapRecordModelToRecordResponse(record)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
