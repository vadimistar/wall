package wall_test

import (
	"log"
	"os"
	"testing"
	"wall"
)

func TestNamedNodeFromFile(t *testing.T) {
	/*

		a.wl:

		import B
		fun a() {}

		b.wl:

		import C
		fun b() {}

		c.wl:

		import A
		fun c() {}

	*/

	ac := []byte("import B\nfun a() {}\n")
	bc := []byte("import C\nfun b() {}\n")
	cc := []byte("import A\nfun c() {}\n")

	err := os.WriteFile("./A.wl", ac, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("./B.wl", bc, 0644)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("./C.wl", cc, 0644)
	if err != nil {
		log.Fatal(err)
	}

	ac, err = os.ReadFile("./A.wl")
	if err != nil {
		log.Fatal(err)
	}
	parsedFile, err := wall.ParseFile("./A.wl", ac)
	if err != nil {
		log.Fatal(err)
	}
	nameNode, parsedFiles, err := wall.NameNodeFromFile(parsedFile)
	if err != nil {
		log.Fatal(err)
	}
	if nameNode.Name() != "A" {
		log.Fatalf("expected name 'A', but got %s", nameNode.Name())
	}
	if nameNode.Children()[0].Name() != "a" {
		log.Fatalf("expected name 'a', but got %s", nameNode.Children()[0].Name())
	}
	B := nameNode.Children()[1]
	if B.Name() != "B" {
		log.Fatalf("expected name 'B', but got %s", B.Name())
	}
	if B.Children()[0].Name() != "b" {
		log.Fatalf("expected name 'b', but got %s", B.Children()[0].Name())
	}
	C := B.Children()[1]
	if C.Name() != "C" {
		log.Fatalf("expected name 'C', but got %s", C.Name())
	}
	if C.Children()[0].Name() != "c" {
		log.Fatalf("expected name 'c', but got %s", C.Children()[0].Name())
	}
	A := C.Children()[1]
	if A.Name() != "A" {
		log.Fatalf("expected name 'A', but got %s", A.Name())
	}
	if A != nameNode {
		log.Fatalf("expected the same instance of 'A', but got another one")
	}
	if len(parsedFiles) != 2 {
		log.Fatalf("expected 2 parsed files, but parsed %d", len(parsedFiles))
	}
}
