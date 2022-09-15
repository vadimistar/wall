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
var depth = 0
var buff = make([]byte, 0)

func repl() {
	fmt.Print("> ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadBytes('\n')
	text = text[:len(text)-1]
	for _, ch := range text {
		if ch == '{' {
			depth++
		}
		if ch == '}' {
			depth--
		}
	}
	if depth < 0 {
		panic("extra closing parenthesis")
	}
	if depth > 0 {
		buff = append(buff, text...)
		buff = append(buff, '\n')
		repl()
	}
	tokens, err := wall.ScanTokens("<repl>", append(buff, text...))
	buff = make([]byte, 0)
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
