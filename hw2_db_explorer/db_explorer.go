package main

import (
	"database/sql"
	"net/http"
)

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	return nil, nil
}
