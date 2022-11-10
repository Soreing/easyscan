package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	es "github.com/Soreing/easyscan"
)

var specifiedName = flag.String("output_filename", "", "specify the filename of the output")
var allTypes = flag.Bool("all", false, "generate scanners for all types in a file")
var lowerCase = flag.Bool("lower_case", false, "use lower case names by default")
var camelCase = flag.Bool("camel_case", false, "use camel case names by default")
var kebabCase = flag.Bool("kebab_case", false, "use kebab case names by default")
var snakeCase = flag.Bool("snake_case", false, "use snake case names by default")
var pascalCase = flag.Bool("pascal_case", false, "use pascal case names by default")
var anyOrder = flag.Bool("any_order", false, "allow scanning fields in any order")

func main() {
	flag.Parse()
	files := flag.Args()

	if len(files) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, file := range files {
		if err := generate(file); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func generate(fname string) error {
	finf, err := os.Stat(fname)
	if err != nil {
		return err
	}

	p := es.Parser{AllTypes: *allTypes}
	if err := p.Parse(fname, finf.IsDir()); err != nil {
		return fmt.Errorf("error parsing %v: %v", fname, err)
	}

	var outName string
	if *specifiedName != "" {
		outName = *specifiedName
	} else if finf.IsDir() {
		outName = filepath.Join(fname, p.PkgName+"_easyscan.go")
	} else {
		outName = strings.TrimSuffix(fname, ".go") + "_easyscan.go"
	}

	g := es.NewGenerator()

	g.SetPackage(p.PkgName)
	g.SetAnyOrder(*anyOrder)

	if *lowerCase {
		g.SetDefaultCase(es.LOWER_CASE)
	} else if *camelCase {
		g.SetDefaultCase(es.CAMEL_CASE)
	} else if *kebabCase {
		g.SetDefaultCase(es.KEBAB_CASE)
	} else if *snakeCase {
		g.SetDefaultCase(es.SNAKE_CASE)
	} else if *pascalCase {
		g.SetDefaultCase(es.PASCAL_CASE)
	} else {
		g.SetDefaultCase(es.LOWER_CASE)
	}

	g.WriteHeader()

	for _, st := range p.Structs {
		g.AddScanStruct(st)
	}
	for _, lt := range p.Lists {
		g.AddScanList(lt)
	}

	src := g.ReadAll()
	err = ioutil.WriteFile(outName, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}

	return nil
}

