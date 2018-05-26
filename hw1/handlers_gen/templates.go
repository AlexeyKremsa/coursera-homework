package main

import (
	"fmt"
	"log"
	"os"
	"text/template"
)

var imports = `import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)`

var response = `
type response struct {
	Error    string      "json:'error'"
	Response interface{} "json:'response,omitempty'"
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
}`

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
	}`))

var funcDeclarationTmpl = template.Must(template.New("funcDeclarationTmpl").Parse(`
func (srv *{{.ReceiverType}}) wrapper{{.HandlerName}}(w http.ResponseWriter, r *http.Request) {`))

var declareParamsTmpl = template.Must(template.New("declareParamsTmpl").Parse(`
	{{- range .Fields}}
	var {{.Name}} {{.Type}}
	{{- end}}`))

// We assume that all int parameters will have `Int` suffix
var atoiTmpl = template.Must(template.New("atoiTmpl").Parse(`
	{{.FieldName}}Int, err := strconv.Atoi({{.FieldName}})
	if err != nil {
		writeResponseJSON(w, http.StatusBadRequest, nil, "{{.FieldName}} must be int")
		return
	}
	`))

var minIntTmpl = template.Must(template.New("minIntTmpl").Parse(`
	if {{.FieldName}}Int < {{.MinValue}} {
		writeResponseJSON(w, http.StatusBadRequest, nil, "{{.FieldName}} must be >= {{.MinValue}}")
		return
	}
	`))

var maxIntTmpl = template.Must(template.New("maxIntTmpl").Parse(`
	if {{.FieldName}}Int > {{.MaxValue}} {
		writeResponseJSON(w, http.StatusBadRequest, nil, "{{.FieldName}} must be <= {{.MaxValue}}")
		return
	}
	`))

var minStringTmpl = template.Must(template.New("minStringTmpl").Parse(`
	if len({{.FieldName}}) < {{.MinValue}} {
		writeResponseJSON(w, http.StatusBadRequest, nil, "{{.FieldName}} must be more than {{.MinValue}} characters")
		return
	}
	`))

var maxStringTmpl = template.Must(template.New("maxStringTmpl").Parse(`
	if len({{.FieldName}}) > {{.MinValue}} {
		writeResponseJSON(w, http.StatusBadRequest, nil, "{{.FieldName}} must be less than {{.MaxValue}} characters")
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
		writeResponseJSON(w, http.StatusBadRequest, nil, "unknown status: {{.FieldName}}")
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

var getFromQueryParam = "%s = r.URL.Query().Get(`%s`)\n"
var getFromForm = "%s = r.FormValue(`%s`)\n"

func checkRequestMethodTmpl(out *os.File, allowedMethod string) {
	// Both POST and GET allowed
	if allowedMethod == "" {
		return
	}

	if allowedMethod == "GET" {
		fmt.Fprint(out, `
	if r.Method != http.MethodGet {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}
	`)
	}

	if allowedMethod == "POST" {
		fmt.Fprint(out, `	
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}
	`)
	}
}

func declareParams(out *os.File, fields []Field) {
	fmt.Fprintln(out)

	if len(fields) == 0 {
		log.Fatal("There are no fields to read")
	}

	flds := Fields{
		Fields: fields,
	}

	err := declareParamsTmpl.Execute(out, flds)
	if err != nil {
		log.Fatal(err)
	}
}

func readParams(out *os.File, fields []Field, httpMethod string) {
	fmt.Fprintln(out)

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
		fmt.Fprintln(out, `
	if r.Method == http.MethodGet {`)
		getParamFrom = fmt.Sprintf("       %s", getFromQueryParam)
	case "POST":
		fmt.Fprintln(out, `
	if r.Method == http.MethodPost {`)
		getParamFrom = fmt.Sprintf("       %s", getFromForm)
	default:
		log.Fatal("unsupported http method: ", httpMethod)
	}

	for _, f := range fields {
		if f.Tag == "" {
			fmt.Fprintf(out, getParamFrom, f.Name, f.Name)
			continue
		}

		tags, err := parseApivalidatorTags(f.Type, f.Tag)
		if err != nil {
			log.Fatal(err)
		}

		// we can reuse tags without parsing them again
		// this approach is used to avoid validation code duplicate for post and get methods
		fieldApivalidatorTags[f.Name] = tags

		if tags.ParamName != "" {
			fmt.Fprintf(out, getParamFrom, f.Name, tags.ParamName)
		} else {
			fmt.Fprintf(out, getParamFrom, f.Name, f.Name)
		}
	}

	fmt.Fprintln(out, "    }")
}

func validateParams(out *os.File, fields []Field) {
	for _, f := range fields {
		if f.Tag == "" {
			continue
		}

		tags, ok := fieldApivalidatorTags[f.Name]
		if !ok {
			log.Println("Can't find apivalidator tag for field: ", f.Name)
			continue
		}

		// required
		if tags.Required {
			fmt.Fprintf(out, `
	if %s == "" {
		writeResponseJSON(w, http.StatusBadRequest, nil, "%s must me not empty")
		return
	}
	`, f.Name, f.Name)
		}

		model := minMaxIntTmplModel{
			FieldName: f.Name,
			MinValue:  tags.Min,
			MaxValue:  tags.Max,
		}

		if f.Type == "int" {
			err := atoiTmpl.Execute(out, model)
			if err != nil {
				log.Fatal("atoiTmpl: ", err)
			}
		}

		// default
		if tags.DefaultInt != "" || tags.DefaultString != "" {
			switch f.Type {
			case "int":
				fmt.Fprintf(out, `
	if %sInt == 0 {
		%sInt = %s
	}
	`, f.Name, f.Name, tags.DefaultInt)

			case "string":
				fmt.Fprintf(out, `
	if %s == "" {
		%s = %s
	}
	`, f.Name, f.Name, tags.DefaultString)

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
					if err != nil {
						log.Fatal("minIntTmpl: ", err)
					}
				}

				if tags.Max != "" {
					err := maxIntTmpl.Execute(out, model)
					if err != nil {
						log.Fatal("maxIntTmpl: ", err)
					}
				}

			case "string":
				if tags.Min != "" {
					err := minStringTmpl.Execute(out, model)
					if err != nil {
						log.Fatal("minStringTmpl: ", err)
					}
				}

				if tags.Max != "" {
					err := maxStringTmpl.Execute(out, model)
					if err != nil {
						log.Fatal("maxStringTmpl: ", err)
					}
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
			fmt.Fprintf(out, fmt.Sprintf("statusList := %#v", tags.Enum))

			err := enumTmpl.Execute(out, model)
			if err != nil {
				log.Fatal("enumTmpl: ", err.Error())
			}
		}
	}
}

func declareObject(out *os.File, structName string, fields []Field) {
	model := createObjModel{
		StructName: structName,
		Fields:     fields,
	}

	err := createObjTmpl.Execute(out, model)
	if err != nil {
		log.Fatal("declareObject: ", err.Error())
	}
}
