package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"wall"

	"github.com/jessevdk/go-flags"
)

var options struct {
	Source string `short:"s" required:"true" description:"source file"`
	Output string `short:"o" optional:"true" default:"" description:"output file"`
}

func main() {
	args, err := flags.Parse(&options)
	if len(args) > 0 {
		panic(fmt.Sprintf("unparsed args: %s", args[0]))
	}
	check(err)
	if filepath.Ext(options.Source) != ".wl" {
		panic("a source file with an extension .wl is expected")
	}
	if options.Output == "" {
		options.Output = strings.TrimSuffix(options.Source, ".wl")
		if runtime.GOOS == "windows" {
			options.Output = options.Output + ".exe"
		}
	}
	bytes, err := os.ReadFile(options.Source)
	check(err)
	parsedFile, err := wall.ParseCompilationUnit(options.Source, bytes)
	check(err)
	// _, err = wall.Check(parsedFile)
	check(err)
	cSource := wall.Codegen(parsedFile)
	cFilename := strings.TrimSuffix(options.Source, ".wl") + ".c"
	check(os.WriteFile(cFilename, []byte(cSource), 1066))
	// linking
	gcc := exec.Command("gcc", cFilename, "-o", options.Output)
	output, err := gcc.Output()
	fmt.Printf("%s", output)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
