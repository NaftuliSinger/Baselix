package db

import (
	"baselix/internal/models"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func GetProjectTablesWithFields(ctx context.Context, projectID uuid.UUID) ([]models.Table, error) {
	/*
		We want to get all tables for a project, along with their fields.
		We are utilising Bun's hook to filter the tables by project ID
	*/

	var tables []models.Table
	err := DB.NewSelect().
		Model(&tables).
		Column("id", "name", "project_id").
		Relation("Project", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("id", "name")
		}).
		Relation("Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("id", "table_id", "name", "type")
		}).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

func GetTableWithFields(ctx context.Context, projectID uuid.UUID, tableName string) (*models.Table, error) {
	/*
		We want to get a specific table for a project, along with its fields.
		We are utilising Bun's hook to filter the table by project ID and table name
	*/

	var table models.Table
	err := DB.NewSelect().
		Model(&table).
		Column("id", "name", "project_id").
		Relation("Project", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("id", "name")
		}).
		Relation("Fields", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("id", "table_id", "name", "type")
		}).
		Where("\"table\".name = ?", tableName).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &table, nil
}

func ExistsOrCreateTable(ctx context.Context, projectID uuid.UUID, tableName string, fields []models.Field) (table *models.Table, created bool, err error) {
	table, err = GetTableWithFields(ctx, projectID, tableName)
	if err != nil {
		return nil, false, err
	}
	if table != nil {
		return table, false, nil
	}
	table, err = CreateTableWithFields(ctx, projectID, tableName, fields)
	if err != nil {
		return nil, false, err
	}
	return table, true, nil
}

func CreateTableWithFields(ctx context.Context, projectID uuid.UUID, tableName string, fields []models.Field) (table *models.Table, err error) {
	// Start a transaction
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Create the table
	table = &models.Table{
		ProjectID: projectID,
		Name:      tableName,
	}
	if _, err = tx.NewInsert().Model(table).Exec(ctx); err != nil {
		return nil, err
	}

	// Create the fields
	for i := range fields {
		field := &fields[i]
		field.TableID = table.ID
		if _, err = tx.NewInsert().Model(field).Exec(ctx); err != nil {
			return nil, err
		}
	}

	return table, nil
}

// UpdateTableFields reconciles the fields of an existing table with the
// provided fields slice. It inserts new fields, updates the type of fields
// whose type has changed (only when no values exist yet), and deletes fields
// that are no longer present. All changes are applied inside a single transaction.
func UpdateTableFields(ctx context.Context, tableName string, fields []models.Field) (table *models.Table, err error) {
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Fetch the table together with its current fields so we can diff them.
	table = &models.Table{}
	if err = tx.NewSelect().Model(table).
		Relation("Fields").
		Where("name = ?", tableName).
		Scan(ctx); err != nil {
		return nil, err
	}

	// Index the current fields by name for O(1) look-ups below.
	existingFields := make(map[string]*models.Field, len(table.Fields))
	for _, f := range table.Fields {
		existingFields[f.Name] = f
	}

	// Track which field names are present in the incoming payload so we can
	// identify and delete fields that were removed.
	incomingNames := make(map[string]bool, len(fields))
	for _, field := range fields {
		incomingNames[field.Name] = true
	}

	// Insert new fields; update the type of existing fields when safe to do so.
	for i := range fields {
		field := &fields[i]
		existing, exists := existingFields[field.Name]
		if !exists {
			// New field — insert it.
			field.TableID = table.ID
			if _, err = tx.NewInsert().Model(field).Exec(ctx); err != nil {
				return nil, err
			}
			continue
		}
		if existing.Type == field.Type {
			// Type unchanged — nothing to do.
			continue
		}
		// Type changed — only allow if no values have been stored yet,
		// otherwise we would silently lose data.
		var count int
		if count, err = tx.NewSelect().Model((*models.Value)(nil)).
			Where("field_id = ?", existing.ID).
			Count(ctx); err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, fmt.Errorf("cannot change type of field %q: %d existing value(s) would be lost", field.Name, count)
		}
		if _, err = tx.NewUpdate().Model((*models.Field)(nil)).
			Set("type = ?", field.Type).
			Where("id = ?", existing.ID).
			Exec(ctx); err != nil {
			return nil, err
		}
	}

	// Delete fields that were omitted from the incoming payload.
	for _, f := range table.Fields {
		if !incomingNames[f.Name] {
			if _, err = tx.NewDelete().Model((*models.Field)(nil)).
				Where("id = ?", f.ID).
				Exec(ctx); err != nil {
				return nil, err
			}
		}
	}

	return table, nil
}

func DeleteTable(ctx context.Context, tableName string) error {
	// We rely on ON DELETE CASCADE to automatically delete the fields, records and values related to this table, so we only need to delete the table itself.
	result, err := DB.NewDelete().Model((*models.Table)(nil)).
		Where("name = ?", tableName).
		Exec(ctx)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("table %q not found", tableName)
	}
	return nil
}
