package main

type DBInfo struct {
	Tables []*Table
}

type Table struct {
	Name    string
	Columns []*ColumnInfo
}

type ColumnInfo struct {
	Field   string
	Type    string
	Null    bool
	Key     string
	Default *string
	Extra   string
}
