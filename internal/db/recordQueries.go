package db

import (
	"baselix/internal/config"
	"baselix/internal/models"
	"baselix/internal/types"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
)

var uniqueViolationRe = regexp.MustCompile(`Key \((.+?)\)=\((.+?)\)`)

func parseUniqueViolationDetail(detail string) (value string) {
	// detail format: 'Key (field_name)=(value) already exists.'
	if m := uniqueViolationRe.FindStringSubmatch(detail); len(m) == 3 {
		return m[2]
	}
	return "unknown"
}

func InsertRecords(ctx context.Context, records []*models.Record) ([]*models.Record, error) {
	if len(records) == 0 {
		return records, nil
	}

	// if no project id is set on the records, we can't proceed
	if records[0].ProjectID == uuid.Nil {
		return nil, fmt.Errorf("project ID is required on records")
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert records; bun populates IDs via RETURNING
	_, err = tx.NewInsert().Model(&records).Exec(ctx)

	if err != nil {
		return nil, err
	}

	// Collect all values, assigning the now-populated RecordIDs
	var allValues []*models.Value
	for _, record := range records {
		for _, v := range record.Values {
			v.RecordID = record.ID
			allValues = append(allValues, v)
		}
	}

	if len(allValues) > 0 {
		_, err = tx.NewInsert().Model(&allValues).Exec(ctx)
		if err != nil {
			// if error is a unique violation, we want to return a custom error with the field name and value that caused the violation
			var pgErr pgdriver.Error
			if errors.As(err, &pgErr) && pgErr.Field('C') == "23505" {
				// Detail format for composite unique index: "Key (field_id, unique_value_str)=(uuid, actual-value) already exists."
				// m[1] = column names, m[2] = corresponding values
				rawValues := parseUniqueViolationDetail(pgErr.Field('D'))
				parts := strings.SplitN(rawValues, ", ", 2)

				var fieldNameStr, dupValue string
				if len(parts) == 2 {
					dupValue = strings.TrimSpace(parts[1])
					if fieldID, err := uuid.Parse(strings.TrimSpace(parts[0])); err == nil {
						for _, v := range allValues {
							if v.FieldID == fieldID && v.Field != nil {
								fieldNameStr = v.Field.Name
								break
							}
						}
					}
				}

				return nil, &types.UniqueDuplicateValueError{FieldName: fieldNameStr, Value: dupValue}
			}
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return records, nil
}

func UpdateRecordsByID(ctx context.Context, projectID uuid.UUID, tableName string, updates map[uuid.UUID]map[string]any) ([]*models.Record, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	table, err := GetTableWithFieldsByName(ctx, projectID, tableName)
	if err != nil {
		return nil, err
	}

	fieldMap := make(map[string]*models.Field, len(table.Fields))
	for _, f := range table.Fields {
		fieldMap[f.Name] = f
	}

	// Validate all field names and build all values before touching the DB
	var allValues []*models.Value
	var updatedRecordIDs []uuid.UUID
	for recordID, fieldValues := range updates {
		for fieldName, val := range fieldValues {
			field, ok := fieldMap[fieldName]
			if !ok {
				return nil, fmt.Errorf("field '%s' does not exist in table '%s'", fieldName, tableName)
			}
			v, err := NewValue(table.ID, recordID, field.ID, field.Type, field.Unique, val)
			if err != nil {
				return nil, err
			}
			allValues = append(allValues, v)
		}
		if len(fieldValues) > 0 {
			updatedRecordIDs = append(updatedRecordIDs, recordID)
		}
	}

	// Collect all record IDs for a single batch existence check
	allRecordIDs := make([]uuid.UUID, 0, len(updates))
	for recordID := range updates {
		allRecordIDs = append(allRecordIDs, recordID)
	}

	var existingRecords []struct {
		ID uuid.UUID `bun:"id"`
	}
	err = DB.NewSelect().
		TableExpr("records").
		Column("id").
		Where("id IN (?) AND project_id = ? AND table_id = ?", bun.In(allRecordIDs), projectID, table.ID).
		Scan(ctx, &existingRecords)
	if err != nil {
		return nil, err
	}
	existingSet := make(map[uuid.UUID]struct{}, len(existingRecords))
	for _, r := range existingRecords {
		existingSet[r.ID] = struct{}{}
	}
	for _, id := range allRecordIDs {
		if _, ok := existingSet[id]; !ok {
			return nil, &types.RecordNotFoundError{TableName: tableName, RecordID: id.String()}
		}
	}

	if len(updatedRecordIDs) == 0 {
		return nil, nil
	}

	now := time.Now()

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Single batch upsert for all values across all records
	_, err = tx.NewInsert().
		Model(&allValues).
		On("CONFLICT (record_id, field_id) DO UPDATE").
		Set("value_str = EXCLUDED.value_str").
		Set("value_int = EXCLUDED.value_int").
		Set("value_float = EXCLUDED.value_float").
		Set("value_bool = EXCLUDED.value_bool").
		Set("value_time = EXCLUDED.value_time").
		Set("value_json = EXCLUDED.value_json").
		Set("value_uuid = EXCLUDED.value_uuid").
		Set("unique_value_str = EXCLUDED.unique_value_str").
		Set("unique_value_int = EXCLUDED.unique_value_int").
		Set("unique_value_float = EXCLUDED.unique_value_float").
		Set("unique_value_uuid = EXCLUDED.unique_value_uuid").
		Exec(ctx)
	if err != nil {
		var pgErr pgdriver.Error
		if errors.As(err, &pgErr) && pgErr.Field('C') == "23505" {
			rawValues := parseUniqueViolationDetail(pgErr.Field('D'))
			parts := strings.SplitN(rawValues, ", ", 2)
			var fieldName, dupValue string
			if len(parts) == 2 {
				dupValue = strings.TrimSpace(parts[1])
				if fieldID, parseErr := uuid.Parse(strings.TrimSpace(parts[0])); parseErr == nil {
					for name, f := range fieldMap {
						if f.ID == fieldID {
							fieldName = name
							break
						}
					}
				}
			}
			return nil, &types.UniqueDuplicateValueError{FieldName: fieldName, Value: dupValue}
		}
		return nil, err
	}

	// Single batch update of updated_at for all affected records
	_, err = tx.NewUpdate().
		Model((*models.Record)(nil)).
		Set("updated_at = ?", now).
		Where(`"record"."id" IN (?) AND "record"."project_id" = ?`, bun.In(updatedRecordIDs), projectID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	var records []*models.Record
	err = DB.NewSelect().
		Model(&records).
		Where(`"record"."id" IN (?) AND "record"."project_id" = ?`, bun.In(updatedRecordIDs), projectID).
		Relation("Values").
		Relation("Values.Field").
		Relation("Table").
		Relation("Table.Fields").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func GetRecordsByTableID(ctx context.Context, projectID uuid.UUID, tableID uuid.UUID, fieldMap map[string]*models.Field, filters []types.Filter) ([]*models.Record, error) {
	var records []*models.Record
	q := DB.NewSelect().
		Model(&records).
		Where(`"record"."table_id" = ? AND "record"."project_id" = ?`, tableID, projectID).
		Relation("Values").
		Relation("Values.Field").
		Relation("Table").
		Relation("Table.Fields").
		Limit(config.Cfg.MaxSelectLimit)

	for _, f := range filters {
		cond, args, err := buildFilterCondition(f, fieldMap)
		if err != nil {
			return nil, err
		}
		q = q.Where(cond, args...)
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return records, nil
}

func GetRecordByID(ctx context.Context, projectID uuid.UUID, recordID uuid.UUID) (*models.Record, error) {
	var record models.Record
	err := DB.NewSelect().
		Model(&record).
		Where("id = ? AND project_id = ?", recordID, projectID).
		Relation("Values").
		Relation("Values.Field").
		Relation("Table").
		Relation("Table.Fields").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func DeleteRecordByID(ctx context.Context, projectID uuid.UUID, tableName string, recordID uuid.UUID) error {

	// delete all values for the record, then delete the record itself, as one transaction.
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check if table exists and get table id in the same query
	tableSubq := tx.NewSelect().TableExpr("tables").Column("id").Where("name = ? AND project_id = ?", tableName, projectID)

	// delete values for matching record id
	_, err = tx.NewDelete().Model((*models.Value)(nil)).Where("record_id = ? AND table_id = (?)", recordID, tableSubq).Exec(ctx)
	if err != nil {
		return err
	}

	// delete the record itself
	res, err := tx.NewDelete().Model((*models.Record)(nil)).Where("project_id = ? AND table_id = (?) AND id = ?", projectID, tableSubq, recordID).Exec(ctx)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &types.RecordNotFoundError{TableName: tableName, RecordID: recordID.String()}
	}

	return tx.Commit()
}

func CountRecordsByProjectID(ctx context.Context, projectID uuid.UUID) (int, error) {
	count, err := DB.NewSelect().Model((*models.Record)(nil)).Where("project_id = ?", projectID).Count(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}
