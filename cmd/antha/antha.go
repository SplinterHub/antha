// antha/cmd/antha/antha.go: Part of the Antha language
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

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/SplinterHub/antha/antha/ast"
	"github.com/SplinterHub/antha/antha/compile"
	"github.com/SplinterHub/antha/antha/execute"
	"github.com/SplinterHub/antha/antha/parser"
	"github.com/SplinterHub/antha/antha/scanner"
	"github.com/SplinterHub/antha/antha/token"
)

// execution variables
var (
	exitCode   = 0
	fileSet    = token.NewFileSet() // per process FileSet
	parserMode parser.Mode

	componentLibrary = make([]execute.ComponentInfo, 0)
)

// parameters to control code formatting
const (
	tabWidth    = 8
	printerMode = compile.UseSpaces | compile.TabIndent
)

// command line parameters
var (
	// main operation modes
	trace                  = flag.Bool("trace", false, "show AST trace")
	allErrors              = flag.Bool("errors", false, "report all errors (not just the first 10 on different lines)")
	genComponentLib        = flag.Bool("componentLib", false, "generate component lib (instead of standalone runner)")
	genComponentLibPackage = flag.String("componentLibPackage", "componentlib", "name of package to use when componentLib is true")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: antha [flags] [path ...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

// utility function
func report(err error) {
	scanner.PrintError(os.Stderr, err)
	exitCode = 2
}

func walkDir(path string) {
	filepath.Walk(path, visitFile)
}

func initParserMode() {
	parserMode = parser.ParseComments
	if *allErrors {
		parserMode |= parser.AllErrors
	}
	if *trace {
		parserMode |= parser.Trace
	}
}

// Utility function to check file extension
func isAnthaFile(f os.FileInfo) bool {
	// ignore non-Antha or Go files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".an")
}

func main() {
	// call gofmtMain in a separate function
	// so that it can use defer and have them
	// run before the exit.
	anthaMain()
	os.Exit(exitCode)
}

func anthaMain() {
	flag.Usage = usage
	flag.Parse()

	initParserMode()

	// try to parse standard input if no files or directories were passed in
	if flag.NArg() == 0 {
		if err := processFile("<standard input>", os.Stdin, os.Stdout, true, false); err != nil {
			report(err)
		}
		return
	}

	// parse every filename or directory passed in as input
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			report(err)
		case dir.IsDir():
			walkDir(path)
		default:
			if err := processFile(path, nil, os.Stdout, false, flag.NArg() > 1); err != nil {
				report(err)
			}
		}
	}

	if len(componentLibrary) > 1 {
		ret, err := getPackageRoute(nil)
		if err != nil {
			report(err)
		}
		packagePath := *ret

		var binBuf bytes.Buffer
		var outName string
		if *genComponentLib {
			compile.GenerateComponentLib(&binBuf, componentLibrary, packagePath, *genComponentLibPackage)
			outName = *genComponentLibPackage + ".go"
		} else {
			compile.GenerateGraphRunner(&binBuf, componentLibrary, packagePath)
			outName = "main.go"
		}
		res := binBuf.Bytes()
		err = ioutil.WriteFile(outName, res, 0777)
		if err != nil {
			report(err)
		}
	}
}

// recursive support function for walkpath
func visitFile(path string, f os.FileInfo, err error) error {
	if err == nil && isAnthaFile(f) {
		err = processFile(path, nil, os.Stdout, false, false) // TODO this might be an issue since we have to analise the contents in order to establish wether more than one component exist
	}
	if err != nil {
		report(err)
	}
	return nil
}

// If in == nil, the source is the contents of the file with the given filename.
// @argument graph bool wether we want the output for a single component or a graph binary
func processFile(filename string, in io.Reader, out io.Writer, stdin bool, graph bool) error {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	file, adjust, err := parse(fileSet, filename, src, stdin)
	if err != nil {
		return err
	}

	// if this isn't an antha file, bail
	if file.Tok != token.PROTOCOL {
		return errors.New(filename + " is not a valid Antha file")
	}

	ast.SortImports(fileSet, file)

	var buf bytes.Buffer
	compiler := &compile.Config{Mode: printerMode, Tabwidth: tabWidth}
	//TODO probably here is a good fit for a one compiler.init execution
	err = compiler.Fprint(&buf, fileSet, file)
	if err != nil {
		return err
	}
	res := buf.Bytes()
	if adjust != nil {
		res = adjust(src, res)
	}
	comp := compiler.GetFileComponentInfo(fileSet, file)
	componentLibrary = append(componentLibrary, comp)

	var filename2 string
	if graph {
		st, _ := os.Stat(filename) //we already checked it existed...
		//we need to create a folder too
		dirName := strings.TrimSuffix(comp.Name, ".an") //TODO this is wrong, could have a different componentName
		fi, err := os.Stat(dirName)
		if err != nil { //does not exist
			err = os.Mkdir(dirName, 0777)
			if err != nil {
				panic(err)
				return err
			}
			fi, err = os.Stat(dirName)
			if err != nil {
				return err
			}
		}
		if !fi.IsDir() {
			return errors.New(dirName + " is not a Directory")
		}
		filename2 = dirName + string(os.PathSeparator) + strings.TrimSuffix(st.Name(), ".an") + ".go"

	} else {
		filename2 = strings.TrimSuffix(filename, ".an") + ".go"
	}
	// save the output as a translated .go file in same location as .an
	err = ioutil.WriteFile(filename2, res, 0777)
	if err != nil {
		return err
	}

	if !graph { //get the normal main
		//dirName is the new directory where we will store the main file
		dirName := fmt.Sprintf("%s_run", comp.Name)
		newdir, _ := getPackageRoute(&filename)
		//packagePath is the route used to import the go generated file
		packagePath := *newdir

		var mainBuf bytes.Buffer
		err = compiler.MainFprint(&mainBuf, fileSet, file, packagePath)
		if err != nil {
			return err
		}
		res = mainBuf.Bytes()
		fi, err := os.Stat(dirName)
		if err != nil { //does not exist
			err = os.Mkdir(dirName, 0777)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fi, err = os.Stat(dirName)
			if err != nil {
				return err
			}
		}
		if !fi.IsDir() {
			return errors.New(dirName + " is not a Directory")
		}
		err = ioutil.WriteFile(dirName+"/main.go", res, 0777)
		if err != nil {
			return err
		}
	}
	return err
}

//getPackageRoute returns for a given file the route relative to the GOPATH configured that should be included when
//importing it. If file is nil, then the route is calculated relatively to the working directory
func getPackageRoute(file *string) (*string, error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return nil, errors.New("GOPATH is not configured")
	}
	if file == nil {
		workingDirectory, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(workingDirectory, goPath) {
			return nil, errors.New("antha must be called from a valid go src folder")
		}
		ret := workingDirectory[len(goPath)+5:]
		return &ret, nil //5 is for /src/
	} else {
		//This will give me the real directory for a file
		filedirectory := filepath.Dir(*file)
		path, err := filepath.Abs(filedirectory)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(path, goPath) {
			return nil, errors.New("protocl must be located in a valid go src folder")
		}
		ret := path[len(goPath)+5:]
		return &ret, nil //5 is for /src/
	}
	return nil, nil
}

// parse parses src, which was read from filename,
// as a Antha source file or statement list.
func parse(fset *token.FileSet, filename string, src []byte, stdin bool) (*ast.File, func(orig, src []byte) []byte, error) {
	// Try as whole source file.
	file, err := parser.ParseFile(fset, filename, src, parserMode)
	if err == nil {
		return file, nil, nil
	}
	// If the error is that the source file didn't begin with a
	// package line and this is standard input, fall through to
	// try as a source fragment.  Stop and return on any other error.
	if !stdin || !strings.Contains(err.Error(), "expected 'package'") {
		return nil, nil, err
	}

	// If this is a declaration list, make it a source file
	// by inserting a package clause.
	// Insert using a ;, not a newline, so that the line numbers
	// in psrc match the ones in src.
	psrc := append([]byte("protocol p;"), src...)
	file, err = parser.ParseFile(fset, filename, psrc, parserMode)
	if err == nil {
		adjust := func(orig, src []byte) []byte {
			// Remove the package clause.
			// Anthafmt has turned the ; into a \n.
			src = src[len("protocol p\n"):]
			return matchSpace(orig, src)
		}
		return file, adjust, nil
	}
	// If the error is that the source file didn't begin with a
	// declaration, fall through to try as a statement list.
	// Stop and return on any other error.
	if !strings.Contains(err.Error(), "expected declaration") {
		return nil, nil, err
	}

	// If this is a statement list, make it a source file
	// by inserting a package clause and turning the list
	// into a function body.  This handles expressions too.
	// Insert using a ;, not a newline, so that the line numbers
	// in fsrc match the ones in src.
	fsrc := append(append([]byte("protocol p; func _() {"), src...), '}')
	file, err = parser.ParseFile(fset, filename, fsrc, parserMode)
	if err == nil {
		adjust := func(orig, src []byte) []byte {
			// Remove the wrapping.
			// Anthafmt has turned the ; into a \n\n.
			src = src[len("protocol p\n\nfunc _() {"):]
			src = src[:len(src)-len("}\n")]
			// Anthafmt has also indented the function body one level.
			// Remove that indent.
			src = bytes.Replace(src, []byte("\n\t"), []byte("\n"), -1)
			return matchSpace(orig, src)
		}
		return file, adjust, nil
	}

	// Failed, and out of options.
	return nil, nil, err
}

// Utility function for matchSpace
func cutSpace(b []byte) (before, middle, after []byte) {
	i := 0
	for i < len(b) && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n') {
		i++
	}
	j := len(b)
	for j > 0 && (b[j-1] == ' ' || b[j-1] == '\t' || b[j-1] == '\n') {
		j--
	}
	if i <= j {
		return b[:i], b[i:j], b[j:]
	}
	return nil, nil, b[j:]
}

// matchSpace reformats src to use the same space context as orig.
// 1) If orig begins with blank lines, matchSpace inserts them at the beginning of src.
// 2) matchSpace copies the indentation of the first non-blank line in orig
//    to every non-blank line in src.
// 3) matchSpace copies the trailing space from orig and uses it in place
//   of src's trailing space.
func matchSpace(orig []byte, src []byte) []byte {
	before, _, after := cutSpace(orig)
	i := bytes.LastIndex(before, []byte{'\n'})
	before, indent := before[:i+1], before[i+1:]

	_, src, _ = cutSpace(src)

	var b bytes.Buffer
	b.Write(before)
	for len(src) > 0 {
		line := src
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, src = line[:i+1], line[i+1:]
		} else {
			src = nil
		}
		if len(line) > 0 && line[0] != '\n' { // not blank
			b.Write(indent)
		}
		b.Write(line)
	}
	b.Write(after)
	return b.Bytes()
}
