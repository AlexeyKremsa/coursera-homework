package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
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
	out, err := os.Create("../api_generated.go")
	checkError(err)

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "../api.go", nil, parser.ParseComments)
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
