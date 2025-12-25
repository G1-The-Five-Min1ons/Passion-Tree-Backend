package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

func NewAzureSQL(username, password, hostname, dbName string) (*sql.DB, error) {
	query := url.Values{}
	query.Add("database", dbName)
	query.Add("encrypt", "true")

	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(username, password),
		Host:     fmt.Sprintf("%s:%d", hostname, 1433),
		RawQuery: query.Encode(),
	}

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}