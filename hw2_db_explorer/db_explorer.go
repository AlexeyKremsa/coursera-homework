package main

import (
	"database/sql"
	"net/http"
)

type DBExplorer struct {
	db *sql.DB
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	exp := DBExplorer{db: db}
	return &exp, nil
}

func (exp *DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		exp.GetAllTables(w, r)
	default:
		writeResponseJSON(w, http.StatusNotFound, nil, "unknown table")
	}
}

type Table struct {
	Name string
}

type TablesResp struct {
	Tables []string `json:"tables"`
}

func (exp *DBExplorer) GetAllTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}

	rows, err := exp.db.Query("SHOW TABLES")
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, err.Error(), "db error")
	}

	resp := TablesResp{}
	resp.Tables = make([]string, 0)

	for rows.Next() {
		var tableName string

		err = rows.Scan(&tableName)
		if err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, err.Error(), "db error")
		}

		resp.Tables = append(resp.Tables, tableName)
	}

	writeResponseJSON(w, http.StatusOK, resp, "")
}
