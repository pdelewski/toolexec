package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
			f.WriteString(strconv.Itoa(i) + ":" + a)
			f.WriteString("\n")
			f.WriteString(filepath.Dir(string(args[i+1])))
			f.WriteString("\n")
		}
		if a == "-pack" {
			for j := i + 1; j < argsLen; j++ {
				f.WriteString(string(args[j]))
				f.WriteString("\n")
			}
		}
	}
	executePass(args[0:])
}
