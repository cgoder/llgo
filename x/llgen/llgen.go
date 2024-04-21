/*
 * Copyright (c) 2024 The GoPlus Authors (goplus.org). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package llgen

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"

	"github.com/goplus/llgo/cl"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	llssa "github.com/goplus/llgo/ssa"
)

func Init() {
	llssa.Initialize(llssa.InitAll)
	llssa.SetDebug(llssa.DbgFlagAll)
	cl.SetDebug(cl.DbgFlagAll)
}

func Do(inFile, outFile string) {
	ret := Gen(inFile, nil)
	err := os.WriteFile(outFile, []byte(ret), 0644)
	check(err)
}

func Gen(inFile string, src any) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, inFile, src, parser.ParseComments)
	check(err)

	files := []*ast.File{f}
	name := f.Name.Name
	pkg := types.NewPackage(name, name)
	ssaPkg, _, err := ssautil.BuildPackage(
		&types.Config{Importer: importer.Default()}, fset, pkg, files, ssa.SanityCheckFunctions)
	check(err)

	ssaPkg.WriteTo(os.Stderr)

	prog := llssa.NewProgram(nil)
	ret, err := cl.NewPackage(prog, ssaPkg, nil)
	check(err)

	return ret.String()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}