package main

import (
	"bufio"
	"fmt"
	"github.com/yyotti/golame"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	cpu := runtime.NumCPU()
	if cpu == 1 {
		runtime.GOMAXPROCS(2)
	} else {
		runtime.GOMAXPROCS(cpu)
	}

	_, err := exec.LookPath("lame")
	if err != nil {
		fmt.Println("`lame` command not found.")
		os.Exit(1)
	}

	lame := golame.Lame{Out: os.Stdout, Err: os.Stderr}
	exitCode := lame.Run(os.Args[1:])

	fmt.Println("\nPress ENTER to quit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	os.Exit(exitCode)
}
