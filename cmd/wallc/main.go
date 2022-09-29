package main

import (
	"fmt"
	"os"
	"path/filepath"
	"wall"
)

func main() {
	args := os.Args
	if len(args) > 2 {
		panic(fmt.Sprintf("unparsed args: %s", args[2]))
	}
	source := args[1]
	if filepath.Ext(source) != ".wl" {
		panic("a source file with an extension .wl is expected")
	}
	bytes, err := os.ReadFile(source)
	check(err)
	parsedFile, err := wall.ParseCompilationUnit(source, bytes)
	check(err)
	checkedFile, err := wall.CheckCompilationUnit(parsedFile)
	check(err)
	cSource := wall.CodegenCompilationUnit(checkedFile)
	fmt.Println(cSource)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
