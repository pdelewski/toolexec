package main

import (
	"fmt"
	"go/ast"
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

func inspectFuncs(file *ast.File, fset *token.FileSet, info *types.Info, f *os.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		if funDeclNode, ok := n.(*ast.FuncDecl); ok {
			f.WriteString("FuncDecl:" + fset.Position(funDeclNode.Pos()).String() + file.Name.Name + "." + funDeclNode.Name.String())
			f.WriteString("\n")
		}
		return true
	})
}

func analyzePackage(filePaths []string, f *os.File) {
	fset := token.NewFileSet()
	_ = fset
	info := &types.Info{
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
	var files []*ast.File
	for _, filePath := range filePaths {
		file, err := parser.ParseFile(fset, filePath, nil, 0)
		if err != nil {
			f.WriteString(err.Error())
			f.WriteString("\n")
		}
		files = append(files, file)
	}

	for _, file := range files {
		inspectFuncs(file, fset, info, f)
	}
}

func compile(args []string, f *os.File) {
	prog := `package main
import "fmt"
func main() {fmt.Println("hello")}`
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
			var files []string
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
				files = append(files, filePath)
			}
			analyzePackage(files, f)

		}
	}
	executePass(args[0:])

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
	compile(args, f)
}
