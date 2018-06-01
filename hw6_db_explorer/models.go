package main

type DBInfo struct {
	Tables []*Table
}

type Table struct {
	Name    string
	Columns []*Column
}

type Column struct {
	Name       string
	Type       string
	IsNullable bool
}

type ColumnInfo struct {
	Field       string
	Type        string
	Collation   *string
	Null        string
	Key         string
	Default     *string
	Extra       string
	Privelegies string
	Comment     string
}
