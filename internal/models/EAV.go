package models

import (
	"github.com/uptrace/bun"
)

type Entity struct {
	bun.BaseModel `bun:"table:entities"`
}

type Attribute struct {
	bun.BaseModel `bun:"table:attributes"`
}

type Value struct {
	bun.BaseModel `bun:"table:values"`
}
