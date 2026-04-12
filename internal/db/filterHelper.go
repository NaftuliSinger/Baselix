package db

import (
	"baselix/internal/models"
	"baselix/internal/types"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// valueColumn returns the values table column name for the given field type and uniqueness.
func valueColumn(fieldType string, unique bool) string {
	switch fieldType {
	case "string":
		if unique {
			return "unique_value_str"
		}
		return "value_str"
	case "int":
		if unique {
			return "unique_value_int"
		}
		return "value_int"
	case "float":
		if unique {
			return "unique_value_float"
		}
		return "value_float"
	case "bool":
		return "value_bool"
	case "time":
		return "value_time"
	case "json":
		return "value_json"
	case "uuid":
		if unique {
			return "unique_value_uuid"
		}
		return "value_uuid"
	default:
		return "value_str"
	}
}

// toSQLOp maps a filter operator string to its SQL operator symbol.
func toSQLOp(op string) (string, error) {
	switch strings.ToLower(op) {
	case "eq":
		return "=", nil
	case "neq", "nq":
		return "<>", nil
	case "gt":
		return ">", nil
	case "gte", "ge":
		return ">=", nil
	case "lt":
		return "<", nil
	case "lte", "le":
		return "<=", nil
	case "contains":
		return "ILIKE", nil
	default:
		return "", fmt.Errorf("unknown filter operator %q", op)
	}
}

// coerceFilterValue parses the string value from the filter into the correct Go type
func coerceFilterValue(val string, fieldType string, op string) (any, error) {
	normOp := strings.ToLower(op)
	// Wrap % around the contains value for ILIKE query
	if normOp == "contains" {
		if fieldType != "string" && fieldType != "json" {
			return nil, fmt.Errorf("operator %q is only supported for string and json fields", op)
		}
		return "%" + val + "%", nil
	}
	switch fieldType {
	case "string", "json":
		return val, nil
	case "int":
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("filter value %q is not a valid int", val)
		}
		return n, nil
	case "float":
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("filter value %q is not a valid float", val)
		}
		return f, nil
	case "bool":
		b, err := strconv.ParseBool(val)
		if err != nil {
			return nil, fmt.Errorf("filter value %q is not a valid bool", val)
		}
		return b, nil
	case "time":
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return nil, fmt.Errorf("filter value %q is not a valid RFC3339 timestamp", val)
		}
		return t, nil
	case "uuid":
		u, err := uuid.Parse(val)
		if err != nil {
			return nil, fmt.Errorf("filter value %q is not a valid UUID", val)
		}
		return u, nil
	default:
		//
		return nil, fmt.Errorf("unknown field type %q for filter value", fieldType)
	}
}

var metaFilterFields = map[string]string{
	"id":         `"record"."id"`,
	"created_at": `"record"."created_at"`,
	"updated_at": `"record"."updated_at"`,
}

// buildFilterCondition builds a bun-compatible WHERE clause fragment for a single filter.
func buildFilterCondition(f types.Filter, fieldMap map[string]*models.Field) (string, []any, error) {
	normOp := strings.ToLower(f.Operator)

	// Meta fields: direct column comparison; null not meaningful here.
	if col, isMeta := metaFilterFields[f.Field]; isMeta {
		if f.Value == "null" {
			return "", nil, fmt.Errorf("null filter is not supported for meta field %q", f.Field)
		}
		sqlOp, err := toSQLOp(f.Operator)
		if err != nil {
			return "", nil, err
		}
		// For meta fields treat the type as string for coercion routing;
		// id is uuid, timestamps are time — handle specially.
		var coerced any
		switch f.Field {
		case "id":
			coerced, err = coerceFilterValue(f.Value, "uuid", f.Operator)
		case "created_at", "updated_at":
			coerced, err = coerceFilterValue(f.Value, "time", f.Operator)
		default:
			coerced, err = coerceFilterValue(f.Value, "string", f.Operator)
		}
		if err != nil {
			return "", nil, err
		}
		return col + " " + sqlOp + " ?", []any{coerced}, nil
	}

	// EAV field
	field, ok := fieldMap[f.Field]
	if !ok {
		// Caller should have already validated; guard defensively.
		return "", nil, fmt.Errorf("filter field %q does not exist in table schema", f.Field)
	}

	// null existence check
	if f.Value == "null" {
		switch normOp {
		case "eq":
			cond := `NOT EXISTS (SELECT 1 FROM "values" AS fv WHERE fv.record_id = "record"."id" AND fv.field_id = ?)`
			return cond, []any{field.ID}, nil
		case "neq", "nq":
			cond := `EXISTS (SELECT 1 FROM "values" AS fv WHERE fv.record_id = "record"."id" AND fv.field_id = ?)`
			return cond, []any{field.ID}, nil
		default:
			return "", nil, fmt.Errorf("operator %q cannot be used with null value", f.Operator)
		}
	}

	sqlOp, err := toSQLOp(f.Operator)
	if err != nil {
		return "", nil, err
	}
	coerced, err := coerceFilterValue(f.Value, field.Type, f.Operator)
	if err != nil {
		return "", nil, err
	}
	col := valueColumn(field.Type, field.Unique)
	cond := `EXISTS (SELECT 1 FROM "values" AS fv WHERE fv.record_id = "record"."id" AND fv.field_id = ? AND fv.` + col + ` ` + sqlOp + ` ?)`
	return cond, []any{field.ID, coerced}, nil
}
