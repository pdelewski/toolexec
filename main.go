package main

import (
	"fmt"
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

func main() {
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
	for i, a := range args {
		if a == "-o" {
			f.WriteString("dest path:" + filepath.Dir(string(args[i+1])))
			f.WriteString("\n")
		}
		if a == "-pack" {
			pathReported := false
			for j := i + 1; j < argsLen; j++ {
				if string(args[j]) == "-asmhdr" {
					j = j + 2
				}
				if !strings.HasSuffix(args[j], ".go") {
					continue
				}
				filename := filepath.Base(args[j])
				p := filepath.Dir(args[j])
				_ = p
				if !pathReported {
					f.WriteString("src path:" + p)
					f.WriteString("\n")
					pathReported = true
				}
				f.WriteString(filename)
				f.WriteString("\n")
			}
		}
	}
	executePass(args[0:])
}
