package main

import (
	"net/http"
)

func DeclareRoutes(exp *DBExplorer) {
	exp.router.RegisterRoute("/", exp.GetAllTables, "GET")
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

//
//func (exp *DBExplorer) GetRecordsFromTable(w http.ResponseWriter, r *http.Request) {
//	if r.Method != http.MethodGet {
//		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
//		return
//	}
//
//	limit := r.URL.Query().Get("limit")
//	if limit == "" {
//		limit = "5"
//	}
//
//	offset := r.URL.Query().Get("offset")
//	if offset == "" {
//		offset = "0"
//	}
//
//	query := fmt.Sprintf(`SELECT * FROM %s LIMIT %s, %s`, r.URL.Path, limit, offset)
//
//	rows, err := exp.db.Query(query)
//	if err != nil {
//		writeResponseJSON(w, http.StatusInternalServerError, err.Error(), "db error")
//	}
//
//}
