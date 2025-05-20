package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

type Symbol struct {
	Name string
	Type string
	Doc  string
}

type ShimData struct {
	ImportPath string
	Version    string
	Types      []Symbol
	Vars       []Symbol
	Consts     []Symbol
	Funcs      []Symbol
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: make-latest ../../v1.2.3")
		fmt.Println("(creates ../jsonata.go with wrappers from the package in the v1.2.3 directory)")
		os.Exit(1)
	}

	sourcePath := os.Args[1]

	// Extract version from path or use the final directory name
	version := extractVersionFromPath(sourcePath)

	// If version not found in path, use the final directory name
	if version == "" {
		// Get the last part of the path
		version = filepath.Base(sourcePath)
	}

	fmt.Printf("Inspecting jsonata package at: %s (version %s)\n", sourcePath, version)

	// Construct import path for the package
	importPath := fmt.Sprintf("github.com/blues/jsonata-go/%s", version)
	fmt.Printf("Using import path: %s\n", importPath)

	// Get exported symbols
	data, err := getExportedSymbols(importPath, version)
	if err != nil {
		fmt.Printf("Error inspecting package: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d exported symbols\n",
		len(data.Types)+len(data.Vars)+len(data.Consts)+len(data.Funcs))

	err = generateWrapper(data)
	if err != nil {
		fmt.Printf("Error generating wrapper: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully generated jsonata.go wrapper")
}

// extractVersionFromPath tries to extract the version from the path
func extractVersionFromPath(path string) string {
	// Try to find a version pattern like v1.2.3 in the path
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) > 1 {
			// Check if the rest of the string could be a version number
			if _, err := fmt.Sscanf(part[1:], "%f", &struct{ f float64 }{}); err == nil {
				return part
			}
		}
	}
	return ""
}

func getExportedSymbols(importPath string, version string) (*ShimData, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports | packages.NeedName,
		Env:  os.Environ(),
		Dir:  ".", // start from current module
	}

	pkgs, err := packages.Load(cfg, importPath)
	if err != nil {
		return nil, fmt.Errorf("error loading package: %v", err)
	}

	if packages.PrintErrors(pkgs) > 0 || len(pkgs) == 0 {
		return nil, fmt.Errorf("no package found for %s", importPath)
	}

	pkg := pkgs[0].Types

	data := &ShimData{
		ImportPath: importPath,
		Version:    version,
	}

	scope := pkg.Scope()
	for _, name := range scope.Names() {
		// Check if the name starts with an uppercase letter (exported)
		if !ast.IsExported(name) {
			continue
		}

		obj := scope.Lookup(name)
		symbol := Symbol{Name: name}

		switch obj.(type) {
		case *types.TypeName:
			symbol.Type = "Type"
			data.Types = append(data.Types, symbol)
		case *types.Var:
			symbol.Type = "Var"
			data.Vars = append(data.Vars, symbol)
		case *types.Const:
			symbol.Type = "Const"
			data.Consts = append(data.Consts, symbol)
		case *types.Func:
			symbol.Type = "Func"
			data.Funcs = append(data.Funcs, symbol)
		}
	}

	// Sort each category by name
	sort.Slice(data.Types, func(i, j int) bool {
		return data.Types[i].Name < data.Types[j].Name
	})
	sort.Slice(data.Consts, func(i, j int) bool {
		return data.Consts[i].Name < data.Consts[j].Name
	})
	sort.Slice(data.Vars, func(i, j int) bool {
		return data.Vars[i].Name < data.Vars[j].Name
	})
	sort.Slice(data.Funcs, func(i, j int) bool {
		return data.Funcs[i].Name < data.Funcs[j].Name
	})

	return data, nil
}

func generateWrapper(data *ShimData) error {
	// Create template functions
	funcMap := template.FuncMap{
		"join": func(strs []string, sep string) string {
			return strings.Join(strs, sep)
		},
	}

	// Create and parse template
	t := template.New("wrapper").Funcs(funcMap)

	const templateText = `// Copyright Blues Inc.  All rights reserved.

// Package jsonata is a query and transformation language for JSON.
// This is a wrapper package that provides the same interface as the v{{.Version}} package.
// Generated automatically.
package jsonata

import (
	latest "{{.ImportPath}}"
)

// Types
{{range .Types}}// {{.Name}} is a type from the latest JSONata package
type {{.Name}} = latest.{{.Name}}
{{end}}

// Constants
{{range .Consts}}// {{.Name}} is a constant from the latest JSONata package
const {{.Name}} = latest.{{.Name}}
{{end}}

// Variables
{{range .Vars}}// {{.Name}} is a variable from the latest JSONata package
var {{.Name}} = latest.{{.Name}}
{{end}}

// Functions
{{range .Funcs}}// {{.Name}} is a function from the latest JSONata package
var {{.Name}} = latest.{{.Name}}
{{end}}
`

	t, err := t.Parse(templateText)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	// Create output file
	outputPath := filepath.Join("..", "jsonata.go")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file %s: %v", outputPath, err)
	}
	defer file.Close()

	// Execute template
	err = t.Execute(file, data)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}
