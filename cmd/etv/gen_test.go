package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var header = `package main

import (
	"html/template"
)`

//go:generate go test gen_test.go -v -run Generate

func Generate() {
	f, err := os.Create("uiresource.go")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(f, header)

	files, err := filepath.Glob("templates/*.html")
	if err != nil {
		panic(err)
	}

	for _, fname := range files {
		text, err := ioutil.ReadFile(fname)
		if err != nil {
			panic(err)
		}
		base := strings.Replace(filepath.Base(fname), ".html", "", 1)
		fmt.Fprintf(f, "\nconst %sText  = `%s`\n", base, string(text))
	}

	fmt.Fprintln(f, "func init() {")
	for _, fname := range files {
		base := strings.Replace(filepath.Base(fname), ".html", "", 1)
		fmt.Fprintf(f, `	uiT  = template.Must(uiT.New("%s").Parse(%sText))`, base, base)
		fmt.Fprintln(f)
	}
	fmt.Fprintln(f, "}")

	if err := f.Close(); err != nil {
		panic(err)
	}
	fmt.Println("uiresource.go generated")
}

func ExampleGenerate() {
	Generate()
	// Output: uiresource.go generated
}
