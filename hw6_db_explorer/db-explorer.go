package main

import (
	"database/sql"
	"net/http"

	"log"

	"fmt"

	"github.com/AlexeyKremsa/coursera-homework/hw6_db_explorer/router"
)

type DBExplorer struct {
	db     *sql.DB
	router *router.Router
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	exp := DBExplorer{db: db, router: router.New()}
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
		rows, err := exp.db.Query(fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", t.Name))
		if err != nil {
			log.Fatal(err)
		}

		//columns := make([]*Column, 0)
		for rows.Next() {
			col := &ColumnInfo{}
			err = rows.Scan(&col.Field, &col.Type, &col.Collation, &col.Null, &col.Key, &col.Default, &col.Extra, &col.Privelegies, &col.Comment)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%+v", col)
			return
		}
	}
}
