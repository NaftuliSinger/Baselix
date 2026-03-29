package models

import (
	"context"

	"github.com/uptrace/bun"
)

type contextKey string

const ProjectIDKey contextKey = "project_id"

func WithProjectID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ProjectIDKey, id)
}

func ProjectIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ProjectIDKey).(string)
	return id, ok && id != ""
}

// ProjectScopedModel is a mixin that automatically adds a project_id filter
// to every SELECT query when a project ID is present in the context.
type ProjectScopedModel struct{}

func (ProjectScopedModel) BeforeSelect(ctx context.Context, q *bun.SelectQuery) error {
	if id, ok := ProjectIDFromContext(ctx); ok {
		q.Where("project_id = ?", id)
	}
	return nil
}

// LimitedSelectModel is a mixin that caps every SELECT query at 1000 rows.
type LimitedSelectModel struct{}

func (LimitedSelectModel) BeforeSelect(_ context.Context, q *bun.SelectQuery) error {
	q.Limit(1000)
	return nil
}
