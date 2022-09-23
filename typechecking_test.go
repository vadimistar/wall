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

func TestCheckTypesSignatures(t *testing.T) {
	file := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.StructDef{
				Name: wall.Token{Content: []byte("a")},
			},
		},
	}
	mod := wall.NewModule()
	wall.CheckTypesSignatures(file, mod)
	typ, _ := mod.GlobalScope.Type("a")
	assert.Equal(t, typ, wall.NewStructType())
}

func TestCheckFunctionsSignatures(t *testing.T) {
	file := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.FunDef{
				Fun: wall.Token{},
				Id:  wall.Token{Content: []byte("a")},
				Params: []wall.FunParam{{
					Type: &wall.IdTypeNode{Token: wall.Token{Content: []byte("A")}},
				}},
				ReturnType: &wall.IdTypeNode{Token: wall.Token{Content: []byte("B")}},
			},
		},
	}
	mod := wall.NewModule()
	typeIdA := mod.DefType("A", &wall.StructType{})
	typeIdB := mod.DefType("B", &wall.StructType{})
	assert.NoError(t, wall.CheckFunctionsSignatures(file, mod))
	if fun, ok := mod.GlobalScope.Funs["a"]; assert.Equal(t, ok, true) {
		assert.Equal(t, fun, mod.TypeId(&wall.FunctionType{
			Args: []wall.TypeId{
				typeIdA,
			},
			Returns: typeIdB,
		}))
	}
}

func TestCheckTypesContents(t *testing.T) {
	file := &wall.FileNode{
		Defs: []wall.DefNode{
			&wall.StructDef{
				Name: wall.Token{Content: []byte("Employee")},
				Fields: []wall.StructField{
					{
						Name: wall.Token{Content: []byte("name")},
						Type: &wall.IdTypeNode{
							Token: wall.Token{Content: []byte("String")},
						},
					},
					{
						Name: wall.Token{Content: []byte("age")},
						Type: &wall.IdTypeNode{
							Token: wall.Token{Content: []byte("int")},
						},
					},
				},
			},
		},
	}
	mod := wall.NewModule()
	employeeTypeId := mod.DefType("Employee", wall.NewStructType())
	stringTypeId := mod.DefType("String", wall.NewStructType())
	intTypeId := mod.DefType("int", wall.NewStructType())
	assert.NoError(t, wall.CheckTypesContents(file, mod))
	if assert.IsType(t, mod.Types[employeeTypeId], &wall.StructType{}) {
		emp := mod.Types[employeeTypeId].(*wall.StructType)
		assert.Equal(t, emp.Fields, map[string]wall.TypeId{
			"name": stringTypeId,
			"age":  intTypeId,
		})
	}
}
