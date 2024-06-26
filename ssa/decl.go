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

package ssa

import (
	"go/types"
	"log"
	"strconv"

	"github.com/goplus/llvm"
)

// -----------------------------------------------------------------------------

const (
	ClosureCtx  = "__llgo_ctx"
	ClosureStub = "__llgo_stub."
	NameValist  = "__llgo_va_list"
)

func VArg() *types.Var {
	return types.NewParam(0, nil, NameValist, types.Typ[types.Invalid])
}

func IsVArg(arg *types.Var) bool {
	return arg.Name() == NameValist
}

func HasVArg(t *types.Tuple, n int) bool {
	return n > 0 && IsVArg(t.At(n-1))
}

// -----------------------------------------------------------------------------

type aNamedConst struct {
}

// A NamedConst is a Member of a Package representing a package-level
// named constant.
//
// Pos() returns the position of the declaring ast.ValueSpec.Names[*]
// identifier.
//
// NB: a NamedConst is not a Value; it contains a constant Value, which
// it augments with the name and position of its 'const' declaration.
type NamedConst = *aNamedConst

// -----------------------------------------------------------------------------

type aGlobal struct {
	Expr
}

// A Global is a named Value holding the address of a package-level
// variable.
type Global = *aGlobal

// NewVar creates a new global variable.
func (p Package) NewVar(name string, typ types.Type, bg Background) Global {
	if v, ok := p.vars[name]; ok {
		return v
	}
	t := p.Prog.Type(typ, bg)
	gbl := llvm.AddGlobal(p.mod, t.ll, name)
	ret := &aGlobal{Expr{gbl, t}}
	p.vars[name] = ret
	return ret
}

// VarOf returns a global variable by name.
func (p Package) VarOf(name string) Global {
	return p.vars[name]
}

// Init initializes the global variable with the given value.
func (g Global) Init(v Expr) {
	g.impl.SetInitializer(v.impl)
}

// -----------------------------------------------------------------------------

// Function represents the parameters, results, and code of a function
// or method.
//
// If Blocks is nil, this indicates an external function for which no
// Go source code is available.  In this case, FreeVars, Locals, and
// Params are nil too.  Clients performing whole-program analysis must
// handle external functions specially.
//
// Blocks contains the function's control-flow graph (CFG).
// Blocks[0] is the function entry point; block order is not otherwise
// semantically significant, though it may affect the readability of
// the disassembly.
// To iterate over the blocks in dominance order, use DomPreorder().
//
// Recover is an optional second entry point to which control resumes
// after a recovered panic.  The Recover block may contain only a return
// statement, preceded by a load of the function's named return
// parameters, if any.
//
// A nested function (Parent()!=nil) that refers to one or more
// lexically enclosing local variables ("free variables") has FreeVars.
// Such functions cannot be called directly but require a
// value created by MakeClosure which, via its Bindings, supplies
// values for these parameters.
//
// If the function is a method (Signature.Recv() != nil) then the first
// element of Params is the receiver parameter.
//
// A Go package may declare many functions called "init".
// For each one, Object().Name() returns "init" but Name() returns
// "init#1", etc, in declaration order.
//
// Pos() returns the declaring ast.FuncLit.Type.Func or the position
// of the ast.FuncDecl.Name, if the function was explicit in the
// source.  Synthetic wrappers, for which Synthetic != "", may share
// the same position as the function they wrap.
// Syntax.Pos() always returns the position of the declaring "func" token.
//
// Type() returns the function's Signature.
//
// A generic function is a function or method that has uninstantiated type
// parameters (TypeParams() != nil). Consider a hypothetical generic
// method, (*Map[K,V]).Get. It may be instantiated with all ground
// (non-parameterized) types as (*Map[string,int]).Get or with
// parameterized types as (*Map[string,U]).Get, where U is a type parameter.
// In both instantiations, Origin() refers to the instantiated generic
// method, (*Map[K,V]).Get, TypeParams() refers to the parameters [K,V] of
// the generic method. TypeArgs() refers to [string,U] or [string,int],
// respectively, and is nil in the generic method.
type aFunction struct {
	Expr
	Pkg  Package
	Prog Program

	blks []BasicBlock

	params   []Type
	freeVars Expr
	base     int // base = 1 if hasFreeVars; base = 0 otherwise
	hasVArg  bool
}

// Function represents a function or method.
type Function = *aFunction

// NewFunc creates a new function.
func (p Package) NewFunc(name string, sig *types.Signature, bg Background) Function {
	return p.NewFuncEx(name, sig, bg, false)
}

// NewFuncEx creates a new function.
func (p Package) NewFuncEx(name string, sig *types.Signature, bg Background, hasFreeVars bool) Function {
	if v, ok := p.fns[name]; ok {
		return v
	}
	t := p.Prog.FuncDecl(sig, bg)
	if debugInstr {
		log.Println("NewFunc", name, t.raw.Type, "hasFreeVars:", hasFreeVars)
	}
	fn := llvm.AddFunction(p.mod, name, t.ll)
	ret := newFunction(fn, t, p, p.Prog, hasFreeVars)
	p.fns[name] = ret
	return ret
}

// FuncOf returns a function by name.
func (p Package) FuncOf(name string) Function {
	return p.fns[name]
}

func newFunction(fn llvm.Value, t Type, pkg Package, prog Program, hasFreeVars bool) Function {
	params, hasVArg := newParams(t, prog)
	base := 0
	if hasFreeVars {
		base = 1
	}
	return &aFunction{Expr{fn, t}, pkg, prog, nil, params, Expr{}, base, hasVArg}
}

func newParams(fn Type, prog Program) (params []Type, hasVArg bool) {
	sig := fn.raw.Type.(*types.Signature)
	in := sig.Params()
	if n := in.Len(); n > 0 {
		if hasVArg = HasVArg(in, n); hasVArg {
			n--
		}
		params = make([]Type, n)
		for i := 0; i < n; i++ {
			params[i] = prog.rawType(in.At(i).Type())
		}
	}
	return
}

// Params returns the function's ith parameter.
func (p Function) Param(i int) Expr {
	i += p.base // skip if hasFreeVars
	return Expr{p.impl.Param(i), p.params[i]}
}

func (p Function) closureCtx(b Builder) Expr {
	if p.freeVars.IsNil() {
		if p.base == 0 {
			panic("ssa: function has no free variables")
		}
		ptr := Expr{p.impl.Param(0), p.params[0]}
		p.freeVars = b.Load(ptr)
	}
	return p.freeVars
}

// FreeVar returns the function's ith free variable.
func (p Function) FreeVar(b Builder, i int) Expr {
	ctx := p.closureCtx(b)
	return b.getField(ctx, i)
}

// NewBuilder creates a new Builder for the function.
func (p Function) NewBuilder() Builder {
	prog := p.Prog
	b := prog.ctx.NewBuilder()
	// TODO(xsw): Finalize may cause panic, so comment it.
	// b.Finalize()
	return &aBuilder{b, p, prog}
}

// HasBody reports whether the function has a body.
func (p Function) HasBody() bool {
	return len(p.blks) > 0
}

// MakeBody creates nblk basic blocks for the function, and creates
// a new Builder associated to #0 block.
func (p Function) MakeBody(nblk int) Builder {
	p.MakeBlocks(nblk)
	b := p.NewBuilder()
	b.impl.SetInsertPointAtEnd(p.blks[0].impl)
	return b
}

// MakeBlocks creates nblk basic blocks for the function.
func (p Function) MakeBlocks(nblk int) []BasicBlock {
	n := len(p.blks)
	if n == 0 {
		p.blks = make([]BasicBlock, 0, nblk)
	}
	for i := 0; i < nblk; i++ {
		p.addBlock(n + i)
	}
	return p.blks[n:]
}

func (p Function) addBlock(idx int) BasicBlock {
	label := "_llgo_" + strconv.Itoa(idx)
	blk := llvm.AddBasicBlock(p.impl, label)
	ret := &aBasicBlock{blk, p, idx}
	p.blks = append(p.blks, ret)
	return ret
}

// MakeBlock creates a new basic block for the function.
func (p Function) MakeBlock() BasicBlock {
	return p.addBlock(len(p.blks))
}

// Block returns the ith basic block of the function.
func (p Function) Block(idx int) BasicBlock {
	return p.blks[idx]
}

// -----------------------------------------------------------------------------

type aPyObjRef struct {
	Expr
	Obj Global
}

// PyObjRef represents a python object reference.
type PyObjRef = *aPyObjRef

// NewPyFunc creates a new python function.
func (p Package) NewPyFunc(name string, sig *types.Signature, doInit bool) PyObjRef {
	if v, ok := p.pyobjs[name]; ok {
		return v
	}
	prog := p.Prog
	obj := p.NewVar(name, prog.PyObjectPtrPtr().RawType(), InC)
	if doInit {
		prog.NeedPyInit = true
		obj.Init(prog.Null(obj.Type))
		obj.impl.SetLinkage(llvm.LinkOnceAnyLinkage)
	}
	ty := &aType{obj.ll, rawType{types.NewPointer(sig)}, vkPyFuncRef}
	expr := Expr{obj.impl, ty}
	ret := &aPyObjRef{expr, obj}
	p.pyobjs[name] = ret
	return ret
}

// PyObjOf returns a python object by name.
func (p Package) PyObjOf(name string) PyObjRef {
	return p.pyobjs[name]
}

// PyObjs returns all used python objects in this project.
func (p Package) PyObjs() map[string]PyObjRef {
	return p.pyobjs
}

// -----------------------------------------------------------------------------
