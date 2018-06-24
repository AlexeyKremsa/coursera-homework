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
	exp.router.RegisterRoute("GET", 2, exp.GetRecordByID)
	exp.router.RegisterRoute("PUT", 1, exp.CreateRecord)
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

	// select all columns
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		colNames = append(colNames, col.Field)
	}
	columns := strings.Join(colNames, ", ")

	query := fmt.Sprintf(`SELECT %s FROM %s LIMIT %d OFFSET %d`, columns, tableName, limit, offset)

	resp, err := exp.ExecuteQuery(query, colNames)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, resp, "")
}

func (exp *DBExplorer) GetRecordByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}

	params := strings.Split(r.URL.Path, "/")
	if len(params) != 3 {
		writeResponseJSON(w, http.StatusBadRequest, nil, "invalid URL")
		return
	}

	tableName := params[1]
	id, err := strconv.Atoi(params[2])
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	table, ok := exp.tables[tableName]
	if !ok {
		writeResponseJSON(w, http.StatusNotFound, nil, fmt.Sprintf("table %s doesn't exist", tableName))
		return
	}

	// select all columns
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		colNames = append(colNames, col.Field)
	}
	columns := strings.Join(colNames, ", ")

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d`, columns, tableName, id)

	resp, err := exp.ExecuteQuery(query, colNames)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	if len(resp) == 0 {
		writeResponseJSON(w, http.StatusNotFound, nil, "")
		return
	}

	writeResponseJSON(w, http.StatusOK, resp, "")
}

func (exp *DBExplorer) CreateRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}

	tableName := strings.TrimPrefix(r.URL.Path, "/")
	table, ok := exp.tables[tableName]
	if !ok {
		writeResponseJSON(w, http.StatusNotFound, nil, fmt.Sprintf("table %s doesn't exist", tableName))
		return
	}

	dataToInsert, columnsStr, err := prepareDataToInsert(r, table.Columns)
	if err != nil {
		writeResponseJSON(w, http.StatusBadRequest, nil, err.Error())
		return
	}

	placeholders := "?" + strings.Repeat(", ?", len(dataToInsert)-1)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columnsStr, placeholders)

	_, err = exp.db.Exec(query, dataToInsert...)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, nil, "")
}
