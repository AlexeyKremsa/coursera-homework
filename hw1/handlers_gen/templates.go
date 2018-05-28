package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

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
		checkError(errors.Wrap(err, "checkRequestMethodTmpl"))
	}

	if allowedMethod == "POST" {
		_, err := fmt.Fprint(out, `	
	if r.Method != http.MethodPost {
		writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
		return
	}
	`)
		checkError(errors.Wrap(err, "checkRequestMethodTmpl"))
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
	checkError(errors.Wrap(err, "declareParams"))
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
		checkError(errors.Wrap(err, "readParamsMethodTmpl"))

		getParamFrom = fmt.Sprintf("       %s", getFromQueryParam)
	case "POST":
		_, err := fmt.Fprintln(out, `
	if r.Method == http.MethodPost {`)
		checkError(errors.Wrap(err, "readParamsMethodTmpl"))

		getParamFrom = fmt.Sprintf("       %s", getFromForm)
	default:
		log.Fatal("unsupported http method: ", httpMethod)
	}

	for _, f := range fields {
		if f.Tag == "" {
			_, err := fmt.Fprintf(out, getParamFrom, f.Name, strings.ToLower(f.Name))
			checkError(errors.Wrap(err, "readParamsMethodTmpl"))
			continue
		}

		tags, err := parseApivalidatorTags(f.Type, f.Tag)
		checkError(errors.Wrap(err, "readParamsMethodTmpl"))

		// we can reuse tags without parsing them again
		// this approach is used to avoid validation code duplicate for post and get methods
		fieldApivalidatorTags[f.Name] = tags

		if tags.ParamName != "" {
			_, err := fmt.Fprintf(out, getParamFrom, f.Name, strings.ToLower(tags.ParamName))
			checkError(errors.Wrap(err, "readParamsMethodTmpl"))
		} else {
			_, err := fmt.Fprintf(out, getParamFrom, f.Name, strings.ToLower(f.Name))
			checkError(errors.Wrap(err, "readParamsMethodTmpl"))
		}
	}

	_, err := fmt.Fprintln(out, "    }")
	checkError(errors.Wrap(err, "readParamsMethodTmpl"))
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
			checkError(errors.Wrap(err, "validateParams"))
		}

		model := minMaxIntTmplModel{
			FieldName: f.Name,
			MinValue:  tags.Min,
			MaxValue:  tags.Max,
		}

		if f.Type == "int" {
			err := atoiTmpl.Execute(out, model)
			checkError(errors.Wrap(err, "atoiTmpl"))
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
				checkError(errors.Wrap(err, "validateParams"))

			case "string":
				_, err := fmt.Fprintf(out, `
	if %s == "" {
		%s = "%s"
	}
	`, f.Name, f.Name, tags.DefaultString)
				checkError(errors.Wrap(err, "validateParams"))

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
					checkError(errors.Wrap(err, "minIntTmpl"))
				}

				if tags.Max != "" {
					err := maxIntTmpl.Execute(out, model)
					checkError(errors.Wrap(err, "maxIntTmpl"))
				}

			case "string":
				if tags.Min != "" {
					err := minStringTmpl.Execute(out, model)
					checkError(errors.Wrap(err, "minStrngTmpl"))
				}

				if tags.Max != "" {
					err := maxStringTmpl.Execute(out, model)
					checkError(errors.Wrap(err, "maxStringTmpl"))
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
			checkError(errors.Wrap(err, "validateParams"))

			err = enumTmpl.Execute(out, model)
			checkError(errors.Wrap(err, "enumTmpl"))
		}
	}
}

func declareObject(out *os.File, structName string, fields []Field) {
	model := createObjModel{
		StructName: structName,
		Fields:     fields,
	}

	err := createObjTmpl.Execute(out, model)
	checkError(errors.Wrap(err, "declareObject"))
}

func callMethod(out *os.File, h *handlerTmplModel) {
	err := callMethodTmpl.Execute(out, h)
	checkError(errors.Wrap(err, "callMethod"))
}
