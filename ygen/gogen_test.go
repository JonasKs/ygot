// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ygen

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/goyang/pkg/yang"
	"github.com/openconfig/ygot/ygot"
	"github.com/pmezard/go-difflib/difflib"
)

// generateUnifiedDiff takes two strings and generates a diff that can be
// shown to the user in a test error message.
func generateUnifiedDiff(want, got string) (string, error) {
	diffl := difflib.UnifiedDiff{
		A:        difflib.SplitLines(want),
		B:        difflib.SplitLines(got),
		FromFile: "got",
		ToFile:   "want",
		Context:  3,
		Eol:      "\n",
	}
	return difflib.GetUnifiedDiffString(diffl)
}

// wantGoStructOut is used to store the expected output of a writeGoStructs
// call.
type wantGoStructOut struct {
	wantErr    bool   // wantErr indicates whether errors are expected.
	structs    string // structs contains code repesenting a the mapped struct.
	keys       string // keys contains code representing structs used as list keys.
	methods    string // methods contains code corresponding to methods associated with the mapped struct.
	interfaces string // interfaces contains code corresponding to interfaces associated with the mapped struct.
}

// TestGoCodeStructGeneration tests the code generation from a known schema generates
// the correct structures, key types and methods for a YANG container.
func TestGoCodeStructGeneration(t *testing.T) {
	tests := []struct {
		name          string
		inStructToMap *yangDirectory
		// inMappableEntities is the set of other mappable entities that are
		// in the same module as the struct to map
		inMappableEntities map[string]*yangDirectory
		// inUniqueDirectoryNames is the set of names of structs that have been
		// defined during the pre-processing of the module, it is used to
		// determine the names of referenced lists and structs.
		inUniqueDirectoryNames map[string]string
		wantCompressed         wantGoStructOut
		wantUncompressed       wantGoStructOut
	}{{
		name: "simple single leaf mapping test",
		inStructToMap: &yangDirectory{
			name: "Tstruct",
			fields: map[string]*yang.Entry{
				"f1": {
					Name: "f1",
					Type: &yang.YangType{Kind: yang.Yint8},
					Parent: &yang.Entry{
						Name: "tstruct",
						Parent: &yang.Entry{
							Name: "root-module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Node: &yang.Leaf{
						Name: "f1",
						Parent: &yang.Module{
							Name: "exmod",
						},
					},
				},
				"f2": {
					Name:     "f2",
					Type:     &yang.YangType{Kind: yang.Ystring},
					ListAttr: &yang.ListAttr{},
					Parent: &yang.Entry{
						Name: "tstruct",
						Parent: &yang.Entry{
							Name: "root-module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Node: &yang.Leaf{
						Name: "f2",
						Parent: &yang.Module{
							Name: "exmod",
						},
					},
				},
			},
			path: []string{"", "root-module", "tstruct"},
		},
		wantCompressed: wantGoStructOut{
			structs: `
// Tstruct represents the /root-module/tstruct YANG schema element.
type Tstruct struct {
	F1	*int8	` + "`" + `path:"/tstruct/f1"` + "`" + `
	F2	[]string	` + "`" + `path:"/tstruct/f2"` + "`" + `
}

// IsYANGGoStruct ensures that Tstruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*Tstruct) IsYANGGoStruct() {}
`,
			methods: `
// Validate validates s against the YANG schema corresponding to its type.
func (s *Tstruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["Tstruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
		wantUncompressed: wantGoStructOut{
			structs: `
// Tstruct represents the /root-module/tstruct YANG schema element.
type Tstruct struct {
	F1	*int8	` + "`" + `path:"/tstruct/f1"` + "`" + `
	F2	[]string	` + "`" + `path:"/tstruct/f2"` + "`" + `
}

// IsYANGGoStruct ensures that Tstruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*Tstruct) IsYANGGoStruct() {}
`,
			methods: `
// Validate validates s against the YANG schema corresponding to its type.
func (s *Tstruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["Tstruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
	}, {
		name: "struct with a multi-type union",
		inStructToMap: &yangDirectory{
			name: "InputStruct",
			fields: map[string]*yang.Entry{
				"u1": {
					Name: "u1",
					Parent: &yang.Entry{
						Name: "input-struct",
						Parent: &yang.Entry{
							Name: "module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Node: &yang.Leaf{Parent: &yang.Module{Name: "exmod"}},
					Type: &yang.YangType{
						Kind: yang.Yunion,
						Type: []*yang.YangType{
							{Kind: yang.Ystring},
							{Kind: yang.Yint8},
						},
					},
				},
			},
			path: []string{"", "module", "input-struct"},
		},
		inUniqueDirectoryNames: map[string]string{"/module/input-struct": "InputStruct"},
		wantCompressed: wantGoStructOut{
			structs: `
// InputStruct represents the /module/input-struct YANG schema element.
type InputStruct struct {
	U1	InputStruct_U1_Union	` + "`" + `path:"/input-struct/u1"` + "`" + `
}

// IsYANGGoStruct ensures that InputStruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*InputStruct) IsYANGGoStruct() {}
`,
			methods: `
// Validate validates s against the YANG schema corresponding to its type.
func (s *InputStruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["InputStruct"], s); err != nil {
		return err
	}
	return nil
}
`,
			interfaces: `
// InputStruct_U1_Union is an interface that is implemented by valid types for the union
// for the leaf /module/input-struct/u1 within the YANG schema.
type InputStruct_U1_Union interface {
	Is_InputStruct_U1_Union()
}

// InputStruct_U1_Union_Int8 is used when /module/input-struct/u1
// is to be set to a int8 value.
type InputStruct_U1_Union_Int8 struct {
	Int8	int8
}

// Is_InputStruct_U1_Union ensures that InputStruct_U1_Union_Int8
// implements the InputStruct_U1_Union interface.
func (*InputStruct_U1_Union_Int8) Is_InputStruct_U1_Union() {}

// InputStruct_U1_Union_String is used when /module/input-struct/u1
// is to be set to a string value.
type InputStruct_U1_Union_String struct {
	String	string
}

// Is_InputStruct_U1_Union ensures that InputStruct_U1_Union_String
// implements the InputStruct_U1_Union interface.
func (*InputStruct_U1_Union_String) Is_InputStruct_U1_Union() {}

// To_InputStruct_U1_Union takes an input interface{} and attempts to convert it to a struct
// which implements the InputStruct_U1_Union union. Returns an error if the interface{} supplied
// cannot be converted to a type within the union.
func (t *InputStruct) To_InputStruct_U1_Union(i interface{}) (InputStruct_U1_Union, error) {
	switch v := i.(type) {
	case int8:
		return &InputStruct_U1_Union_Int8{v}, nil
	case string:
		return &InputStruct_U1_Union_String{v}, nil
	default:
		return nil, fmt.Errorf("cannot convert %%v to InputStruct_U1_Union, unknown union type, got: %%T, want any of [int8, string]", i, i)
	}
}
`,
		},
		wantUncompressed: wantGoStructOut{
			structs: `
// InputStruct represents the /module/input-struct YANG schema element.
type InputStruct struct {
	U1	Module_InputStruct_U1_Union	` + "`" + `path:"/input-struct/u1"` + "`" + `
}

// IsYANGGoStruct ensures that InputStruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*InputStruct) IsYANGGoStruct() {}
`,
			methods: `
// Validate validates s against the YANG schema corresponding to its type.
func (s *InputStruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["InputStruct"], s); err != nil {
		return err
	}
	return nil
}
`,
			interfaces: `
// Module_InputStruct_U1_Union is an interface that is implemented by valid types for the union
// for the leaf /module/input-struct/u1 within the YANG schema.
type Module_InputStruct_U1_Union interface {
	Is_Module_InputStruct_U1_Union()
}

// Module_InputStruct_U1_Union_Int8 is used when /module/input-struct/u1
// is to be set to a int8 value.
type Module_InputStruct_U1_Union_Int8 struct {
	Int8	int8
}

// Is_Module_InputStruct_U1_Union ensures that Module_InputStruct_U1_Union_Int8
// implements the Module_InputStruct_U1_Union interface.
func (*Module_InputStruct_U1_Union_Int8) Is_Module_InputStruct_U1_Union() {}

// Module_InputStruct_U1_Union_String is used when /module/input-struct/u1
// is to be set to a string value.
type Module_InputStruct_U1_Union_String struct {
	String	string
}

// Is_Module_InputStruct_U1_Union ensures that Module_InputStruct_U1_Union_String
// implements the Module_InputStruct_U1_Union interface.
func (*Module_InputStruct_U1_Union_String) Is_Module_InputStruct_U1_Union() {}

// To_Module_InputStruct_U1_Union takes an input interface{} and attempts to convert it to a struct
// which implements the Module_InputStruct_U1_Union union. Returns an error if the interface{} supplied
// cannot be converted to a type within the union.
func (t *InputStruct) To_Module_InputStruct_U1_Union(i interface{}) (Module_InputStruct_U1_Union, error) {
	switch v := i.(type) {
	case int8:
		return &Module_InputStruct_U1_Union_Int8{v}, nil
	case string:
		return &Module_InputStruct_U1_Union_String{v}, nil
	default:
		return nil, fmt.Errorf("cannot convert %%v to Module_InputStruct_U1_Union, unknown union type, got: %%T, want any of [int8, string]", i, i)
	}
}
`,
		},
	}, {
		name: "nested container in struct",
		inStructToMap: &yangDirectory{
			name: "InputStruct",
			fields: map[string]*yang.Entry{
				"c1": {
					Name: "c1",
					Dir:  map[string]*yang.Entry{},
					Parent: &yang.Entry{
						Name: "input-struct",
						Parent: &yang.Entry{
							Name: "root-module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Node: &yang.Leaf{Parent: &yang.Module{Name: "exmod"}},
				},
			},
			path: []string{"", "root-module", "input-struct"},
		},
		inUniqueDirectoryNames: map[string]string{"/root-module/input-struct/c1": "InputStruct_C1"},
		wantCompressed: wantGoStructOut{
			structs: `
// InputStruct represents the /root-module/input-struct YANG schema element.
type InputStruct struct {
	C1	*InputStruct_C1	` + "`" + `path:"/input-struct/c1"` + "`" + `
}

// IsYANGGoStruct ensures that InputStruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*InputStruct) IsYANGGoStruct() {}
`,
			methods: `
// Validate validates s against the YANG schema corresponding to its type.
func (s *InputStruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["InputStruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
		wantUncompressed: wantGoStructOut{
			structs: `
// InputStruct represents the /root-module/input-struct YANG schema element.
type InputStruct struct {
	C1	*InputStruct_C1	` + "`" + `path:"/input-struct/c1"` + "`" + `
}

// IsYANGGoStruct ensures that InputStruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*InputStruct) IsYANGGoStruct() {}
`,
			methods: `
// Validate validates s against the YANG schema corresponding to its type.
func (s *InputStruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["InputStruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
	}, {
		name: "struct with missing struct referenced",
		inStructToMap: &yangDirectory{
			name: "AStruct",
			fields: map[string]*yang.Entry{
				"elem": {
					Name: "elem",
					Dir:  map[string]*yang.Entry{},
					Parent: &yang.Entry{
						Name: "a-struct",
						Parent: &yang.Entry{
							Name: "root-module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Node: &yang.Leaf{Parent: &yang.Module{Name: "exmod"}},
				},
			},
			path: []string{"", "root-module", "a-struct"},
		},
		wantCompressed:   wantGoStructOut{wantErr: true},
		wantUncompressed: wantGoStructOut{wantErr: true},
	}, {
		name: "struct with missing list referenced",
		inStructToMap: &yangDirectory{
			name: "BStruct",
			fields: map[string]*yang.Entry{
				"list": {
					Name:     "list",
					Dir:      map[string]*yang.Entry{},
					ListAttr: &yang.ListAttr{},
					Parent: &yang.Entry{
						Name: "b-struct",
						Parent: &yang.Entry{
							Name: "root-module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Node: &yang.Leaf{Parent: &yang.Module{Name: "exmod"}},
				},
			},
			path: []string{"", "root-module", "b-struct"},
		},
		wantCompressed:   wantGoStructOut{wantErr: true},
		wantUncompressed: wantGoStructOut{wantErr: true},
	}, {
		name: "struct with keyless list",
		inStructToMap: &yangDirectory{
			name: "QStruct",
			fields: map[string]*yang.Entry{
				"a-list": {
					Name:     "a-list",
					ListAttr: &yang.ListAttr{},
					Dir:      map[string]*yang.Entry{},
					Parent: &yang.Entry{
						Name: "q-struct",
						Parent: &yang.Entry{
							Name: "root-module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Node: &yang.Leaf{Parent: &yang.Module{Name: "exmod"}},
				},
			},
			path: []string{"", "root-module", "q-struct"},
		},
		inMappableEntities: map[string]*yangDirectory{
			"/root-module/q-struct/a-list": {
				name: "QStruct_AList",
			},
		},
		inUniqueDirectoryNames: map[string]string{
			"/root-module/q-struct/a-list": "QStruct_AList",
		},
		wantCompressed: wantGoStructOut{
			structs: `
// QStruct represents the /root-module/q-struct YANG schema element.
type QStruct struct {
	AList	[]*QStruct_AList	` + "`" + `path:"/q-struct/a-list"` + "`" + `
}

// IsYANGGoStruct ensures that QStruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*QStruct) IsYANGGoStruct() {}
`,
			methods: `
// Validate validates s against the YANG schema corresponding to its type.
func (s *QStruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["QStruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
		wantUncompressed: wantGoStructOut{
			structs: `
// QStruct represents the /root-module/q-struct YANG schema element.
type QStruct struct {
	AList	[]*QStruct_AList	` + "`" + `path:"/q-struct/a-list"` + "`" + `
}

// IsYANGGoStruct ensures that QStruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*QStruct) IsYANGGoStruct() {}
`,
			methods: `
// Validate validates s against the YANG schema corresponding to its type.
func (s *QStruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["QStruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
	}, {
		name: "struct with single key list",
		inStructToMap: &yangDirectory{
			name: "Tstruct",
			fields: map[string]*yang.Entry{
				"listWithKey": {
					Name:     "listWithKey",
					ListAttr: &yang.ListAttr{},
					Key:      "keyLeaf",
					Parent: &yang.Entry{
						Name: "tstruct",
						Parent: &yang.Entry{
							Name: "root-module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Dir: map[string]*yang.Entry{
						"keyLeaf": {
							Name: "keyLeaf",
							Type: &yang.YangType{Kind: yang.Ystring},
						},
					},
					Node: &yang.Leaf{Parent: &yang.Module{Name: "exmod"}},
				},
			},
			path: []string{"", "root-module", "tstruct"},
		},
		inMappableEntities: map[string]*yangDirectory{
			"/root-module/tstruct/listWithKey": {
				name: "ListWithKey",
				listAttr: &yangListAttr{
					keys: map[string]mappedType{
						"keyLeaf": {nativeType: "string"},
					},
					keyElems: []*yang.Entry{
						{
							Name: "keyLeaf",
						},
					},
				},
				path: []string{"", "root-module", "tstruct", "listWithKey"},
			},
		},
		inUniqueDirectoryNames: map[string]string{
			"/root-module/tstruct/listWithKey": "ListWithKey",
		},
		wantCompressed: wantGoStructOut{
			structs: `
// Tstruct represents the /root-module/tstruct YANG schema element.
type Tstruct struct {
	ListWithKey	map[string]*ListWithKey	` + "`" + `path:"/tstruct/listWithKey"` + "`" + `
}

// IsYANGGoStruct ensures that Tstruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*Tstruct) IsYANGGoStruct() {}
`,
			methods: `
// NewListWithKey creates a new entry in the ListWithKey list of the
// Tstruct struct. The keys of the list are populated from the input
// arguments.
func (t *Tstruct) NewListWithKey(KeyLeaf string) (*ListWithKey, error){

	// Initialise the list within the receiver struct if it has not already been
	// created.
	if t.ListWithKey == nil {
		t.ListWithKey = make(map[string]*ListWithKey)
	}

	key := KeyLeaf

	// Ensure that this key has not already been used in the
	// list. Keyed YANG lists do not allow duplicate keys to
	// be created.
	if _, ok := t.ListWithKey[key]; ok {
		return nil, fmt.Errorf("duplicate key %%v for list ListWithKey", key)
	}

	t.ListWithKey[key] = &ListWithKey{
		KeyLeaf: &KeyLeaf,
	}

	return t.ListWithKey[key], nil
}

// Validate validates s against the YANG schema corresponding to its type.
func (s *Tstruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["Tstruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
		wantUncompressed: wantGoStructOut{
			structs: `
// Tstruct represents the /root-module/tstruct YANG schema element.
type Tstruct struct {
	ListWithKey	map[string]*ListWithKey	` + "`" + `path:"/tstruct/listWithKey"` + "`" + `
}

// IsYANGGoStruct ensures that Tstruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*Tstruct) IsYANGGoStruct() {}
`,
			methods: `
// NewListWithKey creates a new entry in the ListWithKey list of the
// Tstruct struct. The keys of the list are populated from the input
// arguments.
func (t *Tstruct) NewListWithKey(KeyLeaf string) (*ListWithKey, error){

	// Initialise the list within the receiver struct if it has not already been
	// created.
	if t.ListWithKey == nil {
		t.ListWithKey = make(map[string]*ListWithKey)
	}

	key := KeyLeaf

	// Ensure that this key has not already been used in the
	// list. Keyed YANG lists do not allow duplicate keys to
	// be created.
	if _, ok := t.ListWithKey[key]; ok {
		return nil, fmt.Errorf("duplicate key %%v for list ListWithKey", key)
	}

	t.ListWithKey[key] = &ListWithKey{
		KeyLeaf: &KeyLeaf,
	}

	return t.ListWithKey[key], nil
}

// Validate validates s against the YANG schema corresponding to its type.
func (s *Tstruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["Tstruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
	}, {
		name: "struct with multi-key list",
		inStructToMap: &yangDirectory{
			name: "Tstruct",
			fields: map[string]*yang.Entry{
				"listWithKey": {
					Name:     "listWithKey",
					ListAttr: &yang.ListAttr{},
					Key:      "keyLeafOne keyLeafTwo",
					Parent: &yang.Entry{
						Name: "tstruct",
						Parent: &yang.Entry{
							Name: "root-module",
							Node: &yang.Module{
								Name: "exmod",
							},
						},
					},
					Dir: map[string]*yang.Entry{
						"keyLeafOne": {
							Name: "keyLeafOne",
							Node: &yang.Leaf{Parent: &yang.Module{Name: "exmodch"}},
						},
						"keyLeafTwo": {
							Name: "keyLeafTwo",
							Node: &yang.Leaf{Parent: &yang.Module{Name: "exmod"}},
						},
					},
					Node: &yang.Leaf{Parent: &yang.Module{Name: "exmod"}},
				},
			},
			path: []string{"", "root-module", "tstruct"},
		},
		inMappableEntities: map[string]*yangDirectory{
			"/root-module/tstruct/listWithKey": {
				name: "ListWithKey",
				listAttr: &yangListAttr{
					keys: map[string]mappedType{
						"keyLeafOne": {nativeType: "string"},
						"keyLeafTwo": {nativeType: "int8"},
					},
				},
				path: []string{"", "root-module", "tstruct", "listWithKey"},
			},
		},
		inUniqueDirectoryNames: map[string]string{
			"/root-module/tstruct/listWithKey": "ListWithKey",
		},
		wantCompressed: wantGoStructOut{
			structs: `
// Tstruct represents the /root-module/tstruct YANG schema element.
type Tstruct struct {
	ListWithKey	map[Tstruct_ListWithKey_Key]*ListWithKey	` + "`" + `path:"/tstruct/listWithKey"` + "`" + `
}

// IsYANGGoStruct ensures that Tstruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*Tstruct) IsYANGGoStruct() {}
`,
			keys: `
// Tstruct_ListWithKey_Key represents the key for list ListWithKey of element /root-module/tstruct.
type Tstruct_ListWithKey_Key struct {
	KeyLeafOne	string	` + "`" + `path:"keyLeafOne"` + "`" + `
	KeyLeafTwo	int8	` + "`" + `path:"keyLeafTwo"` + "`" + `
}
`,
			methods: `
// NewListWithKey creates a new entry in the ListWithKey list of the
// Tstruct struct. The keys of the list are populated from the input
// arguments.
func (t *Tstruct) NewListWithKey(KeyLeafOne string, KeyLeafTwo int8) (*ListWithKey, error){

	// Initialise the list within the receiver struct if it has not already been
	// created.
	if t.ListWithKey == nil {
		t.ListWithKey = make(map[Tstruct_ListWithKey_Key]*ListWithKey)
	}

	key := Tstruct_ListWithKey_Key{
		KeyLeafOne: KeyLeafOne,
		KeyLeafTwo: KeyLeafTwo,
	}

	// Ensure that this key has not already been used in the
	// list. Keyed YANG lists do not allow duplicate keys to
	// be created.
	if _, ok := t.ListWithKey[key]; ok {
		return nil, fmt.Errorf("duplicate key %%v for list ListWithKey", key)
	}

	t.ListWithKey[key] = &ListWithKey{
		KeyLeafOne: &KeyLeafOne,
		KeyLeafTwo: &KeyLeafTwo,
	}

	return t.ListWithKey[key], nil
}

// Validate validates s against the YANG schema corresponding to its type.
func (s *Tstruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["Tstruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
		wantUncompressed: wantGoStructOut{
			structs: `
// Tstruct represents the /root-module/tstruct YANG schema element.
type Tstruct struct {
	ListWithKey	map[Tstruct_ListWithKey_Key]*ListWithKey	` + "`" + `path:"/tstruct/listWithKey"` + "`" + `
}

// IsYANGGoStruct ensures that Tstruct implements the yang.GoStruct
// interface. This allows functions that need to handle this struct to
// identify it as being generated by ygen.
func (*Tstruct) IsYANGGoStruct() {}
`,
			keys: `
// Tstruct_ListWithKey_Key represents the key for list ListWithKey of element /root-module/tstruct.
type Tstruct_ListWithKey_Key struct {
	KeyLeafOne	string	` + "`" + `path:"keyLeafOne"` + "`" + `
	KeyLeafTwo	int8	` + "`" + `path:"keyLeafTwo"` + "`" + `
}
`,
			methods: `
// NewListWithKey creates a new entry in the ListWithKey list of the
// Tstruct struct. The keys of the list are populated from the input
// arguments.
func (t *Tstruct) NewListWithKey(KeyLeafOne string, KeyLeafTwo int8) (*ListWithKey, error){

	// Initialise the list within the receiver struct if it has not already been
	// created.
	if t.ListWithKey == nil {
		t.ListWithKey = make(map[Tstruct_ListWithKey_Key]*ListWithKey)
	}

	key := Tstruct_ListWithKey_Key{
		KeyLeafOne: KeyLeafOne,
		KeyLeafTwo: KeyLeafTwo,
	}

	// Ensure that this key has not already been used in the
	// list. Keyed YANG lists do not allow duplicate keys to
	// be created.
	if _, ok := t.ListWithKey[key]; ok {
		return nil, fmt.Errorf("duplicate key %%v for list ListWithKey", key)
	}

	t.ListWithKey[key] = &ListWithKey{
		KeyLeafOne: &KeyLeafOne,
		KeyLeafTwo: &KeyLeafTwo,
	}

	return t.ListWithKey[key], nil
}

// Validate validates s against the YANG schema corresponding to its type.
func (s *Tstruct) Validate() error {
	if err := ytypes.Validate(SchemaTree["Tstruct"], s); err != nil {
		return err
	}
	return nil
}
`,
		},
	}}

	for _, tt := range tests {
		for compressed, want := range map[bool]wantGoStructOut{true: tt.wantCompressed, false: tt.wantUncompressed} {
			s := newGenState()
			s.uniqueDirectoryNames = tt.inUniqueDirectoryNames

			// Always generate the JSON schema for this test.
			got, errs := writeGoStruct(tt.inStructToMap, tt.inMappableEntities, s, compressed, true)

			if len(errs) != 0 && !want.wantErr {
				t.Errorf("%s writeGoStruct(CompressOCPaths: %v, targetStruct: %v): received unexpected errors: %v",
					tt.name, compressed, tt.inStructToMap, errs)
				continue
			}

			if len(errs) == 0 && want.wantErr {
				t.Errorf("%s writeGoStruct(CompressOCPaths: %v, targetStruct: %v): did not receive expected errors",
					tt.name, compressed, tt.inStructToMap)
				continue
			}

			// If we wanted an error, then skip the rest of the tests as the generated code will not
			// be correct.
			if want.wantErr {
				continue
			}

			if diff := pretty.Compare(want.structs, got.structDef); diff != "" {
				if diffl, err := generateUnifiedDiff(want.structs, got.structDef); err == nil {
					diff = diffl
				}
				t.Errorf("%s writeGoStruct(CompressOCPaths: %v, targetStruct: %v): struct generated code was not correct, diff (-got,+want):\n%s",
					tt.name, compressed, tt.inStructToMap, diff)
			}

			if diff := pretty.Compare(want.keys, got.listKeys); diff != "" {
				if diffl, err := generateUnifiedDiff(want.keys, got.listKeys); err == nil {
					diff = diffl
				}
				t.Errorf("%s writeGoStruct(CompressOCPaths: %v, targetStruct: %v): structs generated as list keys incorrect, diff (-got,+want):\n%s",
					tt.name, compressed, tt.inStructToMap, diff)
			}

			if diff := pretty.Compare(want.methods, got.methods); diff != "" {
				if diffl, err := generateUnifiedDiff(want.methods, got.methods); err == nil {
					diff = diffl
				}
				t.Errorf("%s writeGoStruct(CompressOCPaths: %v, targetStruct: %v): methods generated corresponding to lists incorrect, diff (-got,+want):\n%s",
					tt.name, compressed, tt.inStructToMap, diff)
			}

			if diff := pretty.Compare(want.interfaces, got.interfaces); diff != "" {
				if diffl, err := generateUnifiedDiff(want.interfaces, got.interfaces); err == nil {
					diff = diffl
				}
				t.Errorf("%s: writeGoStruct(CompressOCPaths: %v, targetStruct: %v): interfaces generated for struct incorrect, diff (-got,+want):\n%s",
					tt.name, compressed, tt.inStructToMap, diff)
			}
		}
	}
}

// TestGoCodeEnumGeneration validates the enumerated type code generation from a YANG
// module.
func TestGoCodeEnumGeneration(t *testing.T) {
	// In order to create a mock enum within goyang, we must construct it using the
	// relevant methods, since the field of the EnumType struct (toString) that we
	// need to set is not publicly exported.
	testEnumerations := map[string][]string{
		"enumOne": {"SPEED_2.5G", "SPEED-40G"},
		"enumTwo": {"VALUE_1", "VALUE_2", "VALUE_3", "VALUE_4"},
	}
	testYangEnums := make(map[string]*yang.EnumType)

	for name, values := range testEnumerations {
		enum := yang.NewEnumType()
		for i, enumValue := range values {
			enum.Set(enumValue, int64(i))
		}
		testYangEnums[name] = enum
	}

	tests := []struct {
		name string
		in   *yangEnum
		want goEnumCodeSnippet
	}{{
		name: "enum from identityref",
		in: &yangEnum{
			name: "EnumeratedValue",
			entry: &yang.Entry{
				Type: &yang.YangType{
					IdentityBase: &yang.Identity{
						Values: []*yang.Identity{
							{Name: "VALUE_A", Parent: &yang.Module{Name: "mod"}},
							{Name: "VALUE_C", Parent: &yang.Module{Name: "mod2"}},
							{Name: "VALUE_B", Parent: &yang.Module{Name: "mod3"}},
						},
					},
				},
			},
		},
		want: goEnumCodeSnippet{
			constDef: `
// E_EnumeratedValue is a derived int64 type which is used to represent
// the enumerated node EnumeratedValue. An additional value named
// EnumeratedValue_UNSET is added to the enumeration which is used as
// the nil value, indicating that the enumeration was not explicitly set by
// the program importing the generated structures.
type E_EnumeratedValue int64

// IsYANGGoEnum ensures that EnumeratedValue implements the yang.GoEnum
// interface. This ensures that EnumeratedValue can be identified as a
// mapped type for a YANG enumeration.
func (E_EnumeratedValue) IsYANGGoEnum() {}

// ΛMap returns the value lookup map associated with  EnumeratedValue.
func (E_EnumeratedValue) ΛMap() map[string]map[int64]ygot.EnumDefinition { return ΛEnum; }

const (
	// EnumeratedValue_UNSET corresponds to the value UNSET of EnumeratedValue
	EnumeratedValue_UNSET E_EnumeratedValue = 0
	// EnumeratedValue_VALUE_A corresponds to the value VALUE_A of EnumeratedValue
	EnumeratedValue_VALUE_A E_EnumeratedValue = 1
	// EnumeratedValue_VALUE_B corresponds to the value VALUE_B of EnumeratedValue
	EnumeratedValue_VALUE_B E_EnumeratedValue = 2
	// EnumeratedValue_VALUE_C corresponds to the value VALUE_C of EnumeratedValue
	EnumeratedValue_VALUE_C E_EnumeratedValue = 3
)
`,
			name: "EnumeratedValue",
			valToString: map[int64]ygot.EnumDefinition{
				1: {Name: "VALUE_A", DefiningModule: "mod"},
				2: {Name: "VALUE_B", DefiningModule: "mod3"},
				3: {Name: "VALUE_C", DefiningModule: "mod2"},
			},
		},
	}, {
		name: "enum from enumeration",
		in: &yangEnum{
			name: "EnumeratedValueTwo",
			entry: &yang.Entry{
				Type: &yang.YangType{Enum: testYangEnums["enumOne"]},
			},
		},
		want: goEnumCodeSnippet{
			constDef: `
// E_EnumeratedValueTwo is a derived int64 type which is used to represent
// the enumerated node EnumeratedValueTwo. An additional value named
// EnumeratedValueTwo_UNSET is added to the enumeration which is used as
// the nil value, indicating that the enumeration was not explicitly set by
// the program importing the generated structures.
type E_EnumeratedValueTwo int64

// IsYANGGoEnum ensures that EnumeratedValueTwo implements the yang.GoEnum
// interface. This ensures that EnumeratedValueTwo can be identified as a
// mapped type for a YANG enumeration.
func (E_EnumeratedValueTwo) IsYANGGoEnum() {}

// ΛMap returns the value lookup map associated with  EnumeratedValueTwo.
func (E_EnumeratedValueTwo) ΛMap() map[string]map[int64]ygot.EnumDefinition { return ΛEnum; }

const (
	// EnumeratedValueTwo_UNSET corresponds to the value UNSET of EnumeratedValueTwo
	EnumeratedValueTwo_UNSET E_EnumeratedValueTwo = 0
	// EnumeratedValueTwo_SPEED_2_5G corresponds to the value SPEED_2_5G of EnumeratedValueTwo
	EnumeratedValueTwo_SPEED_2_5G E_EnumeratedValueTwo = 1
	// EnumeratedValueTwo_SPEED_40G corresponds to the value SPEED_40G of EnumeratedValueTwo
	EnumeratedValueTwo_SPEED_40G E_EnumeratedValueTwo = 2
)
`,
			name: "EnumeratedValueTwo",
			valToString: map[int64]ygot.EnumDefinition{
				1: {Name: "SPEED_2.5G"},
				2: {Name: "SPEED-40G"},
			},
		},
	}, {
		name: "enum from longer enumeration",
		in: &yangEnum{
			name: "BaseModule_Enumeration",
			entry: &yang.Entry{
				Type: &yang.YangType{Enum: testYangEnums["enumTwo"]},
			},
		},
		want: goEnumCodeSnippet{
			constDef: `
// E_BaseModule_Enumeration is a derived int64 type which is used to represent
// the enumerated node BaseModule_Enumeration. An additional value named
// BaseModule_Enumeration_UNSET is added to the enumeration which is used as
// the nil value, indicating that the enumeration was not explicitly set by
// the program importing the generated structures.
type E_BaseModule_Enumeration int64

// IsYANGGoEnum ensures that BaseModule_Enumeration implements the yang.GoEnum
// interface. This ensures that BaseModule_Enumeration can be identified as a
// mapped type for a YANG enumeration.
func (E_BaseModule_Enumeration) IsYANGGoEnum() {}

// ΛMap returns the value lookup map associated with  BaseModule_Enumeration.
func (E_BaseModule_Enumeration) ΛMap() map[string]map[int64]ygot.EnumDefinition { return ΛEnum; }

const (
	// BaseModule_Enumeration_UNSET corresponds to the value UNSET of BaseModule_Enumeration
	BaseModule_Enumeration_UNSET E_BaseModule_Enumeration = 0
	// BaseModule_Enumeration_VALUE_1 corresponds to the value VALUE_1 of BaseModule_Enumeration
	BaseModule_Enumeration_VALUE_1 E_BaseModule_Enumeration = 1
	// BaseModule_Enumeration_VALUE_2 corresponds to the value VALUE_2 of BaseModule_Enumeration
	BaseModule_Enumeration_VALUE_2 E_BaseModule_Enumeration = 2
	// BaseModule_Enumeration_VALUE_3 corresponds to the value VALUE_3 of BaseModule_Enumeration
	BaseModule_Enumeration_VALUE_3 E_BaseModule_Enumeration = 3
	// BaseModule_Enumeration_VALUE_4 corresponds to the value VALUE_4 of BaseModule_Enumeration
	BaseModule_Enumeration_VALUE_4 E_BaseModule_Enumeration = 4
)
`,
			name: "BaseModule_Enumeration",
			valToString: map[int64]ygot.EnumDefinition{
				1: {Name: "VALUE_1"},
				2: {Name: "VALUE_2"},
				3: {Name: "VALUE_3"},
				4: {Name: "VALUE_4"},
			},
		},
	}}

	for _, tt := range tests {
		got, err := writeGoEnum(tt.in)
		if err != nil {
			t.Errorf("%s: writeGoEnum(%v): got unexpected error: %v",
				tt.name, tt.in, err)
			continue
		}

		if diff := pretty.Compare(tt.want, got); diff != "" {
			fmt.Println(diff)
			if diffl, err := generateUnifiedDiff(tt.want.constDef, got.constDef); err == nil {
				diff = diffl
			}
			t.Errorf("%s: writeGoEnum(%v): got incorrect output, diff(-got,+want):\n%s",
				tt.name, tt.in, diff)
		}
	}
}

// TestFindMapPaths ensures that the schema paths that an entity should be
// mapped to are properly extracted from a schema element.
func TestFindMapPaths(t *testing.T) {
	tests := []struct {
		name              string
		inStruct          *yangDirectory
		inField           *yang.Entry
		inCompressOCPaths bool
		wantPaths         [][]string
		wantErr           bool
	}{{
		name: "first-level container with path compression off",
		inStruct: &yangDirectory{
			name: "AContainer",
			path: []string{"", "a-module", "a-container"},
		},
		inField: &yang.Entry{
			Name: "field-a",
			Parent: &yang.Entry{
				Name: "a-container",
				Parent: &yang.Entry{
					Name: "a-module",
				},
			},
		},
		wantPaths: [][]string{{"", "a-container", "field-a"}},
	}, {
		name: "invalid parent path",
		inStruct: &yangDirectory{
			name: "AContainer",
			path: []string{"", "a-module", "a-container"},
		},
		inField: &yang.Entry{
			Name: "field-q",
			Parent: &yang.Entry{
				Name: "q-container",
			},
		},
		wantErr: true,
	}, {
		name: "first-level container with path compression on",
		inStruct: &yangDirectory{
			name: "BContainer",
			path: []string{"", "a-module", "b-container"},
		},
		inField: &yang.Entry{
			Name: "field-b",
			Parent: &yang.Entry{
				Name: "config",
				Parent: &yang.Entry{
					Name: "b-container",
					Parent: &yang.Entry{
						Name: "a-module",
					},
				},
			},
		},
		inCompressOCPaths: true,
		wantPaths:         [][]string{{"", "b-container", "config", "field-b"}},
	}, {
		name: "top-level module - not valid to map",
		inStruct: &yangDirectory{
			name: "CContainer",
			path: []string{"", "c-container"}, // Does not have a valid module.
		},
		inField: &yang.Entry{},
		wantErr: true,
	}, {
		name: "list with leafref key",
		inStruct: &yangDirectory{
			name: "DList",
			path: []string{"", "d-module", "d-container", "d-list"},
			listAttr: &yangListAttr{
				keyElems: []*yang.Entry{
					{
						Name: "d-key",
						Type: &yang.YangType{
							Kind: yang.Yleafref,
						},
						Parent: &yang.Entry{
							Name: "config",
							Parent: &yang.Entry{
								Name: "d-list",
								Dir: map[string]*yang.Entry{
									"d-key": {
										Name: "d-key",
										Type: &yang.YangType{Kind: yang.Yleafref},
									},
								},
								Parent: &yang.Entry{
									Name: "d-container",
									Parent: &yang.Entry{
										Name: "d-module",
									},
								},
							},
						},
					},
				},
			},
		},
		inField: &yang.Entry{
			Name: "d-key",
			Type: &yang.YangType{
				Kind: yang.Yleafref,
			},
			Parent: &yang.Entry{
				Name: "config",
				Parent: &yang.Entry{
					Name: "d-list",
					Dir: map[string]*yang.Entry{
						"d-key": {
							Name: "d-key",
							Type: &yang.YangType{Kind: yang.Yleafref},
						},
					},
					Parent: &yang.Entry{
						Name: "d-container",
						Parent: &yang.Entry{
							Name: "d-module",
						},
					},
				},
			},
		},
		inCompressOCPaths: true,
		wantPaths: [][]string{
			{"config", "d-key"},
			{"d-key"},
		},
	}}

	for _, tt := range tests {
		got, err := findMapPaths(tt.inStruct, tt.inField, tt.inCompressOCPaths)
		if err != nil {
			if !tt.wantErr {
				t.Errorf("%s: YANGCodeGenerator.findMapPaths(%v, %v): compress: %v, got unexpected error: %v",
					tt.name, tt.inStruct, tt.inField, tt.inCompressOCPaths, err)
			}
			continue
		}

		if !reflect.DeepEqual(got, tt.wantPaths) {
			t.Errorf("%s: YANGCodeGenerator.findMapPaths(%v, %v): compress: %v, got wrong paths, got: %v, want: %v",
				tt.name, tt.inStruct, tt.inField, tt.inCompressOCPaths, got, tt.wantPaths)
		}
	}
}

func TestGenerateEnumMap(t *testing.T) {
	tests := []struct {
		name    string
		inMap   map[string]map[int64]ygot.EnumDefinition
		wantErr bool
		wantMap string
	}{{
		name: "simple map input",
		inMap: map[string]map[int64]ygot.EnumDefinition{
			"EnumOne": {
				1: {Name: "VAL1"},
				2: {Name: "VAL2"},
			},
		},
		wantMap: `
// ΛEnum is a map, keyed by the name of the type defined for each enum in the
// generated Go code, which provides a mapping between the constant int64 value
// of each value of the enumeration, and the string that is used to represent it
// in the YANG schema. The map is named ΛEnum in order to avoid clash with any
// valid YANG identifier.
var ΛEnum = map[string]map[int64]ygot.EnumDefinition{
	"E_EnumOne": {
		1: {Name: "VAL1"},
		2: {Name: "VAL2"},
	},
}
`,
	}, {
		name: "multiple enum input",
		inMap: map[string]map[int64]ygot.EnumDefinition{
			"EnumOne": {
				1: {Name: "VAL1"},
				2: {Name: "VAL2"},
			},
			"EnumTwo": {
				1: {Name: "VAL42"},
				2: {Name: "VAL43"},
			},
		},
		wantMap: `
// ΛEnum is a map, keyed by the name of the type defined for each enum in the
// generated Go code, which provides a mapping between the constant int64 value
// of each value of the enumeration, and the string that is used to represent it
// in the YANG schema. The map is named ΛEnum in order to avoid clash with any
// valid YANG identifier.
var ΛEnum = map[string]map[int64]ygot.EnumDefinition{
	"E_EnumOne": {
		1: {Name: "VAL1"},
		2: {Name: "VAL2"},
	},
	"E_EnumTwo": {
		1: {Name: "VAL42"},
		2: {Name: "VAL43"},
	},
}
`,
	}}

	for _, tt := range tests {
		got, err := generateEnumMap(tt.inMap)

		if err != nil {
			if !tt.wantErr {
				t.Errorf("%s: got unexpected error when generating map: %v", tt.name, err)
			}
			continue
		}

		if tt.wantMap != got {
			diff := fmt.Sprintf("got: %s, want %s", got, tt.wantMap)
			if diffl, err := generateUnifiedDiff(tt.wantMap, got); err == nil {
				diff = "diff (-got, +want):\n" + diffl
			}
			t.Errorf("%s: did not get expected generated enum, %s", tt.name, diff)
		}
	}
}
