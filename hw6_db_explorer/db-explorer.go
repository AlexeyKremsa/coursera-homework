package main

import (
	"database/sql"
	"net/http"

	"log"

	"fmt"

	"reflect"

	"github.com/AlexeyKremsa/coursera-homework/hw6_db_explorer/router"
)

type DBExplorer struct {
	db     *sql.DB
	router *router.Router
	tables map[string]*Table
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	exp := DBExplorer{db: db, router: router.New(), tables: make(map[string]*Table)}
	DeclareRoutes(&exp)
	exp.loadDBInfo()
	return exp.router, nil
}

func (exp *DBExplorer) loadDBInfo() {
	// get tables info
	rows, err := exp.db.Query("SHOW TABLES")
	if err != nil {
		log.Fatal(err)
	}

	tables := make([]*Table, 0)
	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			log.Fatal(err)
		}

		t := &Table{
			Name: tableName,
		}

		tables = append(tables, t)
	}

	// get columns info
	for _, t := range tables {
		rows, err := exp.db.Query(fmt.Sprintf("SHOW COLUMNS FROM `%s`", t.Name))
		if err != nil {
			log.Fatal(err)
		}

		columns := make([]*ColumnInfo, 0)
		for rows.Next() {
			col := &ColumnInfo{}
			// need string here to read value and convert it from string to bool
			var isNull string
			err = rows.Scan(&col.Field, &col.Type, &isNull, &col.Key, &col.Default, &col.Extra)
			if err != nil {
				log.Fatal(err)
			}

			if isNull == "YES" {
				col.Null = true
			} else {
				col.Null = false
			}

			columns = append(columns, col)
		}

		t.Columns = append(t.Columns, columns...)
		exp.tables[t.Name] = t
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
