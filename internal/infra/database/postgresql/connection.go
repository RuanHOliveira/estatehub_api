package postgresql

import (
	"context"
	"fmt"
	"log"

	"github.com/RuanHOliveira/estatehub_api/internal/core/config"
	"github.com/jackc/pgx/v5"
)

func Connect(ctx context.Context, c *config.PgConfig) *pgx.Conn {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.PgHost,
		c.PgPort,
		c.PgUser,
		c.PgPass,
		c.PgName,
	)

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to create database connection: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL.")

	return conn
}
