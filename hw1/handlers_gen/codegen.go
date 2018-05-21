package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
)

var structHandlers map[string][]handlerTmpl
var structFields map[string][]Field
var fieldApivalidatorTags map[string]*ApiValidatorTags

func init() {
	structHandlers = make(map[string][]handlerTmpl)
	structFields = make(map[string][]Field)
	fieldApivalidatorTags = make(map[string]*ApiValidatorTags)
}

type ApiValidatorTags struct {
	Required      bool
	ParamName     string
	Min           int
	Max           int
	DefaultString string
	DefaultInt    int
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

type serveHttpTmplModel struct {
	StructName string
	Handlers   []handlerTmpl
}

type handlerTmpl struct {
	HandlerName  string
	ReceiverType string
	URL          string
	Method       string
	IsProtected  bool
}

type ApigenComment struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

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
				// if tag.Get("apivalidator") == "" {
				// 	continue
				// }

				f := Field{
					Name: strings.ToLower(field.Names[0].Name),
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

				h := handlerTmpl{}
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
					handlers = make([]handlerTmpl, 1)
					handlers[0] = h
					structHandlers[receiver] = handlers
				}

				// 1. Declare a function
				err = funcDeclarationTmpl.Execute(out, h)
				if err != nil {
					log.Fatal(err)
				}

				// 2. Check if request method is allowed
				checkRequestMethodTmpl(out, h.Method)

				// for _, p := range fn.Type.Params.List {
				// 	fmt.Println("Type: ", p.Type)
				// 	fmt.Println("Func name: ", fn.Name.Name)
				// }

				// fields, ok := structFields[h.ReceiverType]
				// if !ok || len(fields) == 0 {
				// 	log.Println("Can't declare fields for type: ", h.ReceiverType)
				// 	continue
				// }
				// declareParams(out, fields)
				// fmt.Fprintln(out) // empty line

				for _, p := range fn.Type.Params.List {
					fmt.Println("Type: ", p.Type)
					fmt.Println("Func name: ", fn.Name.Name)

					argType := fmt.Sprint(p.Type)
					if argType == "&{context Context}" {
						continue
					}

					fields, ok := structFields[argType]
					if !ok {
						log.Println("Can't find fields for type: ", argType)
						continue
					}

					// 3. Declare necessary fields
					declareParams(out, fields)

					// 4. Read params either from URL query or form body
					readParams(out, fields, h.Method)

					// 5. Validate params according to rules specified in tags
					validateParams(out, fields)
				}

				fmt.Fprintln(out) // empty line
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
		if err != nil {
			log.Fatal(err)
		}

		// for _, h := range v {
		// 	err := funcDeclarationTmpl.Execute(out, h)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// }

		fmt.Fprintln(out) // empty line
	}
}
