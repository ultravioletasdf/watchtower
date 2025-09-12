package utils

import (
	"context"
	"log"
	sqlc "videoapp/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type db struct {
	conn *pgxpool.Pool
	*sqlc.Queries
}

func (db *db) Close(ctx context.Context) {
	db.conn.Close()
}

func ConnectDatabase(cfg Config) db {
	conn, err := pgxpool.New(context.Background(), cfg.PostgresUrl)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	return db{Queries: sqlc.New(conn), conn: conn}
}
