package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"wall"
)

func main() {
	cHeaders := flag.Bool("c", true, "include c standard library headers")
	flag.Parse()
	source := flag.Arg(0)
	if filepath.Ext(source) != ".wall" {
		panic("a source file with an extension .wall is expected")
	}
	bytes, err := os.ReadFile(source)
	check(err)
	source = filepath.Base(source)
	parsedFile, err := wall.ParseCompilationUnit(source, bytes)
	check(err)
	checkedFile, err := wall.CheckCompilationUnit(parsedFile)
	check(err)
	wall.LowerExternFunctions(checkedFile)
	cSource := wall.CodegenCompilationUnit(checkedFile)
	if *cHeaders {
		fmt.Println("#include <stdlib.h>")
		fmt.Println("#include <stdio.h>")
	}
	fmt.Println(cSource)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
