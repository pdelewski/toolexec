package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	//	"golang.org/x/tools/go/packages"
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

// LoadMode. Tells about needed information during analysis.
/*
const LoadMode packages.LoadMode = packages.NeedName |
	packages.NeedTypes |
	packages.NeedSyntax |
	packages.NeedTypesInfo |
	packages.NeedFiles |
	packages.NeedImports

func getPkgs(projectPath string, packagePattern string, fset *token.FileSet) ([]*packages.Package, error) {
	var packageSet []*packages.Package

	cfg := &packages.Config{Fset: fset, Mode: LoadMode, Dir: projectPath}
	pkgs, err := packages.Load(cfg, packagePattern)

	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		fmt.Println("\t", pkg)
		packageSet = append(packageSet, pkg)
	}

	return packageSet, nil
}
*/

func parseFile(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		if funDeclNode, ok := n.(*ast.FuncDecl); ok {
			fmt.Println("FuncDecl:", funDeclNode.Name)

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
				filename := filepath.Base(args[j])
				srcPath := filepath.Dir(args[j])
				fset := token.NewFileSet()
				_ = fset
				// pkgs, _ := getPkgs(srcPath, "./...", fset)
				// _ = pkgs
				file, err := parser.ParseFile(fset, args[j], nil, 0)
				if err != nil {
					fmt.Println(err)
				}
				_ = file
				parseFile(file)
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
