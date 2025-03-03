// antha/parser/error_test.go: Part of the Antha language
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

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements a parser test harness. The files in the testdata
// directory are parsed and the errors reported are compared against the
// error messages expected in the test files. The test files must end in
// .src rather than .go so that they are not disturbed by gofmt runs.
//
// Expected errors are indicated in the test files by putting a comment
// of the form /* ERROR "rx" */ immediately following an offending token.
// The harness will verify that an error matching the regular expression
// rx is reported at that source position.
//
// For instance, the following test file indicates that a "not declared"
// error should be reported for the undeclared variable x:
//
//	package p
//	func f() {
//		_ = x /* ERROR "not declared" */ + 1
//	}

package parser

import (
	"github.com/antha-lang/antha/antha/scanner"
	"github.com/antha-lang/antha/antha/token"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const testdata = "testdata"

var fsetErrs *token.FileSet

// getFile assumes that each filename occurs at most once
func getFile(filename string) (file *token.File) {
	fsetErrs.Iterate(func(f *token.File) bool {
		if f.Name() == filename {
			if file != nil {
				panic(filename + " used multiple times")
			}
			file = f
		}
		return true
	})
	return file
}

func getPos(filename string, offset int) token.Pos {
	if f := getFile(filename); f != nil {
		return f.Pos(offset)
	}
	return token.NoPos
}

// ERROR comments must be of the form /* ERROR "rx" */ and rx is
// a regular expression that matches the expected error message.
// The special form /* ERROR HERE "rx" */ must be used for error
// messages that appear immediately after a token, rather than at
// a token's position.
//
var errRx = regexp.MustCompile(`^/\* *ERROR *(HERE)? *"([^"]*)" *\*/$`)

// expectedErrors collects the regular expressions of ERROR comments found
// in files and returns them as a map of error positions to error messages.
//
func expectedErrors(t *testing.T, filename string, src []byte) map[token.Pos]string {
	errors := make(map[token.Pos]string)

	var s scanner.Scanner
	// file was parsed already - do not add it again to the file
	// set otherwise the position information returned here will
	// not match the position information collected by the parser
	s.Init(getFile(filename), src, nil, scanner.ScanComments)
	var prev token.Pos // position of last non-comment, non-semicolon token
	var here token.Pos // position immediately after the token at position prev

	for {
		pos, tok, lit := s.Scan()
		switch tok {
		case token.EOF:
			return errors
		case token.COMMENT:
			s := errRx.FindStringSubmatch(lit)
			if len(s) == 3 {
				pos := prev
				if s[1] == "HERE" {
					pos = here
				}
				errors[pos] = string(s[2])
			}
		default:
			prev = pos
			var l int // token length
			if tok.IsLiteral() {
				l = len(lit)
			} else {
				l = len(tok.String())
			}
			here = prev + token.Pos(l)
		}
	}
}

// compareErrors compares the map of expected error messages with the list
// of found errors and reports discrepancies.
//
func compareErrors(t *testing.T, expected map[token.Pos]string, found scanner.ErrorList) {
	for _, error := range found {
		// error.Pos is a token.Position, but we want
		// a token.Pos so we can do a map lookup
		pos := getPos(error.Pos.Filename, error.Pos.Offset)
		if msg, found := expected[pos]; found {
			// we expect a message at pos; check if it matches
			rx, err := regexp.Compile(msg)
			if err != nil {
				t.Errorf("%s: %v", error.Pos, err)
				continue
			}
			if match := rx.MatchString(error.Msg); !match {
				t.Errorf("%s: %q does not match %q", error.Pos, error.Msg, msg)
				continue
			}
			// we have a match - eliminate this error
			delete(expected, pos)
		} else {
			// To keep in mind when analyzing failed test output:
			// If the same error position occurs multiple times in errors,
			// this message will be triggered (because the first error at
			// the position removes this position from the expected errors).
			t.Errorf("%s: unexpected error: %s", error.Pos, error.Msg)
		}
	}

	// there should be no expected errors left
	if len(expected) > 0 {
		t.Errorf("%d errors not reported:", len(expected))
		for pos, msg := range expected {
			t.Errorf("%s: %s\n", fsetErrs.Position(pos), msg)
		}
	}
}

func checkErrors(t *testing.T, filename string, input interface{}) {
	src, err := readSource(filename, input)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = ParseFile(fsetErrs, filename, src, DeclarationErrors|AllErrors)
	found, ok := err.(scanner.ErrorList)
	if err != nil && !ok {
		t.Error(err)
		return
	}
	found.RemoveMultiples()

	// we are expecting the following errors
	// (collect these after parsing a file so that it is found in the file set)
	expected := expectedErrors(t, filename, src)

	// verify errors returned by the parser
	compareErrors(t, expected, found)
}

func TestErrors(t *testing.T) {
	fsetErrs = token.NewFileSet()
	list, err := ioutil.ReadDir(testdata)
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range list {
		name := fi.Name()
		if !fi.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".src") {
			checkErrors(t, filepath.Join(testdata, name), nil)
		}
	}
}
