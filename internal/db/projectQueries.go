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
