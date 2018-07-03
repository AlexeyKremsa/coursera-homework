package main

import (
	"encoding/json"
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
	exp.router.RegisterRoute("POST", 2, exp.UpdateRecord)
	exp.router.RegisterRoute("DELETE", 2, exp.DeleteRecord)
}

func (exp *DBExplorer) GetAllTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusMethodNotAllowed, "", nil, "bad method")
		return
	}

	rows, err := exp.db.Query("SHOW TABLES")
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
	}

	resp := make([]string, 0)

	for rows.Next() {
		var tableName string

		err = rows.Scan(&tableName)
		if err != nil {
			writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		}

		resp = append(resp, tableName)
	}

	writeResponseJSON(w, http.StatusOK, "tables", resp, "")
}

func (exp *DBExplorer) GetRecordsFromTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusMethodNotAllowed, "", nil, "bad method")
		return
	}

	tableName := strings.Trim(r.URL.Path, "/")
	table, ok := exp.tables[tableName]
	if !ok {
		writeResponseJSON(w, http.StatusNotFound, "", nil, "unknown table")
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

	resp, err := exp.getAll(limit, offset, table)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, "records", resp, "")
}

func (exp *DBExplorer) GetRecordByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusMethodNotAllowed, "", nil, "bad method")
		return
	}

	params := strings.Split(r.URL.Path, "/")
	if len(params) != 3 {
		writeResponseJSON(w, http.StatusBadRequest, "", nil, "invalid URL")
		return
	}

	tableName := params[1]
	id, err := strconv.Atoi(params[2])
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	table, ok := exp.tables[tableName]
	if !ok {
		writeResponseJSON(w, http.StatusNotFound, "", nil, fmt.Sprintf("table %s doesn't exist", tableName))
		return
	}

	resp, err := exp.getByID(id, table)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	if len(resp) == 0 {
		writeResponseJSON(w, http.StatusNotFound, "", nil, "record not found")
		return
	}

	writeResponseJSON(w, http.StatusOK, "record", resp[0], "")
}

func (exp *DBExplorer) CreateRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeResponseJSON(w, http.StatusMethodNotAllowed, "", nil, "bad method")
		return
	}

	tableName := strings.Trim(r.URL.Path, "/")
	table, ok := exp.tables[tableName]
	if !ok {
		writeResponseJSON(w, http.StatusNotFound, "", nil, fmt.Sprintf("table %s doesn't exist", tableName))
		return
	}

	data := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	lastID, err := exp.insert(data, table.Columns, tableName)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, table.Columns[0].Field, lastID, "")
}

func (exp *DBExplorer) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeResponseJSON(w, http.StatusMethodNotAllowed, "", nil, "bad method")
		return
	}

	params := strings.Split(r.URL.Path, "/")
	if len(params) != 3 {
		writeResponseJSON(w, http.StatusBadRequest, "", nil, "invalid URL")
		return
	}

	tableName := params[1]
	id, err := strconv.Atoi(params[2])
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	table, ok := exp.tables[tableName]
	if !ok {
		writeResponseJSON(w, http.StatusNotFound, "", nil, fmt.Sprintf("table %s doesn't exist", tableName))
		return
	}

	rowsAffected, err := exp.delete(id, table)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, "deleted", rowsAffected, "")
}

func (exp *DBExplorer) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusMethodNotAllowed, "", nil, "bad method")
		return
	}

	params := strings.Split(r.URL.Path, "/")
	if len(params) != 3 {
		writeResponseJSON(w, http.StatusBadRequest, "", nil, "invalid URL")
		return
	}

	tableName := params[1]
	id, err := strconv.Atoi(params[2])
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	table, ok := exp.tables[tableName]
	if !ok {
		writeResponseJSON(w, http.StatusNotFound, "", nil, fmt.Sprintf("table %s doesn't exist", tableName))
		return
	}

	data := make(map[string]interface{})
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	err = validateFields(data, table)
	if err != nil {
		writeResponseJSON(w, http.StatusBadRequest, "", nil, err.Error())
		return
	}

	rowsAffected, err := exp.update(id, data, table.Columns, tableName)
	if err != nil {
		writeResponseJSON(w, http.StatusInternalServerError, "", nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, "updated", rowsAffected, "")
}
