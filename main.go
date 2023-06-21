package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func executePass(args []string) {
	path := args[0]
	args = args[1:]
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if e := cmd.Run(); e != nil {
		fmt.Println(e)
	}
}

func GetCommandName(args []string) string {
	if len(args) == 0 {
		return ""
	}

	cmd := filepath.Base(args[0])
	if ext := filepath.Ext(cmd); ext != "" {
		cmd = strings.TrimSuffix(cmd, ext)
	}
	return cmd
}

// GetMostInnerAstIdent takes most inner identifier used for
// function call. For a.b.foo(), `b` will be the most inner identifier.
func GetMostInnerAstIdent(inSel *ast.SelectorExpr) *ast.Ident {
	var l []*ast.Ident
	var e ast.Expr
	e = inSel
	for e != nil {
		if _, ok := e.(*ast.Ident); ok {
			l = append(l, e.(*ast.Ident))
			break
		} else if _, ok := e.(*ast.SelectorExpr); ok {
			l = append(l, e.(*ast.SelectorExpr).Sel)
			e = e.(*ast.SelectorExpr).X
		} else if _, ok := e.(*ast.CallExpr); ok {
			e = e.(*ast.CallExpr).Fun
		} else if _, ok := e.(*ast.IndexExpr); ok {
			e = e.(*ast.IndexExpr).X
		} else if _, ok := e.(*ast.UnaryExpr); ok {
			e = e.(*ast.UnaryExpr).X
		} else if _, ok := e.(*ast.ParenExpr); ok {
			e = e.(*ast.ParenExpr).X
		} else if _, ok := e.(*ast.SliceExpr); ok {
			e = e.(*ast.SliceExpr).X
		} else if _, ok := e.(*ast.IndexListExpr); ok {
			e = e.(*ast.IndexListExpr).X
		} else if _, ok := e.(*ast.StarExpr); ok {
			e = e.(*ast.StarExpr).X
		} else if _, ok := e.(*ast.TypeAssertExpr); ok {
			e = e.(*ast.TypeAssertExpr).X
		} else if _, ok := e.(*ast.CompositeLit); ok {
			// TODO dummy implementation
			if len(e.(*ast.CompositeLit).Elts) == 0 {
				e = e.(*ast.CompositeLit).Type
			} else {
				e = e.(*ast.CompositeLit).Elts[0]
			}
		} else if _, ok := e.(*ast.KeyValueExpr); ok {
			e = e.(*ast.KeyValueExpr).Value
		} else {
			// TODO this is uncaught expression
			panic("uncaught expression")
		}
	}
	if len(l) < 2 {
		panic("selector list should have at least 2 elems")
	}
	// caller or receiver is always
	// at position 1, function is at 0
	return l[1]
}

func printPackageInfo(pkg *types.Package) {
	fmt.Printf("Package  %q\n", pkg.Path())
	fmt.Printf("Name:    %s\n", pkg.Name())
	// fmt.Printf("Imports: %s\n", pkg.Imports())
	// fmt.Printf("Scope:   %s\n", pkg.Scope())
}

func parseFile(filePath string, f *os.File) {
	fset := token.NewFileSet()
	_ = fset
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		f.WriteString(err.Error())
		f.WriteString("\n")
	}
	_ = file
	srcPath := filepath.Dir(filePath)
	info := &types.Info{
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(srcPath, fset, []*ast.File{file}, info)
	if err != nil {
		f.WriteString(err.Error())
		f.WriteString("\n")
	}
	_ = pkg

	ast.Inspect(file, func(n ast.Node) bool {
		if funDeclNode, ok := n.(*ast.FuncDecl); ok {
			f.WriteString("FuncDecl:" + file.Name.Name + "." + funDeclNode.Name.String())
			f.WriteString("\n")
			//printPackageInfo(pkg)
			funType := info.Defs[funDeclNode.Name].Type()
			var fTypeStr string
			if funType != nil {
				fTypeStr = funType.String()
			}
			f.WriteString("FuncDecl " +
				fset.Position(funDeclNode.Pos()).String() + " " + funDeclNode.Name.String() + " " + fTypeStr)
			f.WriteString("\n")
		}
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if id, ok := callExpr.Fun.(*ast.Ident); ok {
				f.WriteString("CallExpr:" + file.Name.Name + ":" + id.Name)
				f.WriteString("\n")
				//printPackageInfo(pkg)
				if info.Uses[id] != nil {
					f.WriteString("CallExpr " +
						fset.Position(id.Pos()).String() + " " + id.Name + " " + info.Uses[id].String())
				}
			}
			if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				obj := info.Selections[sel]

				if sel.Sel != nil && info.Uses[sel.Sel] != nil {
					f.WriteString("Qualified " + info.Uses[sel.Sel].Name() + " " + info.Uses[sel.Sel].Type().String())
					f.WriteString("\n")
				}
				if obj != nil {
					f.WriteString("CallExpr " + fset.Position(obj.Obj().Pos()).String() + " " + obj.Obj().Name() +
						" uses sel " + info.Uses[GetMostInnerAstIdent(sel)].Type().String() +
						" " + obj.Obj().Type().String())
					f.WriteString("\n")
				}
			}
		}
		return true
	})
}

func main() {
	prog := `package main
import "fmt"
func main() {fmt.Println("hello")}`
	_ = prog
	f, _ := os.OpenFile("args", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	args := os.Args[1:]
	cmdName := GetCommandName(args)
	if cmdName != "compile" {
		executePass(args[0:])
		return
	}

	argsLen := len(args)
	var destPath string
	for i, a := range args {
		if a == "-o" {
			destPath = filepath.Dir(string(args[i+1]))
			f.WriteString("dest path:" + destPath)
			f.WriteString("\n")
		}
		if a == "-pack" {
			pathReported := false
			for j := i + 1; j < argsLen; j++ {
				// omit -asmhdr switch + following header+
				if string(args[j]) == "-asmhdr" {
					j = j + 2
				}
				if !strings.HasSuffix(args[j], ".go") {
					continue
				}
				filePath := args[j]
				filename := filepath.Base(filePath)
				srcPath := filepath.Dir(filePath)
				if !pathReported {
					f.WriteString("src path:" + srcPath)
					f.WriteString("\n")
					pathReported = true
				}
				f.WriteString(filename)
				f.WriteString("\n")
				if filename == "main.go" {

					out, _ := os.Create(destPath + "/" + "main.go")
					out.WriteString(prog)
					out.Close()
					args[j] = destPath + "/" + "main.go"
				}
				parseFile(filePath, f)
			}
		}
	}
	executePass(args[0:])
}
