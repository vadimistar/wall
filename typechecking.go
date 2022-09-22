package wall

func CheckForDuplications(f *FileNode) error {
	notypes := make(map[string]struct{}, 0)
	types := make(map[string]struct{}, 0)
	for _, def := range f.Defs {
		id := string(def.id())
		switch df := def.(type) {
		case *FunDef, *ImportDef, *ParsedImportDef:
			if _, ok := notypes[id]; ok {
				return NewError(df.pos(), "name is already declared: %s", id)
			}
			notypes[id] = struct{}{}
		case *StructDef:
			if _, ok := types[id]; ok {
				return NewError(df.pos(), "type is already declared: %s", id)
			}
			types[id] = struct{}{}
		default:
			panic("unreachable")
		}
	}
	return nil
}
