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
	"tinygo.org/x/go-llvm"
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
	_, err = wall.Check(parsedFile)
	check(err)
	llvmModule := wall.Codegen(parsedFile)
	check(err)
	bytecodeFilename := strings.TrimSuffix(options.Source, ".wl") + ".bc"
	bytecodeFile, err := os.Create(bytecodeFilename)
	check(err)
	llvmModule.Dump()
	llvm.WriteBitcodeToFile(llvmModule, bytecodeFile)
	llvmModule.Dispose()
	// compiling the bytecode to an object file
	objectFilename := strings.TrimSuffix(options.Source, ".wl") + ".o"
	llc := exec.Command("llc", "-filetype=obj", "-o", objectFilename, bytecodeFilename)
	output, err := llc.Output()
	fmt.Printf("%s", output)
	check(err)
	// linking
	ld := exec.Command("gcc", objectFilename, "-o", options.Output)
	output, err = ld.Output()
	fmt.Printf("%s", output)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
