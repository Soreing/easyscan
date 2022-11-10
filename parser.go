package easyscan

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Field struct {
	name string
	tag  string
}

func newField(f *ast.Field) (Field, error) {
	if len(f.Names) == 0 || f.Names[0].Name == "" {
		return Field{}, errors.New("field has no name")
	}

	tag := ""
	if f.Tag != nil {
		tag = f.Tag.Value
	}
	return Field{
		name: f.Names[0].Name,
		tag:  tag,
	}, nil
}

type Struct struct {
	name   string
	fields []Field
}

func newStruct(name string, st *ast.StructType) Struct {
	flds := []Field{}
	for _, f := range st.Fields.List {
		if fl, err := newField(f); err == nil {
			flds = append(flds, fl)
		} else {
			fmt.Println(err.Error())
		}
	}
	return Struct{
		name:   name,
		fields: flds,
	}
}

type List struct {
	typeName string
	elemName string
}

func newList(name string, at *ast.ArrayType) (List, error) {
	ename := ""
	if ident, ok := at.Elt.(*ast.Ident); ok {
		if ident.Obj != nil {
			ename = ident.Obj.Name
		}
	}
	if ename == "" {
		return List{}, errors.New("element has no name")
	}
	return List{
		typeName: name,
		elemName: ename,
	}, nil
}

type Parser struct {
	PkgDir   string
	PkgName  string
	Structs  []Struct
	Lists    []List
	AllTypes bool
}

type visitor struct {
	*Parser

	skip bool
	expl bool
}

const (
	explicitComment = "easyscan:explicit"
	skipComment     = "easyscan:skip"
)

func (v *visitor) handleComment(comments *ast.CommentGroup) {
	if comments == nil {
		return
	}

	for _, c := range comments.List {
		comment := c.Text
		if len(comment) < 3 {
			return
		}

		switch comment[1] {
		case '/':
			comment = comment[2:]
		case '*':
			comment = comment[2 : len(comment)-2]
		}

		for _, comment := range strings.Split(comment, "\n") {
			comment = strings.TrimSpace(comment)
			v.skip = v.skip || strings.HasPrefix(comment, skipComment)
			v.expl = v.expl || strings.HasPrefix(comment, explicitComment)
		}
	}
}

func (v *visitor) Visit(n ast.Node) (w ast.Visitor) {
	switch n := n.(type) {
	case *ast.Package:
		return v

	case *ast.GenDecl:
		return v

	case *ast.File:
		v.PkgName = n.Name.String()
		return v

	case *ast.CommentGroup:
		v.handleComment(n)
		return v

	case *ast.TypeSpec:
		if !v.skip && (v.expl || v.AllTypes) {
			name := n.Name.String()
			switch t := n.Type.(type) {
			case *ast.StructType:
				fmt.Println("STRUCT TYPE", name)
				st := newStruct(name, t)
				v.Structs = append(v.Structs, st)
			case *ast.ArrayType:
				fmt.Println("ARRAY TYPE DEF", name)
				if lt, err := newList(name, t); err == nil {
					v.Lists = append(v.Lists, lt)
				}
			}

		}

		v.skip, v.expl = false, false
	}

	return nil
}

func (p *Parser) Parse(fname string, isDir bool) (err error) {
	info, err := os.Stat(fname)
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() {
		p.PkgDir = fname
	} else {
		p.PkgDir = filepath.Dir(fname)
	}

	fset := token.NewFileSet()
	if isDir {
		packages, err := parser.ParseDir(
			fset,
			fname,
			excludeTestFiles,
			parser.ParseComments,
		)
		if err != nil {
			return err
		}
		for _, pckg := range packages {
			ast.Walk(&visitor{Parser: p}, pckg)
		}
	} else {
		f, err := parser.ParseFile(
			fset,
			fname,
			nil,
			parser.ParseComments,
		)
		if err != nil {
			return err
		}
		ast.Walk(&visitor{Parser: p}, f)
	}
	return nil
}

func excludeTestFiles(fi os.FileInfo) bool {
	return !strings.HasSuffix(fi.Name(), "_test.go")
}

