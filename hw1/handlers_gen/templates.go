package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
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
		}`)
	}

	if allowedMethod == "POST" {
		fmt.Fprint(out, `	
		if r.Method != http.MethodPost {
			writeResponseJSON(w, http.StatusNotAcceptable, nil, "bad method")
			return
		}`)
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
	if err != nil {
		log.Fatal(err)
	}
}

// func readParams(out *os.File, fields []Field, httpMethod string) {
// 	for _, field := range fields {
// 		switch httpMethod {
// 		case "GET":
// 			fmt.Fprint(out, `if r.Method == http.MethodGet {`)
// 			for _, f := range fields {
// 				if f.Tag == "" {
// 					fmt.Fprintf(out, `%s = r.URL.Query().Get("%s")`, f.Name, f.Name)
// 				}

// 			}
// 		default:
// 			log.Fatal("Unsupported http method: ", httpMethod)
// 		}
// 	}
// }
