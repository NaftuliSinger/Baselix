package db

import (
	"baselix/internal/types"
	"strings"

	"github.com/uptrace/bun"
)

func ApplyFilters(q *bun.SelectQuery, filters []types.Filter) *bun.SelectQuery {
	for _, f := range filters {
		switch f.Operator {
		case "eq":
			q = q.Where("? = ?", bun.Ident(f.Field), f.Value)

		case "neq":
			q = q.Where("? != ?", bun.Ident(f.Field), f.Value)

		case "gt":
			q = q.Where("? > ?", bun.Ident(f.Field), f.Value)

		case "gte":
			q = q.Where("? >= ?", bun.Ident(f.Field), f.Value)

		case "lt":
			q = q.Where("? < ?", bun.Ident(f.Field), f.Value)

		case "lte":
			q = q.Where("? <= ?", bun.Ident(f.Field), f.Value)

		case "like":
			q = q.Where("? LIKE ?", bun.Ident(f.Field), "%"+f.Value+"%")

		case "ilike":
			q = q.Where("? ILIKE ?", bun.Ident(f.Field), "%"+f.Value+"%")
		}
	}

	return q
}

func ApplySorts(q *bun.SelectQuery, sorts []types.Sort) *bun.SelectQuery {
	for _, s := range sorts {
		dir := "ASC"
		if strings.ToLower(s.Dir) == "desc" {
			dir = "DESC"
		}

		q = q.OrderExpr("? "+dir, bun.Ident(s.Field))
	}

	return q
}
