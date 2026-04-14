package db

import (
	"baselix/internal/models"
	"context"
)

func SelectProjectsByUserID(ctx context.Context, userID string) ([]*models.Project, error) {
	var projects []*models.Project
	err := DB.NewSelect().Model(&projects).Where("user_id = ?", userID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func SelectProjectByID(ctx context.Context, id string) (*models.Project, error) {
	var project models.Project
	err := DB.NewSelect().Model(&project).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func SelectProjectWithUserByAPIKeyHash(ctx context.Context, apiKeyHash string) (*models.Project, error) {
	var project models.Project
	err := DB.NewSelect().Model(&project).Relation("User").Where("api_key_hash = ?", apiKeyHash).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func InsertProject(ctx context.Context, project *models.Project) error {
	_, err := DB.NewInsert().Model(project).Exec(ctx)
	return err
}

func UpdateProjectByID(ctx context.Context, project *models.Project) error {
	_, err := DB.NewUpdate().Model(project).Where("id = ?", project.ID).Exec(ctx)
	return err
}

func CountProjectsByUserID(ctx context.Context, userID string) (int, error) {
	count, err := DB.NewSelect().Model((*models.Project)(nil)).Where("user_id = ?", userID).Count(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteProjectByID deletes a project and all its related data (tables, fields, records, values) by project ID.
func DeleteProjectByID(ctx context.Context, projectID string) error {
	// new transaction
	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Delete values belonging to any record of this project.
	if _, err = tx.NewDelete().Model((*models.Value)(nil)).
		Where("record_id IN (SELECT id FROM records WHERE project_id = ?)", projectID).
		Exec(ctx); err != nil {
		return err
	}

	// Delete records belonging to this project.
	if _, err = tx.NewDelete().Model((*models.Record)(nil)).
		Where("project_id = ?", projectID).
		Exec(ctx); err != nil {
		return err
	}

	// Delete fields belonging to tables of this project.
	if _, err = tx.NewDelete().Model((*models.Field)(nil)).
		Where("table_id IN (SELECT id FROM tables WHERE project_id = ?)", projectID).
		Exec(ctx); err != nil {
		return err
	}

	// Delete tables belonging to this project.
	if _, err = tx.NewDelete().Model((*models.Table)(nil)).
		Where("project_id = ?", projectID).
		Exec(ctx); err != nil {
		return err
	}

	// Delete the project itself.
	if _, err = tx.NewDelete().Model((*models.Project)(nil)).
		Where("id = ?", projectID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}
