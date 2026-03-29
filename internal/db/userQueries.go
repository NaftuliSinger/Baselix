package db

import (
	"baselix/internal/models"
	"context"
)

func SelectUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := DB.NewSelect().Model(&user).Where("user_id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func InsertUser(ctx context.Context, user *models.User) error {
	_, err := DB.NewInsert().Model(user).Exec(ctx)
	return err
}

func UpdateUser(ctx context.Context, user *models.User) error {
	_, err := DB.NewUpdate().Model(user).Where("user_id = ?", user.UserID).Exec(ctx)
	return err
}
