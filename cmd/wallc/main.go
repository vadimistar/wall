package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"wall"
)

func main() {
	cHeaders := flag.Bool("c", true, "include c standard library headers")
	emitParsedAst := flag.Bool("p", false, "emit a parsed ast")
	flag.Parse()
	source := flag.Arg(0)
	if filepath.Ext(source) != ".wall" {
		panic("a source file with an extension .wall is expected")
	}
	bytes, err := os.ReadFile(source)
	check(err)
	workpath := filepath.Dir(source)
	source = filepath.Base(source)
	parsedFile, err := wall.ParseCompilationUnit(source, string(bytes), workpath)
	check(err)
	if *emitParsedAst {
		j, err := json.Marshal(parsedFile)
		check(err)
		fmt.Println(j)
		return
	}
	checkedFile, err := wall.CheckCompilationUnit(parsedFile)
	check(err)
	wall.LowerExternFunctions(checkedFile)
	cSource := wall.CodegenCompilationUnit(checkedFile)
	if *cHeaders {
		fmt.Println("#include <stdlib.h>")
		fmt.Println("#include <stdio.h>")
		fmt.Println("#include <stdint.h>")
		fmt.Println("#include <string.h>")
		fmt.Println("#include <stddef.h>")
	}
	fmt.Println(cSource)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
