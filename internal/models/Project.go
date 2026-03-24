package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Project struct {
	bun.BaseModel `bun:"table:projects"`

	ID          uuid.UUID `bun:"id,pk,type:uuid"`
	UserID      string    `bun:"user_id,notnull"`
	APIKeyHash  string    `bun:"api_key_hash,unique,notnull"`
	Name        string    `bun:"name,notnull"`
	Description string    `bun:"description"`
}
