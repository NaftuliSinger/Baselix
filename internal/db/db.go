package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	"baselix/internal/config"
	"baselix/internal/models"
)

var DB *bun.DB

type debugQueryHook struct{}

func (h *debugQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *debugQueryHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
	if config.Cfg.Debug {
		log.Printf("[SQL] %s", event.Query)
		if event.Err != nil {
			log.Printf("[SQL ERR] %v", event.Err)
		}
	}
}

func Init(cfg *config.Config) {
	dsn := cfg.DB_DSN

	createTables := cfg.CreateTables

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	DB = bun.NewDB(sqldb, pgdialect.New())
	DB.AddQueryHook(&debugQueryHook{})

	if createTables {
		if err := createSchema(context.Background()); err != nil {
			log.Fatal(err)
		}
	}

	if err := DB.PingContext(context.Background()); err != nil {
		log.Fatal(err)
	}

	log.Println("DB connected")
}

func createSchema(ctx context.Context) error {
	models := []interface{}{
		// Core models
		(*models.User)(nil),
		(*models.Project)(nil),
		// EAV models
		(*models.Table)(nil),
		(*models.Field)(nil),
		(*models.Record)(nil),
		(*models.Value)(nil),
	}

	for _, model := range models {
		if _, err := DB.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}
