package main

type serveHttpTmplModel struct {
	StructName string
	Handlers   []handlerTmplModel
}

type handlerTmplModel struct {
	HandlerName  string
	ReceiverType string
	URL          string
	Method       string
	IsProtected  bool
}

type minMaxIntTmplModel struct {
	FieldName string
	MinValue  string
	MaxValue  string
}

type enumTmplModel struct {
	FieldName string
	Enum      []string
}
