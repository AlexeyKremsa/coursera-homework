package main

import (
	"fmt"
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

func (exp *DBExplorer) insert(data map[string]interface{}, columns []*ColumnInfo, tableName string) (int64, error) {
	colNames := make([]string, 0)
	values := make([]interface{}, 0)

	// skip 1st element, normally it`s an id
	for i := 1; i < len(columns); i++ {
		colNames = append(colNames, columns[i].Field)

		val, ok := data[columns[i].Field]
		if !ok {
			if columns[i].Null {
				val = nil
			} else {
				val = ""
			}
		}

		values = append(values, val)
	}

	columnsStr := strings.Join(colNames, ", ")
	placeholders := "?" + strings.Repeat(", ?", len(values)-1)
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, columnsStr, placeholders)
	res, err := exp.db.Exec(query, values...)
	if err != nil {
		return -1, err
	}

	return res.LastInsertId()
}

func (exp *DBExplorer) update(id int, data map[string]interface{}, columns []*ColumnInfo, tableName string) (int64, error) {
	setStmts := make([]string, 0)
	values := make([]interface{}, 0)

	for k, v := range data {
		setStmts = append(setStmts, fmt.Sprintf("%v = ?", k))
		values = append(values, v)
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %d", tableName, strings.Join(setStmts, ", "), columns[0].Field, id)
	fmt.Println(query)
	res, err := exp.db.Exec(query, values...)
	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}

func (exp *DBExplorer) getAll(limit, offset int64, table *Table) ([]map[string]interface{}, error) {
	// select all columns
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		colNames = append(colNames, col.Field)
	}
	columns := strings.Join(colNames, ", ")

	query := fmt.Sprintf(`SELECT %s FROM %s LIMIT %d OFFSET %d`, columns, table.Name, limit, offset)
	resp, err := exp.ExecuteQuery(query, colNames)

	return resp, err
}

func (exp *DBExplorer) getByID(id int, table *Table) ([]map[string]interface{}, error) {
	// select all columns
	colNames := make([]string, 0)
	for _, col := range table.Columns {
		colNames = append(colNames, col.Field)
	}
	columns := strings.Join(colNames, ", ")

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE %s = %d`, columns, table.Name, table.Columns[0].Field, id)

	resp, err := exp.ExecuteQuery(query, colNames)

	return resp, err
}

func (exp *DBExplorer) delete(id int, table *Table) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table.Name, table.Columns[0].Field)

	res, err := exp.db.Exec(query, id)
	if err != nil {
		return -1, err
	}

	return res.RowsAffected()
}
