package main

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
