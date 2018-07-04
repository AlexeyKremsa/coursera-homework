package main

type tableInfo struct {
	name    string
	columns []*columnInfo
}

type columnInfo struct {
	field      string
	typeName   string
	isNull     bool
	key        string
	defaultVal *string
	extra      string
}
