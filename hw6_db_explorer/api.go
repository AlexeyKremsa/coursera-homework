package main

import (
	"fmt"
	"net/http"
	"strings"
)

func DeclareRoutes(exp *DBExplorer) {
	exp.router.RegisterRoute("GET", 0, exp.GetAllTables)
	exp.router.RegisterRoute("GET", 1, exp.GetRecordsFromTable)
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
			writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		}

		resp.Tables = append(resp.Tables, tableName)
	}

	writeResponseJSON(w, http.StatusOK, resp, "")
}

func (exp *DBExplorer) GetRecordsFromTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}

	tableName := strings.TrimPrefix(r.URL.Path, "/")
	table, ok := exp.tables[tableName]
	if !ok {
		writeResponseJSON(w, http.StatusNotFound, nil, fmt.Sprintf("table %s doesn't exist", tableName))
		return
	}

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "5"
	}

	offset := r.URL.Query().Get("offset")
	if offset == "" {
		offset = "0"

	}

	colNames := make([]string, 0)
	for _, col := range table.Columns {
		colNames = append(colNames, col.Field)
	}
	columns := strings.Join(colNames, ", ")
	query := fmt.Sprintf(`SELECT %s FROM %s LIMIT %s OFFSET %s`, columns, tableName, limit, offset)

	rows, err := exp.db.Query(query)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, rows, "")
}
