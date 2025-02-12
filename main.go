package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
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

func lockFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_EX) // Exclusive lock
}

func unlockFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN) // Unlock
}

func compile(args []string, f *os.File) {
	for _, a := range args {
		lockFile(f)
		f.WriteString(a)
		f.WriteString("\n")
		unlockFile(f)
	}
	executePass(args[0:])

}

func main() {
	f, _ := os.OpenFile("args", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	args := os.Args[1:]
	compile(args, f)
}
