package main

import (
	"database/sql"
	"errors"
	"net/http"

	"log"

	"fmt"

	"reflect"

	"github.com/AlexeyKremsa/coursera-homework/hw6_db_explorer/router"
)

type dBExplorer struct {
	db     *sql.DB
	router *router.Router
	tables map[string]*tableInfo
}

func newDbExplorer(db *sql.DB) (http.Handler, error) {
	exp := dBExplorer{db: db, router: router.New(), tables: make(map[string]*tableInfo)}
	declareRoutes(&exp)
	exp.loadDBInfo()
	return exp.router, nil
}

func (exp *dBExplorer) loadDBInfo() {
	// get tables info
	rows, err := exp.db.Query("SHOW TABLES")
	if err != nil {
		log.Fatal(err)
	}

	tables := make([]*tableInfo, 0)
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}

		t := &tableInfo{
			name: tableName,
		}

		tables = append(tables, t)
	}

	// get columns info
	for _, t := range tables {
		rows, err := exp.db.Query(fmt.Sprintf("SHOW COLUMNS FROM `%s`", t.name))
		if err != nil {
			log.Fatal(err)
		}

		columns := make([]*columnInfo, 0)
		for rows.Next() {
			col := &columnInfo{}
			// need string here to read value and convert it from string to bool
			var isNull string
			err = rows.Scan(&col.field, &col.typeName, &isNull, &col.key, &col.defaultVal, &col.extra)
			if err != nil {
				log.Fatal(err)
			}

			if isNull == "YES" {
				col.isNull = true
			} else {
				col.isNull = false
			}

			columns = append(columns, col)
		}

		t.columns = append(t.columns, columns...)
		exp.tables[t.name] = t
	}
}

func getVariable(varType string) (interface{}, error) {
	var varInt int
	var varString sql.NullString
	switch varType {
	case "INT":
		return &varInt, nil

	case "VARCHAR", "TEXT":
		return &varString, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", varType)
	}
}

func prepareResponse(data []interface{}, colNames []string) (map[string]interface{}, error) {
	resp := make(map[string]interface{}, 0)

	for i := 0; i < len(data); i++ {
		switch v := data[i].(type) {
		case *int:
			resp[colNames[i]] = *v

		case *sql.NullString:
			if v.Valid {
				resp[colNames[i]] = v.String
			} else {
				resp[colNames[i]] = nil
			}

		default:
			return nil, fmt.Errorf("unsupported type: %s", reflect.TypeOf(v).String())
		}
	}

	return resp, nil
}

func validateFields(data map[string]interface{}, table *tableInfo) error {
	if len(data) == 1 {
		_, ok := data["id"]
		if ok {
			return errors.New("field id have invalid type")
		}
	}

	for _, column := range table.columns {
		val, ok := data[column.field]
		if !ok {
			continue
		}

		if val == nil && column.isNull {
			return nil
		}

		if val == nil && !column.isNull {
			return fmt.Errorf("field %s have invalid type", column.field)
		}

		if !compareTypes(column.typeName, reflect.TypeOf(val).Name()) {
			return fmt.Errorf("field %s have invalid type", column.field)
		}
	}

	return nil
}

func compareTypes(colType, fieldType string) bool {
	switch fieldType {
	case "string":
		if colType == "varchar(255)" || colType == "text" {
			return true
		}

	default:
		return false
	}

	return false
}
