package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/AlexeyKremsa/coursera-homework/hw6_db_explorer/router"
)

type DBExplorer struct {
	db     *sql.DB
	router *router.Router
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	exp := DBExplorer{db: db, router: router.New()}
	router.Init(exp.router)
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

func (exp *DBExplorer) GetAllTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}
	fmt.Println(r.URL.Query().Get("test"))
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
