package postgres

import (
	"database/sql"
	"fmt"
	"take-home-test/internal/configs"

	_ "github.com/lib/pq"
)

func Open(cfg *configs.Config) (*sql.DB, error) {
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresConfig.Host,
		cfg.PostgresConfig.Port,
		cfg.PostgresConfig.Username,
		cfg.PostgresConfig.Password,
		cfg.PostgresConfig.DBName)

	db, err := sql.Open("postgres", psqlConn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
