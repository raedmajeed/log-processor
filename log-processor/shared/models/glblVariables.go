package models

import "database/sql"

type SvcConfig struct {
	DbConn                *sql.DB
	RUNNING_PORT          string
	JWT_SECRET            string
	SUPEBASE_API          string
	SUPEBASE_API_KEY      string
	SUPEBASE_BUCKET       string
	SUPEBASE_STORAGE_BASE string
	SUPEBASE_REST_BASE    string
	DB_USER               string
	DB_PASSWORD           string
	DB_DATABASE           string
	DB_PORT               string
	DB_HOST               string
	REDIS_ADDR            string
	KEYWORD_CONFIG        string
}
