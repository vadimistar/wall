package wall_test

import (
	"testing"
	"wall"

	"github.com/stretchr/testify/assert"
)

func TestCheckForDuplications(t *testing.T) {
	notypes := []wall.DefNode{
		&wall.FunDef{
			Id: wall.Token{Content: []byte("main")},
		},
		&wall.ImportDef{
			Name: wall.Token{Content: []byte("main")},
		},
		&wall.ParsedImportDef{
			ImportDef: wall.ImportDef{
				Name: wall.Token{Content: []byte("main")},
			},
		},
	}
	types := []wall.DefNode{
		&wall.StructDef{
			Name: wall.Token{Content: []byte("main")},
		},
	}
	for _, ntp := range notypes {
		for _, tp := range types {
			f := &wall.FileNode{
				Defs: []wall.DefNode{
					ntp, tp,
				},
			}
			assert.NoError(t, wall.CheckForDuplications(f))
		}
	}
	for _, ntp1 := range notypes {
		for _, ntp2 := range notypes {
			f := &wall.FileNode{
				Defs: []wall.DefNode{
					ntp1, ntp2,
				},
			}
			assert.Error(t, wall.CheckForDuplications(f))
		}
	}
	for _, tp1 := range types {
		for _, tp2 := range types {
			f := &wall.FileNode{
				Defs: []wall.DefNode{
					tp1, tp2,
				},
			}
			assert.Error(t, wall.CheckForDuplications(f))
		}
	}
}

func TestCheckImports(t *testing.T) {
	fileA := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.ParsedImportDef{
				ImportDef: wall.ImportDef{
					Name: wall.Token{Content: []byte("B")},
				},
				ParsedNode: nil,
			},
		},
	}
	fileB := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.ParsedImportDef{
				ImportDef: wall.ImportDef{
					Name: wall.Token{Content: []byte("A")},
				},
				ParsedNode: fileA,
			},
		},
	}
	fileA.Defs[0].(*wall.ParsedImportDef).ParsedNode = fileB
	moduleA := wall.NewModule()
	wall.CheckImports(fileA, moduleA)
	moduleB := moduleA.GlobalScope.Import("B")
	if assert.NotNil(t, moduleB) {
		moduleA2 := moduleB.GlobalScope.Import("A")
		assert.Equal(t, moduleA, moduleA2)
	}
}
