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

	"github.com/google/uuid"
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

func GetRecordsByTableID(ctx context.Context, projectID uuid.UUID, tableID uuid.UUID) ([]*models.Record, error) {
	var records []*models.Record
	err := DB.NewSelect().
		Model(&records).
		Where("table_id = ? AND project_id = ?", tableID, projectID).
		Relation("Values").
		Relation("Values.Field").
		Limit(config.Cfg.MaxSelectLimit).
		Scan(ctx)
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
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &record, nil
}
