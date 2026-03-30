package utils

import (
	"baselix/internal/models"
	"reflect"
	"testing"
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
		{"name": "Alice", "age": 30},
		{"name": "Bob", "age": 25, "active": true},
	}
	expected := map[string]interface{}{
		"name":   "string",
		"age":    "float",
		"active": "bool",
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
