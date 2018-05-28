package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

var structHandlers map[string][]handlerTmplModel
var structFields map[string][]Field
var fieldApivalidatorTags map[string]*ApiValidatorTags

func init() {
	structHandlers = make(map[string][]handlerTmplModel)
	structFields = make(map[string][]Field)
	fieldApivalidatorTags = make(map[string]*ApiValidatorTags)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Invalid arguments provided")
	}

	fileToParse := os.Args[1]
	fileToGenerate := os.Args[2]

	if fileToGenerate == "" || fileToParse == "" {
		log.Fatal("Flags can not be empty")
	}

	out, err := os.Create(fileToGenerate)
	checkError(err)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileToParse, nil, parser.ParseComments)
	checkError(err)

	_, err = fmt.Fprintln(out, `package `+node.Name.Name)
	checkError(err)
	_, err = fmt.Fprint(out, imports)
	checkError(err)
	_, err = fmt.Fprintf(out, response, "`json:\"error\"`", "`json:\"response,omitempty\"`")
	checkError(err)

	// Parse structs
	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				continue
			}

			fields := make([]Field, 0)
			fmt.Println(currType.Name)
			for _, field := range currStruct.Fields.List {
				var tag string
				if field.Tag != nil {
					tag = fmt.Sprint(reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]))
				}

				f := Field{
					Name: field.Names[0].Name,
					Type: fmt.Sprint(field.Type),
					Tag:  tag,
				}
				fields = append(fields, f)
				fmt.Printf("fieldName: %s, type: %s, tag: %s\n", field.Names[0].Name, fmt.Sprint(field.Type), tag)
			}

			structFields[fmt.Sprint(currType.Name)] = fields
		}
	}

	// Parse func declarations
	for _, f := range node.Decls {
		fn, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Doc.Text() == "" {
			continue
		}

		apigen, err := parseApigenComment(fn.Doc.Text())
		if err != nil {
			log.Println("Unknown handler tag: ", apigen)
			continue
		}

		if fn.Recv == nil {
			continue
		} else {
			for _, r := range fn.Recv.List {
				receiver := parseReceiverType(fmt.Sprint(r.Type))
				if receiver == "ApiError" || receiver == "" {
					log.Println("Receiver is ApiError or empty. Going to be skipped...")
					continue
				}

				h := handlerTmplModel{}
				h.HandlerName = fn.Name.Name
				h.ReceiverType = receiver
				h.URL = apigen.URL
				h.Method = apigen.Method
				h.IsProtected = apigen.Auth

				handlers, ok := structHandlers[receiver]
				if ok {
					handlers = append(handlers, h)
					structHandlers[receiver] = handlers
				} else {
					handlers = make([]handlerTmplModel, 1)
					handlers[0] = h
					structHandlers[receiver] = handlers
				}

				// 1. Declare a function
				err = funcDeclarationTmpl.Execute(out, h)
				checkError(err)

				// 2. Check if request method is allowed
				checkRequestMethodTmpl(out, h.Method)

				// 3. Authentication
				if h.IsProtected {
					err = authTmpl.Execute(out, nil)
					checkError(err)
				}

				// loop through method params
				for _, p := range fn.Type.Params.List {
					fmt.Println("Type: ", p.Type)
					fmt.Println("Func name: ", fn.Name.Name)

					currentStruct := fmt.Sprint(p.Type)
					if currentStruct == "&{context Context}" {
						continue
					}

					fields, ok := structFields[currentStruct]
					if !ok {
						log.Println("Can't find fields for type: ", currentStruct)
						continue
					}

					// 4. Declare necessary fields
					declareParams(out, fields)

					// 5. Read params either from URL query or form body
					readParams(out, fields, h.Method)

					// 6. Validate params according to rules specified in tags
					validateParams(out, fields)

					// 7. Create an object and call receiver's method
					declareObject(out, currentStruct, fields)

					// 8. Call method
					callMethod(out, &h)
				}
			}
		}
	}

	// Generate ServeHttp
	for k, v := range structHandlers {
		model := serveHttpTmplModel{
			StructName: k,
			Handlers:   v,
		}

		err = serveHttpTmpl.Execute(out, model)
		checkError(err)
	}
}

// parser

func parseReceiverType(name string) string {
	splittedArr := strings.Split(name, " ")
	if len(splittedArr) == 0 {
		return ""
	}

	lastElem := splittedArr[len(splittedArr)-1]
	lastElem = strings.TrimRight(lastElem, "}")

	return lastElem
}

func parseApigenComment(comment string) (*ApigenComment, error) {
	start := strings.Index(comment, "{")
	end := strings.Index(comment, "}")
	finalStr := comment[start : end+1]

	tag := strings.TrimSpace(comment[:start])
	if tag != "apigen:api" {
		return nil, fmt.Errorf("unknown tag: %s", tag)
	}

	apigen := &ApigenComment{}
	err := json.Unmarshal([]byte(finalStr), apigen)
	if err != nil {
		return nil, err
	}

	return apigen, nil
}

func getApivalidatorTag(tag string) ([]string, error) {
	if tag == "" {
		return nil, fmt.Errorf("Empty tag, nothing to parse")
	}
	splitted := strings.Split(tag, ":")
	if splitted[0] != "apivalidator" {
		return nil, fmt.Errorf("Unknown tag: %s", splitted[0])
	}

	rules := strings.Split(strings.Trim(splitted[1], `"`), ",")

	return rules, nil
}

func parseApivalidatorTags(fieldType string, tag string) (*ApiValidatorTags, error) {
	rules, err := getApivalidatorTag(tag)
	if err != nil {
		return nil, err
	}

	tags := &ApiValidatorTags{}

	switch fieldType {
	case "int":
		for _, r := range rules {
			if r == "" {
				return nil, fmt.Errorf("parseApivalidatorInt: Empty rule")
			}

			if r == "required" {
				tags.Required = true
				continue
			}

			if strings.Contains(r, "paramname") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorInt: invalid `paramname` declaration")
				}

				tags.ParamName = splitted[1]
				continue
			}

			if strings.Contains(r, "default") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorInt: invalid `default` declaration")
				}

				tags.DefaultInt = splitted[1]
				continue
			}

			if strings.HasPrefix(r, "min") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorInt: invalid `min` declaration")
				}

				tags.Min = splitted[1]
				continue
			}

			if strings.HasPrefix(r, "max") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorInt: invalid `max` declaration")
				}

				tags.Max = splitted[1]
			}
		}

		return tags, nil

	case "string":
		for _, r := range rules {
			if r == "" {
				return nil, fmt.Errorf("parseApivalidatorString: Empty rule")
			}

			if r == "required" {
				tags.Required = true
				continue
			}

			if strings.Contains(r, "paramname") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `paramname` declaration")
				}

				tags.ParamName = splitted[1]
				continue
			}

			if strings.Contains(r, "default") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `default` declaration")
				}

				tags.DefaultString = splitted[1]
				continue
			}

			if strings.HasPrefix(r, "min") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `min` declaration")
				}

				tags.Min = splitted[1]
				continue
			}

			if strings.HasPrefix(r, "max") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `max` declaration")
				}

				tags.Max = splitted[1]
			}

			if strings.Contains(r, "enum") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `max` declaration")
				}

				roles := strings.Split(splitted[1], "|")
				if len(roles) == 0 || len(roles) > 3 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid enum declaration")
				}

				tags.Enum = roles
			}
		}

		return tags, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", fieldType)
	}
}

// models

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

// templates

var imports = `
import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)
`

var response = `type response struct {
	Error    string      %s
	Response interface{} %s
}

func writeResponseJSON(w http.ResponseWriter, status int, data interface{}, errorText string) {
	w.Header().Set("Content-Type", "application/json")
	resp := response{
		Error:    errorText,
		Response: data,
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err.Error())
	} else {
		w.WriteHeader(status)
		w.Write(jsonResp)
	}
}
`

var serveHttpTmpl = template.Must(template.New("serveHttpTmpl").Parse(`
func (srv *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { {{if .Handlers -}}
		{{- range .Handlers}}
		case "{{.URL}}":
			srv.wrapper{{.HandlerName}}(w, r)
		{{- end}}
	{{- end}}
		default:
			writeResponseJSON(w, http.StatusNotFound, nil, "unknown method")
		}
}
`))

var funcDeclarationTmpl = template.Must(template.New("funcDeclarationTmpl").Parse(`
func (srv *{{.ReceiverType}}) wrapper{{.HandlerName}}(w http.ResponseWriter, r *http.Request) {`))

var declareParamsTmpl = template.Must(template.New("declareParamsTmpl").Parse(`
	
	{{- range .Fields}}
	var {{.Name}} string
	{{- end}}`))

// We assume that all int parameters will have `Int` suffix
var atoiTmpl = template.Must(template.New("atoiTmpl").Parse(`
	{{.FieldName}}Int, err := strconv.Atoi({{.FieldName}})
	if err != nil {
		writeResponseJSON(w, http.StatusBadRequest, nil, strings.ToLower("{{.FieldName}} must be int"))
		return
	}
	`))

var minIntTmpl = template.Must(template.New("minIntTmpl").Parse(`
	if {{.FieldName}}Int < {{.MinValue}} {
		writeResponseJSON(w, http.StatusBadRequest, nil, strings.ToLower("{{.FieldName}} must be >= {{.MinValue}}"))
		return
	}
	`))

var maxIntTmpl = template.Must(template.New("maxIntTmpl").Parse(`
	if {{.FieldName}}Int > {{.MaxValue}} {
		writeResponseJSON(w, http.StatusBadRequest, nil, strings.ToLower("{{.FieldName}} must be <= {{.MaxValue}}"))
		return
	}
	`))

var minStringTmpl = template.Must(template.New("minStringTmpl").Parse(`
	if len({{.FieldName}}) < {{.MinValue}} {
		writeResponseJSON(w, http.StatusBadRequest, nil, strings.ToLower("{{.FieldName}} len must be >= {{.MinValue}}"))
		return
	}
	`))

var maxStringTmpl = template.Must(template.New("maxStringTmpl").Parse(`
	if len({{.FieldName}}) > {{.MinValue}} {
		writeResponseJSON(w, http.StatusBadRequest, nil, strings.ToLower("{{.FieldName}} len must be <= {{.MaxValue}}"))
		return
	}
	`))

var enumTmpl = template.Must(template.New("enumTmpl").Parse(`
	isStatusValid := false
	for _, item := range statusList {
		if item == {{.FieldName}} {
			isStatusValid = true
				break
		}
	}

	if !isStatusValid {
		writeResponseJSON(w, http.StatusBadRequest, nil, fmt.Sprintf("status must be one of [%s]", strings.Join(statusList, ", ")))
		return
	}
`))

var authTmpl = template.Must(template.New(`authTmpl`).Parse(`
	if r.Header.Get("X-Auth") != "100500" {
		writeResponseJSON(w, http.StatusForbidden, nil, "unauthorized")
		return
	}`))

var createObjTmpl = template.Must(template.New(`createObjTmpl`).Parse(`
	paramsToPass := {{.StructName}} {
		{{- range .Fields}}
			{{- if (eq .Type "int")}}
		{{.Name}}: {{.Name}}Int,
			{{- else}}
		{{.Name}}: {{.Name}},
			{{- end}}
		{{- end}}
	}
	`))

var callMethodTmpl = template.Must(template.New(`callMethodTmpl`).Parse(`
	resp, err := srv.{{.HandlerName}}(r.Context(), paramsToPass)
	if err != nil {
		apiErr, ok := err.(ApiError)
		if ok {
			writeResponseJSON(w, apiErr.HTTPStatus, nil, apiErr.Err.Error())
			return
		}

		writeResponseJSON(w, http.StatusInternalServerError, nil, err.Error())
		return
	}

	writeResponseJSON(w, http.StatusOK, resp, "")
}
`))

var getFromQueryParam = "%s = r.URL.Query().Get(`%s`)\n"
var getFromForm = "%s = r.FormValue(`%s`)\n"

func checkRequestMethodTmpl(out *os.File, allowedMethod string) {
	// Both POST and GET allowed
	if allowedMethod == "" {
		return
	}

	if allowedMethod == "GET" {
		_, err := fmt.Fprint(out, `
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}
	`)
		checkError(err)
	}

	if allowedMethod == "POST" {
		_, err := fmt.Fprint(out, `	
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}
	`)
		checkError(err)
	}
}

func declareParams(out *os.File, fields []Field) {
	if len(fields) == 0 {
		log.Fatal("There are no fields to read")
	}

	flds := Fields{
		Fields: fields,
	}

	err := declareParamsTmpl.Execute(out, flds)
	checkError(err)
}

func readParams(out *os.File, fields []Field, httpMethod string) {
	_, err := fmt.Fprintln(out)
	checkError(err)

	if httpMethod == "" {
		readParamsMethodTmpl(out, fields, "GET")
		readParamsMethodTmpl(out, fields, "POST")
	} else {
		readParamsMethodTmpl(out, fields, httpMethod)
	}
}

func readParamsMethodTmpl(out *os.File, fields []Field, httpMethod string) {
	var getParamFrom string
	switch httpMethod {
	case "GET":
		_, err := fmt.Fprintln(out, `
	if r.Method == http.MethodGet {`)
		checkError(err)

		getParamFrom = fmt.Sprintf("       %s", getFromQueryParam)
	case "POST":
		_, err := fmt.Fprintln(out, `
	if r.Method == http.MethodPost {`)
		checkError(err)

		getParamFrom = fmt.Sprintf("       %s", getFromForm)
	default:
		log.Fatal("unsupported http method: ", httpMethod)
	}

	for _, f := range fields {
		if f.Tag == "" {
			_, err := fmt.Fprintf(out, getParamFrom, f.Name, strings.ToLower(f.Name))
			checkError(err)
			continue
		}

		tags, err := parseApivalidatorTags(f.Type, f.Tag)
		checkError(err)

		// we can reuse tags without parsing them again
		// this approach is used to avoid validation code duplicate for post and get methods
		fieldApivalidatorTags[f.Name] = tags

		if tags.ParamName != "" {
			_, err := fmt.Fprintf(out, getParamFrom, f.Name, strings.ToLower(tags.ParamName))
			checkError(err)
		} else {
			_, err := fmt.Fprintf(out, getParamFrom, f.Name, strings.ToLower(f.Name))
			checkError(err)
		}
	}

	_, err := fmt.Fprintln(out, "    }")
	checkError(err)
}

func validateParams(out *os.File, fields []Field) {
	for _, f := range fields {
		if f.Tag == "" {
			continue
		}

		tags, ok := fieldApivalidatorTags[f.Name]
		if !ok {
			continue
		}

		// required
		if tags.Required {
			_, err := fmt.Fprintf(out, `
	if %s == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "%s must me not empty")
		return
	}
	`, f.Name, strings.ToLower(f.Name))
			checkError(err)
		}

		model := minMaxIntTmplModel{
			FieldName: f.Name,
			MinValue:  tags.Min,
			MaxValue:  tags.Max,
		}

		if f.Type == "int" {
			err := atoiTmpl.Execute(out, model)
			checkError(err)
		}

		// default
		if tags.DefaultInt != "" || tags.DefaultString != "" {
			switch f.Type {
			case "int":
				_, err := fmt.Fprintf(out, `
	if %sInt == 0 {
		%sInt = %s
	}
	`, f.Name, f.Name, tags.DefaultInt)
				checkError(err)

			case "string":
				_, err := fmt.Fprintf(out, `
	if %s == "" {
		%s = "%s"
	}
	`, f.Name, f.Name, tags.DefaultString)
				checkError(err)

			default:
				log.Fatalf("Unsupported type: %s", f.Type)
			}
		}

		// min, max
		if tags.Max != "" || tags.Min != "" {
			switch f.Type {
			case "int":

				if tags.Min != "" {
					err := minIntTmpl.Execute(out, model)
					checkError(err)
				}

				if tags.Max != "" {
					err := maxIntTmpl.Execute(out, model)
					checkError(err)
				}

			case "string":
				if tags.Min != "" {
					err := minStringTmpl.Execute(out, model)
					checkError(err)
				}

				if tags.Max != "" {
					err := maxStringTmpl.Execute(out, model)
					checkError(err)
				}

			default:
				log.Fatalf("Unsupported type: %s", f.Type)
			}
		}

		// enum
		if len(tags.Enum) != 0 {
			model := enumTmplModel{
				FieldName: f.Name,
				Enum:      tags.Enum,
			}

			// declare array with enums
			_, err := fmt.Fprintf(out, fmt.Sprintf("statusList := %#v", tags.Enum))
			checkError(err)

			err = enumTmpl.Execute(out, model)
			checkError(err)
		}
	}
}

func declareObject(out *os.File, structName string, fields []Field) {
	model := createObjModel{
		StructName: structName,
		Fields:     fields,
	}

	err := createObjTmpl.Execute(out, model)
	checkError(err)
}

func callMethod(out *os.File, h *handlerTmplModel) {
	err := callMethodTmpl.Execute(out, h)
	checkError(err)
}
