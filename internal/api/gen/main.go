package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
	"text/template"
)

type Method struct {
	Name      string
	Params    []Param
	Results   []Result
	HasError  bool   // whether the last result is error
	HasValue  bool   // whether there is a non-error result
	ValueType string // type of the non-error result, if any
}

type Param struct {
	Name        string // may be empty
	Type        string
	Variadic    bool
	LoggingFunc string // "String", "Any", "Int", "Bool", "Float64"
}

type Result struct {
	Name string // may be empty
	Type string
}

func main() {
	var (
		inputFile    string
		outputFile   string
		templateFile string
	)
	flag.StringVar(&inputFile, "input", "", "Input Go file containing interface")
	flag.StringVar(&outputFile, "output", "", "Output Go file")
	flag.StringVar(&templateFile, "template", "", "Template file")
	flag.Parse()

	if inputFile == "" || outputFile == "" || templateFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	methods, err := parseInterface(inputFile, "Client")
	if err != nil {
		log.Fatal(err)
	}

	tmplBytes, err := os.ReadFile(templateFile)
	if err != nil {
		log.Fatal(err)
	}
	funcMap := template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}
	tmpl, err := template.New("wrapper").Funcs(funcMap).Parse(string(tmplBytes))
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, methods); err != nil {
		log.Fatal(err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("warning: failed to format generated code: %v", err)
		formatted = buf.Bytes()
	}

	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated %s with %d methods\n", outputFile, len(methods))
}

func parseInterface(filename, interfaceName string) ([]Method, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var methods []Method
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != interfaceName {
				continue
			}
			iface, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			for _, method := range iface.Methods.List {
				// Assume each method has a single name (no multiple names)
				if len(method.Names) == 0 {
					continue
				}
				name := method.Names[0].Name
				ft, ok := method.Type.(*ast.FuncType)
				if !ok {
					continue
				}
				m := Method{Name: name}
				// Parse parameters
				if ft.Params != nil {
					for _, param := range ft.Params.List {
						variadic := false
						var typ string
						switch t := param.Type.(type) {
						case *ast.Ellipsis:
							variadic = true
							typ = "..." + exprToString(t.Elt)
						default:
							typ = exprToString(t)
						}
						loggingFunc := loggingFunc(typ)
						// If multiple names, create separate params
						if len(param.Names) == 0 {
							m.Params = append(m.Params, Param{Type: typ, Variadic: variadic, LoggingFunc: loggingFunc})
						} else {
							for _, n := range param.Names {
								m.Params = append(m.Params, Param{Name: n.Name, Type: typ, Variadic: variadic, LoggingFunc: loggingFunc})
							}
						}
					}
				}
				// Parse results
				if ft.Results != nil {
					for _, result := range ft.Results.List {
						typ := exprToString(result.Type)
						if len(result.Names) == 0 {
							m.Results = append(m.Results, Result{Type: typ})
						} else {
							for _, n := range result.Names {
								m.Results = append(m.Results, Result{Name: n.Name, Type: typ})
							}
						}
					}
				}
				// Determine if last result is error
				if len(m.Results) > 0 && m.Results[len(m.Results)-1].Type == "error" {
					m.HasError = true
					if len(m.Results) == 2 {
						m.HasValue = true
						m.ValueType = m.Results[0].Type
					}
				}
				methods = append(methods, m)
			}
		}
	}
	return methods, nil
}

// exprToString converts an ast.Expr to a string representation.
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.Ellipsis:
		return "..." + exprToString(e.Elt)
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	case *ast.FuncType:
		// Not needed for our interface
		return "func"
	default:
		// Fallback: print as string
		var buf bytes.Buffer
		fset := token.NewFileSet()
		if err := format.Node(&buf, fset, expr); err != nil {
			return fmt.Sprintf("/* error formatting: %v */", err)
		}
		return buf.String()
	}
}

// loggingFunc returns the appropriate logging field function for a given type.
func loggingFunc(typ string) string {
	// Remove leading "..." if present
	typ = strings.TrimPrefix(typ, "...")
	switch typ {
	case "string":
		return "String"
	case "int", "int32", "int64":
		return "Int"
	case "bool":
		return "Bool"
	case "float64":
		return "Float64"
	default:
		return "Any"
	}
}
