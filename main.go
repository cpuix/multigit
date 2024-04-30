package main

import (
	"fmt"
	flagparser "github.com/cpuix/multigit/internal/flag_parser"
	"os"
)

func main() {
	// Flags
	flags := os.Args[1:]
	fmt.Printf("Flags %+v\n", flags)

	flagParser := flagparser.ParseFlags(flags)

	flagParser.Parse()
	fmt.Printf("GITHUB ACCESS TOKEN ENV VAR %v\n", os.Getenv("GITHUB_ACCESS_TOKEN"))
	// fmt.Printf("Flags %+v\n", flags[1])

	// Method

	//h := &h.MultiGit{Cmd: flagParser.Method, Args: flags}

	//h.Exec()
}
