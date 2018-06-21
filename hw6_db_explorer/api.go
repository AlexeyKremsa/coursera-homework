package main

import (
	"fmt"
	"net/http"
	"strconv"
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

	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 0, 32)
	if err != nil {
		limit = 5
	}

	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 0, 32)
	if err != nil {
		offset = 0
	}

	colNames := make([]string, 0)
	for _, col := range table.Columns {
		colNames = append(colNames, col.Field)
	}
	columns := strings.Join(colNames, ", ")

	query := fmt.Sprintf(`SELECT %s FROM %s LIMIT %d OFFSET %d`, columns, tableName, limit, offset)

	rows, err := exp.db.Query(query)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}
	defer rows.Close()

	colsToRead := make([]interface{}, 0)
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	for _, item := range colTypes {
		col, err := getVariable(item.DatabaseTypeName())
		if err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
		colsToRead = append(colsToRead, col)
	}

	for rows.Next() {
		err = rows.Scan(colsToRead...)
		if err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
	}

	resp, err := prepareResponse(colsToRead, colNames)
	if err != nil {
		if err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
			return
		}
	}
	writeResponseJSON(w, http.StatusOK, resp, "")
}
