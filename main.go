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

func printPackageInfo(pkg *types.Package) {
	fmt.Printf("Package  %q\n", pkg.Path())
	fmt.Printf("Name:    %s\n", pkg.Name())
	// fmt.Printf("Imports: %s\n", pkg.Imports())
	// fmt.Printf("Scope:   %s\n", pkg.Scope())
}

func parseFile(filePath string) {
	fset := token.NewFileSet()
	_ = fset
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		fmt.Println(err)
	}
	_ = file
	srcPath := filepath.Dir(filePath)
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(srcPath, fset, []*ast.File{file}, nil)
	_ = pkg

	ast.Inspect(file, func(n ast.Node) bool {
		if funDeclNode, ok := n.(*ast.FuncDecl); ok {
			fmt.Println("FuncDecl:", file.Name.Name, ":", funDeclNode.Name)
			printPackageInfo(pkg)
			//	fmt.Println("Def:", pkg.TypesInfo.Defs[funDeclNode.Name].Name())

		}
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if id, ok := callExpr.Fun.(*ast.Ident); ok {
				fmt.Println("CallExpr:", file.Name.Name, ":", id)
				printPackageInfo(pkg)

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
	for a := range args {
		_ = a
	}
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
				parseFile(filePath)
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
			}
		}
	}
	executePass(args[0:])
}
