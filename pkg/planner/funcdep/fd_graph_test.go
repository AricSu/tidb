// Copyright 2022 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package funcdep

import (
	"testing"

	"github.com/pingcap/tidb/pkg/util/intset"
	"github.com/stretchr/testify/require"
)

func TestAddStrictFunctionalDependency(t *testing.T) {
	fd := FDSet{
		fdEdges: []*fdEdge{},
	}
	fe1 := &fdEdge{
		from: intset.NewFastIntSet(1, 2), // AB -> CDEFG
		to:   intset.NewFastIntSet(3, 4, 5, 6, 7),
	}
	fe2 := &fdEdge{
		from: intset.NewFastIntSet(1, 2), // AB -> CD
		to:   intset.NewFastIntSet(3, 4),
	}
	fe3 := &fdEdge{
		from: intset.NewFastIntSet(1, 2), // AB -> EF
		to:   intset.NewFastIntSet(5, 6),
	}
	// fd: AB -> CDEFG implies all of others.
	assertF := func() {
		require.Equal(t, len(fd.fdEdges), 1)
		from := fd.fdEdges[0].from.SortedArray()
		require.Equal(t, len(from), 2)
		require.Equal(t, from[0], 1)
		require.Equal(t, from[1], 2)
		to := fd.fdEdges[0].to.SortedArray()
		require.Equal(t, len(to), 5)
		require.Equal(t, to[0], 3)
		require.Equal(t, to[1], 4)
		require.Equal(t, to[2], 5)
		require.Equal(t, to[3], 6)
		require.Equal(t, to[4], 7)
	}
	fd.AddStrictFunctionalDependency(fe1.from, fe1.to)
	fd.AddStrictFunctionalDependency(fe2.from, fe2.to)
	fd.AddStrictFunctionalDependency(fe3.from, fe3.to)
	assertF()

	fd.fdEdges = fd.fdEdges[:0]
	fd.AddStrictFunctionalDependency(fe2.from, fe2.to)
	fd.AddStrictFunctionalDependency(fe1.from, fe1.to)
	fd.AddStrictFunctionalDependency(fe3.from, fe3.to)
	assertF()
	// TODO:
	// test reduce col
	// test more edges
}

// Preface Notice:
// For test convenience, we add fdEdge to fdSet directly which is not valid in the procedure.
// Because two difference fdEdge in the fdSet may imply each other which is strictly not permitted in the procedure.
// Use `AddStrictFunctionalDependency` to add the fdEdge to the fdSet in the formal way .
func TestFDSet_ClosureOf(t *testing.T) {
	fd := FDSet{
		fdEdges: []*fdEdge{},
	}
	fe1 := &fdEdge{
		from:   intset.NewFastIntSet(1, 2), // AB -> CD
		to:     intset.NewFastIntSet(3, 4),
		strict: true,
		equiv:  false,
	}
	fe2 := &fdEdge{
		from:   intset.NewFastIntSet(1, 2), // AB -> EF
		to:     intset.NewFastIntSet(5, 6),
		strict: true,
		equiv:  false,
	}
	fe3 := &fdEdge{
		from:   intset.NewFastIntSet(2), // B -> FG
		to:     intset.NewFastIntSet(6, 7),
		strict: true,
		equiv:  false,
	}
	fe4 := &fdEdge{
		from:   intset.NewFastIntSet(1), // A -> DEH
		to:     intset.NewFastIntSet(4, 5, 8),
		strict: true,
		equiv:  false,
	}
	fd.fdEdges = append(fd.fdEdges, fe1, fe2, fe3, fe4)
	// A -> ADEH
	closure := fd.closureOfStrict(intset.NewFastIntSet(1)).SortedArray()
	require.Equal(t, len(closure), 4)
	require.Equal(t, closure[0], 1)
	require.Equal(t, closure[1], 4)
	require.Equal(t, closure[2], 5)
	require.Equal(t, closure[3], 8)
	// AB -> ABCDEFGH
	fd.fdEdges = append(fd.fdEdges, fe1, fe2, fe3, fe4)
	closure = fd.closureOfStrict(intset.NewFastIntSet(1, 2)).SortedArray()
	require.Equal(t, len(closure), 8)
	require.Equal(t, closure[0], 1)
	require.Equal(t, closure[1], 2)
	require.Equal(t, closure[2], 3)
	require.Equal(t, closure[3], 4)
	require.Equal(t, closure[4], 5)
	require.Equal(t, closure[5], 6)
	require.Equal(t, closure[6], 7)
	require.Equal(t, closure[7], 8)
}

func TestFDSet_ReduceCols(t *testing.T) {
	fd := FDSet{
		fdEdges: []*fdEdge{},
	}
	fe1 := &fdEdge{
		from:   intset.NewFastIntSet(1), // A -> CD
		to:     intset.NewFastIntSet(3, 4),
		strict: true,
		equiv:  false,
	}
	fe2 := &fdEdge{
		from:   intset.NewFastIntSet(3), // C -> DE
		to:     intset.NewFastIntSet(4, 5),
		strict: true,
		equiv:  false,
	}
	fe3 := &fdEdge{
		from:   intset.NewFastIntSet(3, 5), // CE -> B
		to:     intset.NewFastIntSet(2),
		strict: true,
		equiv:  false,
	}
	fd.fdEdges = append(fd.fdEdges, fe1, fe2, fe3)
	res := fd.ReduceCols(intset.NewFastIntSet(1, 2)).SortedArray()
	require.Equal(t, len(res), 1)
	require.Equal(t, res[0], 1)
}

func TestFDSet_InClosure(t *testing.T) {
	fd := FDSet{
		fdEdges: []*fdEdge{},
	}
	fe1 := &fdEdge{
		from:   intset.NewFastIntSet(1, 2), // AB -> CD
		to:     intset.NewFastIntSet(3, 4),
		strict: true,
		equiv:  false,
	}
	fe2 := &fdEdge{
		from:   intset.NewFastIntSet(1, 2), // AB -> EF
		to:     intset.NewFastIntSet(5, 6),
		strict: true,
		equiv:  false,
	}
	fe3 := &fdEdge{
		from:   intset.NewFastIntSet(2), // B -> FG
		to:     intset.NewFastIntSet(6, 7),
		strict: true,
		equiv:  false,
	}
	fd.fdEdges = append(fd.fdEdges, fe1, fe2, fe3)
	// A -> F : false (determinants should not be torn apart)
	require.False(t, fd.InClosure(intset.NewFastIntSet(1), intset.NewFastIntSet(6)))
	// B -> G : true (dependency can be torn apart)
	require.True(t, fd.InClosure(intset.NewFastIntSet(2), intset.NewFastIntSet(7)))
	// AB -> E : true (dependency can be torn apart)
	require.True(t, fd.InClosure(intset.NewFastIntSet(1, 2), intset.NewFastIntSet(5)))
	// AB -> FG: true (in closure node set)
	require.True(t, fd.InClosure(intset.NewFastIntSet(1, 2), intset.NewFastIntSet(6, 7)))
	// AB -> DF: true (in closure node set)
	require.True(t, fd.InClosure(intset.NewFastIntSet(1, 2), intset.NewFastIntSet(4, 6)))
	// AB -> EG: true (in closure node set)
	require.True(t, fd.InClosure(intset.NewFastIntSet(1, 2), intset.NewFastIntSet(5, 7)))
	// AB -> EGH: false (H is not in closure node set)
	require.False(t, fd.InClosure(intset.NewFastIntSet(1, 2), intset.NewFastIntSet(5, 7, 8)))

	fe4 := &fdEdge{
		from:   intset.NewFastIntSet(2), // B -> CH
		to:     intset.NewFastIntSet(3, 8),
		strict: true,
		equiv:  false,
	}
	fd.fdEdges = append(fd.fdEdges, fe4)
	// AB -> EGH: true (in closure node set)
	require.True(t, fd.InClosure(intset.NewFastIntSet(1, 2), intset.NewFastIntSet(5, 7, 8)))
}

func TestFDSet_AddConstant(t *testing.T) {
	fd := FDSet{}
	require.Equal(t, "()", fd.ConstantCols().String())

	fd.AddConstants(intset.NewFastIntSet(1, 2)) // {} --> {a,b}
	require.Equal(t, len(fd.fdEdges), 1)
	require.True(t, fd.fdEdges[0].strict)
	require.False(t, fd.fdEdges[0].equiv)
	require.Equal(t, "()", fd.fdEdges[0].from.String())
	require.Equal(t, "(1,2)", fd.fdEdges[0].to.String())
	require.Equal(t, "(1,2)", fd.ConstantCols().String())

	fd.AddConstants(intset.NewFastIntSet(3)) // c, {} --> {a,b,c}
	require.Equal(t, len(fd.fdEdges), 1)
	require.True(t, fd.fdEdges[0].strict)
	require.False(t, fd.fdEdges[0].equiv)
	require.Equal(t, "()", fd.fdEdges[0].from.String())
	require.Equal(t, "(1-3)", fd.fdEdges[0].to.String())
	require.Equal(t, "(1-3)", fd.ConstantCols().String())

	fd.AddStrictFunctionalDependency(intset.NewFastIntSet(3, 4), intset.NewFastIntSet(5, 6)) // {c,d} --> {e,f}
	require.Equal(t, len(fd.fdEdges), 2)
	require.True(t, fd.fdEdges[0].strict)
	require.False(t, fd.fdEdges[0].equiv)
	require.Equal(t, "()", fd.fdEdges[0].from.String())
	require.Equal(t, "(1-3)", fd.fdEdges[0].to.String())
	require.Equal(t, "(1-3)", fd.ConstantCols().String())
	require.True(t, fd.fdEdges[1].strict)
	require.False(t, fd.fdEdges[1].equiv)
	require.Equal(t, "(4)", fd.fdEdges[1].from.String()) // determinant 3 reduced as constant, leaving FD {d} --> {f,g}.
	require.Equal(t, "(5,6)", fd.fdEdges[1].to.String())

	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(7), intset.NewFastIntSet(5, 6)) // {g} ~~> {e,f}
	require.Equal(t, len(fd.fdEdges), 3)
	require.False(t, fd.fdEdges[2].strict)
	require.False(t, fd.fdEdges[2].equiv)
	require.Equal(t, "(7)", fd.fdEdges[2].from.String())
	require.Equal(t, "(5,6)", fd.fdEdges[2].to.String())

	// add d, {} --> {a,b,c,d}, and FD {d} --> {f,g} is transferred to constant closure.
	fd.AddConstants(intset.NewFastIntSet(4))
	// => {} --> {a,b,c,d,e,f}, for lax FD {g} ~~> {e,f}, dependencies are constants, removed.
	require.Equal(t, 1, len(fd.fdEdges))
	require.True(t, fd.fdEdges[0].strict)
	require.False(t, fd.fdEdges[0].equiv)
	require.Equal(t, "()", fd.fdEdges[0].from.String())
	require.Equal(t, "(1-6)", fd.fdEdges[0].to.String())
	require.Equal(t, "(1-6)", fd.ConstantCols().String())
}

func TestFDSet_LaxImplies(t *testing.T) {
	fd := FDSet{}
	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(1), intset.NewFastIntSet(2, 3))
	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(1), intset.NewFastIntSet(2))
	// lax FD won't imply each other once they have the different to side.
	require.Equal(t, "(1)~~>(2,3), (1)~~>(2)", fd.String())

	fd = FDSet{}
	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(1), intset.NewFastIntSet(2))
	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(1), intset.NewFastIntSet(2, 3))
	require.Equal(t, "(1)~~>(2), (1)~~>(2,3)", fd.String())

	fd = FDSet{}
	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(1), intset.NewFastIntSet(3))
	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(1, 2), intset.NewFastIntSet(3))
	// lax FD can imply each other once they have the same to side. {1,2} ~~> {3} implies {1} ~~> {3}
	require.Equal(t, "(1)~~>(3)", fd.String())

	fd = FDSet{}
	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(1), intset.NewFastIntSet(3, 4))
	fd.AddLaxFunctionalDependency(intset.NewFastIntSet(1, 2), intset.NewFastIntSet(3))
	// lax FD won't imply each other once they have the different to side. {1,2} ~~> {3} implies {1} ~~> {3}
	require.Equal(t, "(1)~~>(3,4), (1,2)~~>(3)", fd.String())
}

func TestFDSet_AddEquivalence(t *testing.T) {
	fd := FDSet{}
	require.Equal(t, 0, len(fd.EquivalenceCols()))

	fd.AddEquivalence(intset.NewFastIntSet(1), intset.NewFastIntSet(2)) // {a} == {b}
	require.Equal(t, 1, len(fd.fdEdges))                                // res: {a} == {b}
	require.Equal(t, 1, len(fd.EquivalenceCols()))
	require.True(t, fd.fdEdges[0].strict)
	require.True(t, fd.fdEdges[0].equiv)
	require.Equal(t, "(1,2)", fd.fdEdges[0].from.String())
	require.Equal(t, "(1,2)", fd.fdEdges[0].to.String())
	require.Equal(t, "(1,2)", fd.EquivalenceCols()[0].String())

	fd.AddEquivalence(intset.NewFastIntSet(3), intset.NewFastIntSet(4)) // {c} == {d}
	require.Equal(t, 2, len(fd.fdEdges))                                // res: {a,b} == {a,b}, {c,d} == {c,d}
	require.Equal(t, 2, len(fd.EquivalenceCols()))
	require.True(t, fd.fdEdges[0].strict)
	require.True(t, fd.fdEdges[0].equiv)
	require.Equal(t, "(1,2)", fd.fdEdges[0].from.String())
	require.Equal(t, "(1,2)", fd.fdEdges[0].to.String())
	require.Equal(t, "(1,2)", fd.EquivalenceCols()[0].String())
	require.True(t, fd.fdEdges[1].strict)
	require.True(t, fd.fdEdges[1].equiv)
	require.Equal(t, "(3,4)", fd.fdEdges[1].from.String())
	require.Equal(t, "(3,4)", fd.fdEdges[1].to.String())
	require.Equal(t, "(3,4)", fd.EquivalenceCols()[1].String())

	fd.AddConstants(intset.NewFastIntSet(4, 5)) // {} --> {d,e}
	require.Equal(t, 3, len(fd.fdEdges))        // res: {a,b} == {a,b}, {c,d} == {c,d},{} --> {c,d,e}
	require.True(t, fd.fdEdges[2].strict)       // explain: constant closure is extended by equivalence {c,d} == {c,d}
	require.False(t, fd.fdEdges[2].equiv)
	require.Equal(t, "()", fd.fdEdges[2].from.String())
	require.Equal(t, "(3-5)", fd.fdEdges[2].to.String())
	require.Equal(t, "(3-5)", fd.ConstantCols().String())

	fd.AddStrictFunctionalDependency(intset.NewFastIntSet(2, 3), intset.NewFastIntSet(5, 6)) // {b,c} --> {e,f}
	// res: {a,b} == {a,b}, {c,d} == {c,d},{} --> {c,d,e}, {b} --> {e,f}
	require.Equal(t, 4, len(fd.fdEdges))
	// explain: strict FD's from side c is eliminated by constant closure.
	require.True(t, fd.fdEdges[3].strict)
	require.False(t, fd.fdEdges[3].equiv)
	require.Equal(t, "(2)", fd.fdEdges[3].from.String())
	require.Equal(t, "(5,6)", fd.fdEdges[3].to.String())

	fd.AddEquivalence(intset.NewFastIntSet(2), intset.NewFastIntSet(3)) // {b} == {d}
	// res: {a,b,c,d} == {a,b,c,d}, {} --> {a,b,c,d,e,f}
	// explain:
	// b = d build the connection between {a,b} == {a,b}, {c,d} == {c,d}, make the superset of equivalence closure.
	// the superset equivalence closure extend the existed constant closure in turn, resulting {} --> {a,b,c,d,e}
	// the superset constant closure eliminate existed strict FD, since determinants is constant,
	// so the dependencies must be constant as well.
	// so extending the current constant closure as to {} --> {a,b,c,d,e,f}
	require.Equal(t, 2, len(fd.fdEdges))
	require.Equal(t, 1, len(fd.EquivalenceCols()))
	require.Equal(t, "(1-4)", fd.EquivalenceCols()[0].String())
	require.Equal(t, "(1-6)", fd.ConstantCols().String())
}

func TestFindCommonEquivClasses(t *testing.T) {
	fd1 := &FDSet{}
	// fd1 is with equivalence classes for {1,2} and {3,4}
	fd1.addEquivalence(intset.NewFastIntSet(1, 2))
	fd1.addEquivalence(intset.NewFastIntSet(3, 4))

	fd2 := &FDSet{}
	// fd2 is with equivalence classes for {1,3} and {2,4}
	fd2.addEquivalence(intset.NewFastIntSet(1, 3))
	fd2.addEquivalence(intset.NewFastIntSet(2, 4))

	fd3 := &FDSet{}
	// fd3 is with equivalence classes for {1} and {3,4}
	fd3.addEquivalence(intset.NewFastIntSet(1))
	fd3.addEquivalence(intset.NewFastIntSet(3, 4))

	// find common equivalence classes between fd1 and fd2.
	res := FindCommonEquivClasses([]*FDSet{fd1, fd2})
	require.Equal(t, 0, len(res))

	// find common equivalence classes between fd2 and fd3.
	res = FindCommonEquivClasses([]*FDSet{fd2, fd3})
	require.Equal(t, 0, len(res))

	// find common equivalence classes between fd1 and fd3.
	res = FindCommonEquivClasses([]*FDSet{fd1, fd3})
	require.Equal(t, 1, len(res))
	require.Equal(t, "(3,4)", res[0].String())

	// find common equivalence classes between fd1, fd2 and fd3.
	res = FindCommonEquivClasses([]*FDSet{fd1, fd2, fd3})
	require.Equal(t, 0, len(res))
}
