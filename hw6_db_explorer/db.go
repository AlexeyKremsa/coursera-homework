package main

import (
	"fmt"
	"net/http"
	"strings"
)

func (exp *DBExplorer) ExecuteQuery(query string, colNames []string) ([]map[string]interface{}, error) {
	rows, err := exp.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colsToRead := make([]interface{}, 0)
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	for _, item := range colTypes {
		// we need proper variables to read data
		col, err := getVariable(item.DatabaseTypeName())
		if err != nil {
			return nil, err
		}
		colsToRead = append(colsToRead, col)
	}

	resp := make([]map[string]interface{}, 0)
	for rows.Next() {
		err = rows.Scan(colsToRead...)
		if err != nil {
			return nil, err
		}

		rowData, err := prepareResponse(colsToRead, colNames)
		if err != nil {
			if err != nil {
				return nil, err
			}
		}

		resp = append(resp, rowData)
	}

	return resp, nil
}

func (exp *DBExplorer) insert(r *http.Request, columns []*ColumnInfo, tableName string) error {
	colNames := make([]string, 0)
	dataToInsert := make([]interface{}, 0)

	// skip 1st element, normally it`s an id
	for i := 1; i < len(columns); i++ {
		colNames = append(colNames, columns[i].Field)

		val := r.FormValue(columns[i].Field)
		if val == "" && !columns[i].Null {
			return fmt.Errorf("%s is empty", columns[i].Field)
		}

		if val == "" && columns[i].Null {
			val = "null"
		}

		dataToInsert = append(dataToInsert, val)
	}

	columnsStr := strings.Join(colNames, ", ")
	placeholders := "?" + strings.Repeat(", ?", len(dataToInsert)-1)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columnsStr, placeholders)
	_, err := exp.db.Exec(query, dataToInsert...)

	return err
}

func (exp *DBExplorer) update(id int, data map[string]string, columns []*ColumnInfo, tableName string) error {
	setStmts := make([]string, 0)
	values := make([]interface{}, 0)

	for k, v := range data {
		setStmts = append(setStmts, fmt.Sprintf("%v = ?", k))
		values = append(values, v)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE ID = %d", tableName, strings.Join(setStmts, ", "), id)
	_, err := exp.db.Exec(query, values...)

	return err
}
