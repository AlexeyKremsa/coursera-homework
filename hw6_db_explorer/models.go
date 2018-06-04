package main

type DBInfo struct {
	Tables []*Table
}

type Table struct {
	Name    string
	Columns []*ColumnInfo
}

type Column struct {
	Name       string
	Type       string
	IsNullable bool
}

type ColumnInfo struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default *string
	Extra   string
}
