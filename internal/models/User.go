package models

import (
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	UserID string `bun:"user_id,pk"`
	Plan   string `bun:"plan,notnull"`

	Projects []*Project `bun:"rel:has-many,join:user_id=user_id"`
}
