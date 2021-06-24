package main

import (
	"log"
	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/zerologadapter"
	"github.com/rs/zerolog"
)

func mustSetupPostgres() *pgx.ConnPool {
	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		log.Fatal("specify DATABASE_URL")
	}

	connConfig, err := pgx.ParseConnectionString(pgURL)
	if err != nil {
		log.Fatal(err)
	}

	logger := zerolog.New(os.Stdout)
	connConfig.Logger = zerologadapter.NewLogger(logger)

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: 100,
	})
	if err != nil {
		log.Fatal(err)
	}

	return pool
}
