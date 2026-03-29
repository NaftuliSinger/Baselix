package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Project struct {
	bun.BaseModel `bun:"table:projects"`

	ID          uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	UserID      string    `bun:"user_id,notnull,unique:idx_user_name"` // part of composite unique
	Name        string    `bun:"name,notnull,unique:idx_user_name"`    // part of composite unique
	APIKeyHash  string    `bun:"api_key_hash,unique,notnull"`
	Description string    `bun:"description"`

	// Relationships
	User     *User     `bun:"rel:belongs-to,join:user_id=user_id"`
	Entities []*Table  `bun:"rel:has-many,join:id=project_id"`
	Records  []*Record `bun:"rel:has-many,join:id=project_id"`
}
