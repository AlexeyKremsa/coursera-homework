package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"reflect"
)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "../api.go", nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

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

			fmt.Printf("Struct name: %s\n", currType.Name)

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				continue
			}

			for _, field := range currStruct.Fields.List {

				if field.Tag != nil {
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					if tag.Get("apivalidator") == "-" {
						continue
					}

					fmt.Printf("fieldName: %s, tag: %s\n", field.Names[0].Name, tag)
				}
			}
		}
	}

	for _, f := range node.Decls {
		fn, ok := f.(*ast.FuncDecl)
		if !ok {
			continue
		}

		fmt.Printf("Func name: %s", fn.Name.Name)
		for _, p := range fn.Type.Params.List {
			for _, n := range p.Names {
				fmt.Printf("Name: %s\n", n.Name)
			}

			fmt.Println("Type: ", p.Type)
		}

		fmt.Println(fn.Doc.Text())
	}
}
