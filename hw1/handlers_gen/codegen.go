package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"strings"
)

const (
	imports string = `import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)`

	response = `
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
)

func init() {
	structHandlers = make(map[string][]handlerTmpl)
}

type serveHttpTmplModel struct {
	StructName string
	Handlers   []handlerTmpl
}

var structHandlers map[string][]handlerTmpl

type handlerTmpl struct {
	HandlerName  string
	ReceiverType string
	URL          string
	Method       string
	IsProtected  bool
}

var serveHttpTmpl = template.Must(template.New("serveHttpTmpl").Parse(`
func (api *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path { {{if .Handlers -}}
	{{- range .Handlers}}
	case "{{.URL}}":
		api.wrapper{{.HandlerName}}(w, r)
	{{- end}}
{{- end}}
	default:
		writeResponseJSON(w, http.StatusNotFound, nil, "unknown method")
	}
}`))

func main() {
	out, err := os.Create("../api_generated.go")
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "../api.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(out, `package `+node.Name.Name)
	fmt.Fprintln(out) // empty line
	fmt.Fprintln(out, imports)
	fmt.Fprintln(out, response)

	// Show output
	//for _, f := range node.Decls {
	//	g, ok := f.(*ast.GenDecl)
	//	if !ok {
	//		continue
	//	}
	//
	//	for _, spec := range g.Specs {
	//		currType, ok := spec.(*ast.TypeSpec)
	//		if !ok {
	//			continue
	//		}
	//
	//		fmt.Printf("Struct name: %s\n", currType.Name)
	//
	//		currStruct, ok := currType.Type.(*ast.StructType)
	//		if !ok {
	//			continue
	//		}
	//
	//		for _, field := range currStruct.Fields.List {
	//
	//			if field.Tag != nil {
	//				tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
	//				if tag.Get("apivalidator") == "-" {
	//					continue
	//				}
	//
	//				fmt.Printf("fieldName: %s, tag: %s\n", field.Names[0].Name, tag)
	//			}
	//		}
	//	}
	//}

	for _, f := range node.Decls {
		fn, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Recv == nil {
			continue
		} else {
			for _, r := range fn.Recv.List {
				str := parseReceiverType(fmt.Sprint(r.Type))
				if str == "ApiError" || str == "" {
					continue
				}

				apigen, err := parseApigenComment(fn.Doc.Text())
				if err != nil {
					log.Println("Unknown tag: ", apigen)
					continue
				}

				h := handlerTmpl{}
				h.HandlerName = fn.Name.Name
				h.ReceiverType = str
				h.URL = apigen.URL
				h.Method = apigen.Method
				h.IsProtected = apigen.Auth

				handlers, ok := structHandlers[str]
				if ok {
					handlers = append(handlers, h)
					structHandlers[str] = handlers
				} else {
					handlers = make([]handlerTmpl, 1)
					handlers[0] = h
					structHandlers[str] = handlers
				}
			}
		}

		//for _, p := range fn.Type.Params.List {
		//	for _, n := range p.Names {
		//		fmt.Printf("Name: %s\n", n.Name)
		//	}
		//
		//	fmt.Println("Type: ", p.Type)
		//}
		//
		//fmt.Println(fn.Doc.Text())
	}

	for k, v := range structHandlers {
		model := serveHttpTmplModel{
			StructName: k,
			Handlers:   v,
		}

		err = serveHttpTmpl.Execute(out, model)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func parseReceiverType(name string) string {
	splittedArr := strings.Split(name, " ")
	if len(splittedArr) == 0 {
		return ""
	}

	lastElem := splittedArr[len(splittedArr)-1]
	lastElem = strings.TrimRight(lastElem, "}")

	return lastElem
}

type ApigenComment struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
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
