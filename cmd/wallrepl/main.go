package main

import (
	"bufio"
	"fmt"
	"os"
	"wall"
)

func main() {
	repl()
}

var eval = wall.NewEvaluator()

func repl() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	text, _ := reader.ReadBytes('\n')
	text = text[:len(text)-1]
	tokens, err := wall.ScanTokens("<repl>", text)
	if err != nil {
		panic(err)
	}
	parser := wall.NewParser(tokens)
	node, err2 := parser.ParseStmtAndEof()
	if err2 != nil {
		panic(err2)
	}
	res, err3 := eval.EvaluateStmt(node)
	if err3 != nil {
		panic(err3)
	}
	fmt.Println(res)
	repl()
}
