// antha/token/position_test.go: Part of the Antha language
// Copyright (C) 2014 The Antha authors. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
//
// For more information relating to the software or licensing issues please
// contact license@antha-lang.org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 1 Royal College St, London NW1 0NH UK

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package token

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func checkPos(t *testing.T, msg string, p, q Position) {
	if p.Filename != q.Filename {
		t.Errorf("%s: expected filename = %q; got %q", msg, q.Filename, p.Filename)
	}
	if p.Offset != q.Offset {
		t.Errorf("%s: expected offset = %d; got %d", msg, q.Offset, p.Offset)
	}
	if p.Line != q.Line {
		t.Errorf("%s: expected line = %d; got %d", msg, q.Line, p.Line)
	}
	if p.Column != q.Column {
		t.Errorf("%s: expected column = %d; got %d", msg, q.Column, p.Column)
	}
}

func TestNoPos(t *testing.T) {
	if NoPos.IsValid() {
		t.Errorf("NoPos should not be valid")
	}
	var fset *FileSet
	checkPos(t, "nil NoPos", fset.Position(NoPos), Position{})
	fset = NewFileSet()
	checkPos(t, "fset NoPos", fset.Position(NoPos), Position{})
}

var tests = []struct {
	filename string
	source   []byte // may be nil
	size     int
	lines    []int
}{
	{"a", []byte{}, 0, []int{}},
	{"b", []byte("01234"), 5, []int{0}},
	{"c", []byte("\n\n\n\n\n\n\n\n\n"), 9, []int{0, 1, 2, 3, 4, 5, 6, 7, 8}},
	{"d", nil, 100, []int{0, 5, 10, 20, 30, 70, 71, 72, 80, 85, 90, 99}},
	{"e", nil, 777, []int{0, 80, 100, 120, 130, 180, 267, 455, 500, 567, 620}},
	{"f", []byte("package p\n\nimport \"fmt\""), 23, []int{0, 10, 11}},
	{"g", []byte("package p\n\nimport \"fmt\"\n"), 24, []int{0, 10, 11}},
	{"h", []byte("package p\n\nimport \"fmt\"\n "), 25, []int{0, 10, 11, 24}},
}

func linecol(lines []int, offs int) (int, int) {
	prevLineOffs := 0
	for line, lineOffs := range lines {
		if offs < lineOffs {
			return line, offs - prevLineOffs + 1
		}
		prevLineOffs = lineOffs
	}
	return len(lines), offs - prevLineOffs + 1
}

func verifyPositions(t *testing.T, fset *FileSet, f *File, lines []int) {
	for offs := 0; offs < f.Size(); offs++ {
		p := f.Pos(offs)
		offs2 := f.Offset(p)
		if offs2 != offs {
			t.Errorf("%s, Offset: expected offset %d; got %d", f.Name(), offs, offs2)
		}
		line, col := linecol(lines, offs)
		msg := fmt.Sprintf("%s (offs = %d, p = %d)", f.Name(), offs, p)
		checkPos(t, msg, f.Position(f.Pos(offs)), Position{f.Name(), offs, line, col})
		checkPos(t, msg, fset.Position(p), Position{f.Name(), offs, line, col})
	}
}

func makeTestSource(size int, lines []int) []byte {
	src := make([]byte, size)
	for _, offs := range lines {
		if offs > 0 {
			src[offs-1] = '\n'
		}
	}
	return src
}

func TestPositions(t *testing.T) {
	const delta = 7 // a non-zero base offset increment
	fset := NewFileSet()
	for _, test := range tests {
		// verify consistency of test case
		if test.source != nil && len(test.source) != test.size {
			t.Errorf("%s: inconsistent test case: expected file size %d; got %d", test.filename, test.size, len(test.source))
		}

		// add file and verify name and size
		f := fset.AddFile(test.filename, fset.Base()+delta, test.size)
		if f.Name() != test.filename {
			t.Errorf("expected filename %q; got %q", test.filename, f.Name())
		}
		if f.Size() != test.size {
			t.Errorf("%s: expected file size %d; got %d", f.Name(), test.size, f.Size())
		}
		if fset.File(f.Pos(0)) != f {
			t.Errorf("%s: f.Pos(0) was not found in f", f.Name())
		}

		// add lines individually and verify all positions
		for i, offset := range test.lines {
			f.AddLine(offset)
			if f.LineCount() != i+1 {
				t.Errorf("%s, AddLine: expected line count %d; got %d", f.Name(), i+1, f.LineCount())
			}
			// adding the same offset again should be ignored
			f.AddLine(offset)
			if f.LineCount() != i+1 {
				t.Errorf("%s, AddLine: expected unchanged line count %d; got %d", f.Name(), i+1, f.LineCount())
			}
			verifyPositions(t, fset, f, test.lines[0:i+1])
		}

		// add lines with SetLines and verify all positions
		if ok := f.SetLines(test.lines); !ok {
			t.Errorf("%s: SetLines failed", f.Name())
		}
		if f.LineCount() != len(test.lines) {
			t.Errorf("%s, SetLines: expected line count %d; got %d", f.Name(), len(test.lines), f.LineCount())
		}
		verifyPositions(t, fset, f, test.lines)

		// add lines with SetLinesForContent and verify all positions
		src := test.source
		if src == nil {
			// no test source available - create one from scratch
			src = makeTestSource(test.size, test.lines)
		}
		f.SetLinesForContent(src)
		if f.LineCount() != len(test.lines) {
			t.Errorf("%s, SetLinesForContent: expected line count %d; got %d", f.Name(), len(test.lines), f.LineCount())
		}
		verifyPositions(t, fset, f, test.lines)
	}
}

func TestLineInfo(t *testing.T) {
	fset := NewFileSet()
	f := fset.AddFile("foo", fset.Base(), 500)
	lines := []int{0, 42, 77, 100, 210, 220, 277, 300, 333, 401}
	// add lines individually and provide alternative line information
	for _, offs := range lines {
		f.AddLine(offs)
		f.AddLineInfo(offs, "bar", 42)
	}
	// verify positions for all offsets
	for offs := 0; offs <= f.Size(); offs++ {
		p := f.Pos(offs)
		_, col := linecol(lines, offs)
		msg := fmt.Sprintf("%s (offs = %d, p = %d)", f.Name(), offs, p)
		checkPos(t, msg, f.Position(f.Pos(offs)), Position{"bar", offs, 42, col})
		checkPos(t, msg, fset.Position(p), Position{"bar", offs, 42, col})
	}
}

func TestFiles(t *testing.T) {
	fset := NewFileSet()
	for i, test := range tests {
		base := fset.Base()
		if i%2 == 1 {
			// Setting a negative base is equivalent to
			// fset.Base(), so test some of each.
			base = -1
		}
		fset.AddFile(test.filename, base, test.size)
		j := 0
		fset.Iterate(func(f *File) bool {
			if f.Name() != tests[j].filename {
				t.Errorf("expected filename = %s; got %s", tests[j].filename, f.Name())
			}
			j++
			return true
		})
		if j != i+1 {
			t.Errorf("expected %d files; got %d", i+1, j)
		}
	}
}

// FileSet.File should return nil if Pos is past the end of the FileSet.
func TestFileSetPastEnd(t *testing.T) {
	fset := NewFileSet()
	for _, test := range tests {
		fset.AddFile(test.filename, fset.Base(), test.size)
	}
	if f := fset.File(Pos(fset.Base())); f != nil {
		t.Errorf("expected nil, got %v", f)
	}
}

func TestFileSetCacheUnlikely(t *testing.T) {
	fset := NewFileSet()
	offsets := make(map[string]int)
	for _, test := range tests {
		offsets[test.filename] = fset.Base()
		fset.AddFile(test.filename, fset.Base(), test.size)
	}
	for file, pos := range offsets {
		f := fset.File(Pos(pos))
		if f.Name() != file {
			t.Errorf("expecting %q at position %d, got %q", file, pos, f.Name())
		}
	}
}

// issue 4345. Test concurrent use of FileSet.Pos does not trigger a
// race in the FileSet position cache.
func TestFileSetRace(t *testing.T) {
	fset := NewFileSet()
	for i := 0; i < 100; i++ {
		fset.AddFile(fmt.Sprintf("file-%d", i), fset.Base(), 1031)
	}
	max := int32(fset.Base())
	var stop sync.WaitGroup
	r := rand.New(rand.NewSource(7))
	for i := 0; i < 2; i++ {
		r := rand.New(rand.NewSource(r.Int63()))
		stop.Add(1)
		go func() {
			for i := 0; i < 1000; i++ {
				fset.Position(Pos(r.Int31n(max)))
			}
			stop.Done()
		}()
	}
	stop.Wait()
}