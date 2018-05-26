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

type createObjModel struct {
	StructName string
	Fields     []Field
}

type ApiValidatorTags struct {
	Required      bool
	ParamName     string
	Min           string
	Max           string
	DefaultString string
	DefaultInt    string
	Enum          []string
}

type Fields struct {
	Fields []Field
}

type Field struct {
	Name string
	Type string
	Tag  string
}

type ApigenComment struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}
