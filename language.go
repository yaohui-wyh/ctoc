package gocloc

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/go-enry/go-enry/v2"
)

// ClocLanguage is provided for xml-cloc and json format.
type ClocLanguage struct {
	Name       string `xml:"name,attr" json:"name,omitempty"`
	FilesCount int32  `xml:"files_count,attr" json:"files"`
	Code       int32  `xml:"code,attr" json:"code"`
	Comments   int32  `xml:"comment,attr" json:"comment"`
	Blanks     int32  `xml:"blank,attr" json:"blank"`
}

// Language is a type used to definitions and store statistics for one programming language.
type Language struct {
	Name         string
	lineComments []string
	multiLines   [][]string
	Files        []string
	Code         int32
	Comments     int32
	Blanks       int32
	Total        int32
}

// Languages is an array representation of Language.
type Languages []Language

func (ls Languages) SortByName() {
	sortFunc := func(i, j int) bool {
		return ls[i].Name < ls[j].Name
	}
	sort.Slice(ls, sortFunc)
}

func (ls Languages) SortByFiles() {
	sortFunc := func(i, j int) bool {
		if len(ls[i].Files) == len(ls[j].Files) {
			return ls[i].Code > ls[j].Code
		}
		return len(ls[i].Files) > len(ls[j].Files)
	}
	sort.Slice(ls, sortFunc)
}

func (ls Languages) SortByComments() {
	sortFunc := func(i, j int) bool {
		if ls[i].Comments == ls[j].Comments {
			return ls[i].Code > ls[j].Code
		}
		return ls[i].Comments > ls[j].Comments
	}
	sort.Slice(ls, sortFunc)
}

func (ls Languages) SortByBlanks() {
	sortFunc := func(i, j int) bool {
		if ls[i].Blanks == ls[j].Blanks {
			return ls[i].Code > ls[j].Code
		}
		return ls[i].Blanks > ls[j].Blanks
	}
	sort.Slice(ls, sortFunc)
}

func (ls Languages) SortByCode() {
	sortFunc := func(i, j int) bool {
		return ls[i].Code > ls[j].Code
	}
	sort.Slice(ls, sortFunc)
}

var reShebangEnv = regexp.MustCompile(`^#! *(\S+/env) ([a-zA-Z]+)`)
var reShebangLang = regexp.MustCompile(`^#! *[.a-zA-Z/]+/([a-zA-Z]+)`)

// Exts is the definition of the language name, keyed by the extension for each language.
var Exts = map[string]string{
	"as":          "ActionScript",
	"ada":         "Ada",
	"adb":         "Ada",
	"ads":         "Ada",
	"alda":        "Alda",
	"Ant":         "Ant",
	"adoc":        "AsciiDoc",
	"asciidoc":    "AsciiDoc",
	"asm":         "Assembly",
	"S":           "Assembly",
	"s":           "Assembly",
	"dats":        "ATS",
	"sats":        "ATS",
	"hats":        "ATS",
	"ahk":         "AutoHotkey",
	"awk":         "Awk",
	"bat":         "Batch",
	"btm":         "Batch",
	"bb":          "BitBake",
	"cairo":       "Cairo",
	"carbon":      "Carbon",
	"cbl":         "COBOL",
	"cmd":         "Batch",
	"bash":        "BASH",
	"sh":          "Bourne Shell",
	"c":           "C",
	"carp":        "Carp",
	"csh":         "C Shell",
	"ec":          "C",
	"erl":         "Erlang",
	"hrl":         "Erlang",
	"pgc":         "C",
	"capnp":       "Cap'n Proto",
	"chpl":        "Chapel",
	"circom":      "Circom",
	"cs":          "C#",
	"clj":         "Clojure",
	"coffee":      "CoffeeScript",
	"cfm":         "ColdFusion",
	"cfc":         "ColdFusion CFScript",
	"cmake":       "CMake",
	"cc":          "C++",
	"cpp":         "C++",
	"cxx":         "C++",
	"pcc":         "C++",
	"c++":         "C++",
	"cr":          "Crystal",
	"css":         "CSS",
	"cu":          "CUDA",
	"d":           "D",
	"dart":        "Dart",
	"dhall":       "Dhall",
	"dtrace":      "DTrace",
	"dts":         "Device Tree",
	"dtsi":        "Device Tree",
	"e":           "Eiffel",
	"elm":         "Elm",
	"el":          "LISP",
	"exp":         "Expect",
	"ex":          "Elixir",
	"exs":         "Elixir",
	"feature":     "Gherkin",
	"factor":      "Factor",
	"fish":        "Fish",
	"fr":          "Frege",
	"fst":         "F*",
	"F#":          "F#",   // deplicated F#/GLSL
	"GLSL":        "GLSL", // both use ext '.fs'
	"vs":          "GLSL",
	"shader":      "HLSL",
	"cg":          "HLSL",
	"cginc":       "HLSL",
	"hlsl":        "HLSL",
	"lean":        "Lean",
	"hlean":       "Lean",
	"lgt":         "Logtalk",
	"lisp":        "LISP",
	"lsp":         "LISP",
	"lua":         "Lua",
	"ls":          "LiveScript",
	"sc":          "LISP",
	"f":           "FORTRAN Legacy",
	"F":           "FORTRAN Legacy",
	"f77":         "FORTRAN Legacy",
	"for":         "FORTRAN Legacy",
	"ftn":         "FORTRAN Legacy",
	"pfo":         "FORTRAN Legacy",
	"f90":         "FORTRAN Modern",
	"F90":         "FORTRAN Modern",
	"f95":         "FORTRAN Modern",
	"f03":         "FORTRAN Modern",
	"f08":         "FORTRAN Modern",
	"gleam":       "Gleam",
	"go":          "Go",
	"go2":         "Go",
	"groovy":      "Groovy",
	"gradle":      "Groovy",
	"h":           "C Header",
	"hbs":         "Handlebars",
	"hs":          "Haskell",
	"hpp":         "C++ Header",
	"hh":          "C++ Header",
	"html":        "HTML",
	"ha":          "Hare",
	"hx":          "Haxe",
	"hxx":         "C++ Header",
	"idr":         "Idris",
	"imba":        "Imba",
	"il":          "SKILL",
	"ino":         "Arduino Sketch",
	"io":          "Io",
	"ipynb":       "Jupyter Notebook",
	"jai":         "JAI",
	"java":        "Java",
	"jsp":         "JSP",
	"js":          "JavaScript",
	"jl":          "Julia",
	"janet":       "Janet",
	"json":        "JSON",
	"jsx":         "JSX",
	"kk":          "Koka",
	"kt":          "Kotlin",
	"kts":         "Kotlin",
	"lds":         "LD Script",
	"less":        "LESS",
	"ly":          "Lilypond",
	"Objective-C": "Objective-C", // deplicated Obj-C/Matlab/Mercury
	"Matlab":      "MATLAB",      // both use ext '.m'
	"Mercury":     "Mercury",     // use ext '.m'
	"md":          "Markdown",
	"markdown":    "Markdown",
	"mo":          "Motoko",
	"Motoko":      "Motoko",
	"ne":          "Nearley",
	"nix":         "Nix",
	"nsi":         "NSIS",
	"nsh":         "NSIS",
	"nu":          "Nu",
	"ML":          "OCaml",
	"ml":          "OCaml",
	"mli":         "OCaml",
	"mll":         "OCaml",
	"mly":         "OCaml",
	"mm":          "Objective-C++",
	"maven":       "Maven",
	"makefile":    "Makefile",
	"meson":       "Meson",
	"mustache":    "Mustache",
	"m4":          "M4",
	"mojo":        "Mojo",
	"🔥":           "Mojo",
	"move":        "Move",
	"l":           "lex",
	"nim":         "Nim",
	"njk":         "Nunjucks",
	"odin":        "Odin",
	"ohm":         "Ohm",
	"php":         "PHP",
	"pas":         "Pascal",
	"PL":          "Perl",
	"pl":          "Perl",
	"pm":          "Perl",
	"plan9sh":     "Plan9 Shell",
	"pony":        "Pony",
	"ps1":         "PowerShell",
	"text":        "Plain Text",
	"txt":         "Plain Text",
	"polly":       "Polly",
	"proto":       "Protocol Buffers",
	"py":          "Python",
	"pxd":         "Cython",
	"pyx":         "Cython",
	"q":           "Q",
	"qml":         "QML",
	"r":           "R",
	"R":           "R",
	"raml":        "RAML",
	"Rebol":       "Rebol",
	"red":         "Red",
	"rego":        "Rego",
	"Rmd":         "RMarkdown",
	"rake":        "Ruby",
	"rb":          "Ruby",
	"resx":        "XML resource", // ref: https://docs.microsoft.com/en-us/dotnet/framework/resources/creating-resource-files-for-desktop-apps#ResxFiles
	"ring":        "Ring",
	"rkt":         "Racket",
	"rhtml":       "Ruby HTML",
	"rs":          "Rust",
	"rst":         "ReStructuredText",
	"sass":        "Sass",
	"scala":       "Scala",
	"scss":        "Sass",
	"scm":         "Scheme",
	"sed":         "sed",
	"stan":        "Stan",
	"sml":         "Standard ML",
	"sol":         "Solidity",
	"sql":         "SQL",
	"svelte":      "Svelte",
	"swift":       "Swift",
	"t":           "Terra",
	"tex":         "TeX",
	"thy":         "Isabelle",
	"tla":         "TLA",
	"sty":         "TeX",
	"tcl":         "Tcl/Tk",
	"toml":        "TOML",
	"TypeScript":  "TypeScript",
	"tsx":         "TypeScript",
	"tf":          "HCL",
	"um":          "Umka",
	"mat":         "Unity-Prefab",
	"prefab":      "Unity-Prefab",
	"Coq":         "Coq",
	"vala":        "Vala",
	"Verilog":     "Verilog",
	"csproj":      "MSBuild script",
	"vbproj":      "MSBuild script",
	"vcproj":      "MSBuild script",
	"vb":          "Visual Basic",
	"vim":         "VimL",
	"vue":         "Vue",
	"vy":          "Vyper",
	"xml":         "XML",
	"XML":         "XML",
	"xsd":         "XSD",
	"xsl":         "XSLT",
	"xslt":        "XSLT",
	"wxs":         "WiX",
	"yaml":        "YAML",
	"yml":         "YAML",
	"y":           "Yacc",
	"yul":         "Yul",
	"zep":         "Zephir",
	"zig":         "Zig",
	"zsh":         "Zsh",
}

var shebang2ext = map[string]string{
	"gosh":    "scm",
	"make":    "make",
	"perl":    "pl",
	"rc":      "plan9sh",
	"python":  "py",
	"ruby":    "rb",
	"escript": "erl",
}

func getShebang(line string) (shebangLang string, ok bool) {
	ret := reShebangEnv.FindAllStringSubmatch(line, -1)
	if ret != nil && len(ret[0]) == 3 {
		shebangLang = ret[0][2]
		if sl, ok := shebang2ext[shebangLang]; ok {
			return sl, ok
		}
		return shebangLang, true
	}

	ret = reShebangLang.FindAllStringSubmatch(line, -1)
	if ret != nil && len(ret[0]) >= 2 {
		shebangLang = ret[0][1]
		if sl, ok := shebang2ext[shebangLang]; ok {
			return sl, ok
		}
		return shebangLang, true
	}

	return "", false
}

func getFileTypeByShebang(path string) (shebangLang string, ok bool) {
	f, err := os.Open(path)
	if err != nil {
		return // ignore error
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return
	}
	line = bytes.TrimLeftFunc(line, unicode.IsSpace)

	if len(line) > 2 && line[0] == '#' && line[1] == '!' {
		return getShebang(string(line))
	}
	return
}

func getFileType(path string, opts *ClocOptions) (ext string, ok bool) {
	ext = filepath.Ext(path)
	base := filepath.Base(path)

	switch ext {
	case ".m", ".v", ".fs", ".r", ".ts":
		content, err := os.ReadFile(path)
		if err != nil {
			return "", false
		}
		lang := enry.GetLanguage(path, content)
		if opts.Debug {
			fmt.Printf("path=%v, lang=%v\n", path, lang)
		}
		return lang, true
	case ".mo":
		content, err := os.ReadFile(path)
		if err != nil {
			return "", false
		}
		lang := enry.GetLanguage(path, content)
		if opts.Debug {
			fmt.Printf("path=%v, lang=%v\n", path, lang)
		}
		if lang != "" {
			return "Motoko", true
		}
		return lang, true
	}

	switch base {
	case "meson.build", "meson_options.txt":
		return "meson", true
	case "CMakeLists.txt":
		return "cmake", true
	case "configure.ac":
		return "m4", true
	case "Makefile.am":
		return "makefile", true
	case "build.xml":
		return "Ant", true
	case "pom.xml":
		return "maven", true
	}

	switch strings.ToLower(base) {
	case "makefile":
		return "makefile", true
	case "nukefile":
		return "nu", true
	case "rebar": // skip
		return "", false
	}

	shebangLang, ok := getFileTypeByShebang(path)
	if ok {
		return shebangLang, true
	}

	if len(ext) >= 2 {
		return ext[1:], true
	}
	return ext, ok
}

// NewLanguage create language data store.
func NewLanguage(name string, lineComments []string, multiLines [][]string) *Language {
	return &Language{
		Name:         name,
		lineComments: lineComments,
		multiLines:   multiLines,
		Files:        []string{},
	}
}

func lang2exts(lang string) (exts string) {
	var es []string
	for ext, l := range Exts {
		if lang == l {
			switch lang {
			case "Objective-C", "MATLAB", "Mercury":
				ext = "m"
			case "F#":
				ext = "fs"
			case "GLSL":
				if ext == "GLSL" {
					ext = "fs"
				}
			case "TypeScript":
				ext = "ts"
			case "Motoko":
				ext = "mo"
			}
			es = append(es, ext)
		}
	}
	return strings.Join(es, ", ")
}

// DefinedLanguages is the type information for mapping language name(key) and NewLanguage.
type DefinedLanguages struct {
	Langs map[string]*Language
}

// GetFormattedString return DefinedLanguages as a human-readable string.
func (langs *DefinedLanguages) GetFormattedString() string {
	var buf bytes.Buffer
	var printLangs []string
	for _, lang := range langs.Langs {
		printLangs = append(printLangs, lang.Name)
	}
	sort.Strings(printLangs)
	for _, lang := range printLangs {
		buf.WriteString(fmt.Sprintf("%-30v (%s)\n", lang, lang2exts(lang)))
	}
	return buf.String()
}

// NewDefinedLanguages create DefinedLanguages.
func NewDefinedLanguages() *DefinedLanguages {
	return &DefinedLanguages{
		Langs: map[string]*Language{
			"ActionScript":        NewLanguage("ActionScript", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Ada":                 NewLanguage("Ada", []string{"--"}, [][]string{{"", ""}}),
			"Alda":                NewLanguage("Alda", []string{"#"}, [][]string{{"", ""}}),
			"Ant":                 NewLanguage("Ant", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"AsciiDoc":            NewLanguage("AsciiDoc", []string{}, [][]string{{"", ""}}),
			"Assembly":            NewLanguage("Assembly", []string{"//", ";", "#", "@", "|", "!"}, [][]string{{"/*", "*/"}}),
			"ATS":                 NewLanguage("ATS", []string{"//"}, [][]string{{"/*", "*/"}, {"(*", "*)"}}),
			"AutoHotkey":          NewLanguage("AutoHotkey", []string{";"}, [][]string{{"", ""}}),
			"Awk":                 NewLanguage("Awk", []string{"#"}, [][]string{{"", ""}}),
			"Arduino Sketch":      NewLanguage("Arduino Sketch", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Batch":               NewLanguage("Batch", []string{"REM", "rem"}, [][]string{{"", ""}}),
			"BASH":                NewLanguage("BASH", []string{"#"}, [][]string{{"", ""}}),
			"BitBake":             NewLanguage("BitBake", []string{"#"}, [][]string{{"", ""}}),
			"C":                   NewLanguage("C", []string{"//"}, [][]string{{"/*", "*/"}}),
			"C Header":            NewLanguage("C Header", []string{"//"}, [][]string{{"/*", "*/"}}),
			"C Shell":             NewLanguage("C Shell", []string{"#"}, [][]string{{"", ""}}),
			"Cairo":               NewLanguage("Cairo", []string{"//"}, [][]string{{"", ""}}),
			"Carbon":              NewLanguage("Carbon", []string{"//"}, [][]string{{"", ""}}),
			"Cap'n Proto":         NewLanguage("Cap'n Proto", []string{"#"}, [][]string{{"", ""}}),
			"Carp":                NewLanguage("Carp", []string{";"}, [][]string{{"", ""}}),
			"C#":                  NewLanguage("C#", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Chapel":              NewLanguage("Chapel", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Circom":              NewLanguage("Circom", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Clojure":             NewLanguage("Clojure", []string{"#", "#_"}, [][]string{{"", ""}}),
			"COBOL":               NewLanguage("COBOL", []string{"*", "/"}, [][]string{{"", ""}}),
			"CoffeeScript":        NewLanguage("CoffeeScript", []string{"#"}, [][]string{{"###", "###"}}),
			"Coq":                 NewLanguage("Coq", []string{"(*"}, [][]string{{"(*", "*)"}}),
			"ColdFusion":          NewLanguage("ColdFusion", []string{}, [][]string{{"<!---", "--->"}}),
			"ColdFusion CFScript": NewLanguage("ColdFusion CFScript", []string{"//"}, [][]string{{"/*", "*/"}}),
			"CMake":               NewLanguage("CMake", []string{"#"}, [][]string{{"", ""}}),
			"C++":                 NewLanguage("C++", []string{"//"}, [][]string{{"/*", "*/"}}),
			"C++ Header":          NewLanguage("C++ Header", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Crystal":             NewLanguage("Crystal", []string{"#"}, [][]string{{"", ""}}),
			"CSS":                 NewLanguage("CSS", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Cython":              NewLanguage("Cython", []string{"#"}, [][]string{{"\"\"\"", "\"\"\""}}),
			"CUDA":                NewLanguage("CUDA", []string{"//"}, [][]string{{"/*", "*/"}}),
			"D":                   NewLanguage("D", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Dart":                NewLanguage("Dart", []string{"//", "///"}, [][]string{{"/*", "*/"}}),
			"Dhall":               NewLanguage("Dhall", []string{"--"}, [][]string{{"{-", "-}"}}),
			"DTrace":              NewLanguage("DTrace", []string{}, [][]string{{"/*", "*/"}}),
			"Device Tree":         NewLanguage("Device Tree", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Eiffel":              NewLanguage("Eiffel", []string{"--"}, [][]string{{"", ""}}),
			"Elm":                 NewLanguage("Elm", []string{"--"}, [][]string{{"{-", "-}"}}),
			"Elixir":              NewLanguage("Elixir", []string{"#"}, [][]string{{"", ""}}),
			"Erlang":              NewLanguage("Erlang", []string{"%"}, [][]string{{"", ""}}),
			"Expect":              NewLanguage("Expect", []string{"#"}, [][]string{{"", ""}}),
			"Fish":                NewLanguage("Fish", []string{"#"}, [][]string{{"", ""}}),
			"Frege":               NewLanguage("Frege", []string{"--"}, [][]string{{"{-", "-}"}}),
			"F*":                  NewLanguage("F*", []string{"(*", "//"}, [][]string{{"(*", "*)"}}),
			"F#":                  NewLanguage("F#", []string{"(*"}, [][]string{{"(*", "*)"}}),
			"Lean":                NewLanguage("Lean", []string{"--"}, [][]string{{"/-", "-/"}}),
			"Logtalk":             NewLanguage("Logtalk", []string{"%"}, [][]string{{"", ""}}),
			"Lua":                 NewLanguage("Lua", []string{"--"}, [][]string{{"--[[", "]]"}}),
			"Lilypond":            NewLanguage("Lilypond", []string{"%"}, [][]string{{"", ""}}),
			"LISP":                NewLanguage("LISP", []string{";;"}, [][]string{{"#|", "|#"}}),
			"LiveScript":          NewLanguage("LiveScript", []string{"#"}, [][]string{{"/*", "*/"}}),
			"Factor":              NewLanguage("Factor", []string{"! "}, [][]string{{"", ""}}),
			"FORTRAN Legacy":      NewLanguage("FORTRAN Legacy", []string{"c", "C", "!", "*"}, [][]string{{"", ""}}),
			"FORTRAN Modern":      NewLanguage("FORTRAN Modern", []string{"!"}, [][]string{{"", ""}}),
			"Gherkin":             NewLanguage("Gherkin", []string{"#"}, [][]string{{"", ""}}),
			"Gleam":               NewLanguage("Gleam", []string{"//"}, [][]string{{"", ""}}),
			"GLSL":                NewLanguage("GLSL", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Go":                  NewLanguage("Go", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Groovy":              NewLanguage("Groovy", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Handlebars":          NewLanguage("Handlebars", []string{}, [][]string{{"<!--", "-->"}, {"{{!", "}}"}}),
			"Haskell":             NewLanguage("Haskell", []string{"--"}, [][]string{{"{-", "-}"}}),
			"Haxe":                NewLanguage("Haxe", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Hare":                NewLanguage("Hare", []string{"//"}, [][]string{{"", ""}}),
			"HLSL":                NewLanguage("HLSL", []string{"//"}, [][]string{{"/*", "*/"}}),
			"HTML":                NewLanguage("HTML", []string{"//", "<!--"}, [][]string{{"<!--", "-->"}}),
			"Idris":               NewLanguage("Idris", []string{"--"}, [][]string{{"{-", "-}"}}),
			"Imba":                NewLanguage("Imba", []string{"#"}, [][]string{{"###", "###"}}),
			"Io":                  NewLanguage("Io", []string{"//", "#"}, [][]string{{"/*", "*/"}}),
			"SKILL":               NewLanguage("SKILL", []string{";"}, [][]string{{"/*", "*/"}}),
			"JAI":                 NewLanguage("JAI", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Janet":               NewLanguage("Janet", []string{"#"}, [][]string{{"", ""}}),
			"Java":                NewLanguage("Java", []string{"//"}, [][]string{{"/*", "*/"}}),
			"JSP":                 NewLanguage("JSP", []string{"//"}, [][]string{{"/*", "*/"}}),
			"JavaScript":          NewLanguage("JavaScript", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Julia":               NewLanguage("Julia", []string{"#"}, [][]string{{"#:=", ":=#"}}),
			"Jupyter Notebook":    NewLanguage("Jupyter Notebook", []string{"#"}, [][]string{{"", ""}}),
			"JSON":                NewLanguage("JSON", []string{}, [][]string{{"", ""}}),
			"JSX":                 NewLanguage("JSX", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Koka":                NewLanguage("Koka", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Kotlin":              NewLanguage("Kotlin", []string{"//"}, [][]string{{"/*", "*/"}}),
			"LD Script":           NewLanguage("LD Script", []string{"//"}, [][]string{{"/*", "*/"}}),
			"LESS":                NewLanguage("LESS", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Objective-C":         NewLanguage("Objective-C", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Markdown":            NewLanguage("Markdown", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"Motoko":              NewLanguage("Motoko", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Nearley":             NewLanguage("Nearley", []string{"#"}, [][]string{{"", ""}}),
			"Nix":                 NewLanguage("Nix", []string{"#"}, [][]string{{"/*", "*/"}}),
			"NSIS":                NewLanguage("NSIS", []string{"#", ";"}, [][]string{{"/*", "*/"}}),
			"Nu":                  NewLanguage("Nu", []string{";", "#"}, [][]string{{"", ""}}),
			"OCaml":               NewLanguage("OCaml", []string{}, [][]string{{"(*", "*)"}}),
			"Objective-C++":       NewLanguage("Objective-C++", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Makefile":            NewLanguage("Makefile", []string{"#"}, [][]string{{"", ""}}),
			"MATLAB":              NewLanguage("MATLAB", []string{"%"}, [][]string{{"%{", "}%"}}),
			"Mercury":             NewLanguage("Mercury", []string{"%"}, [][]string{{"/*", "*/"}}),
			"Maven":               NewLanguage("Maven", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"Meson":               NewLanguage("Meson", []string{"#"}, [][]string{{"", ""}}),
			"Mojo":                NewLanguage("Mojo", []string{"#"}, [][]string{{"", ""}}),
			"Move":                NewLanguage("Move", []string{"//"}, [][]string{{"", ""}}),
			"Mustache":            NewLanguage("Mustache", []string{}, [][]string{{"{{!", "}}"}}),
			"M4":                  NewLanguage("M4", []string{"#"}, [][]string{{"", ""}}),
			"Nim":                 NewLanguage("Nim", []string{"#"}, [][]string{{"#[", "]#"}}),
			"Nunjucks":            NewLanguage("Nunjucks", []string{}, [][]string{{"{#", "#}"}, {"<!--", "-->"}}),
			"lex":                 NewLanguage("lex", []string{}, [][]string{{"/*", "*/"}}),
			"Odin":                NewLanguage("Odin", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Ohm":                 NewLanguage("Ohm", []string{"//"}, [][]string{{"/*", "*/"}}),
			"PHP":                 NewLanguage("PHP", []string{"#", "//"}, [][]string{{"/*", "*/"}}),
			"Pascal":              NewLanguage("Pascal", []string{"//"}, [][]string{{"{", ")"}}),
			"Perl":                NewLanguage("Perl", []string{"#"}, [][]string{{":=", ":=cut"}}),
			"Plain Text":          NewLanguage("Plain Text", []string{}, [][]string{{"", ""}}),
			"Plan9 Shell":         NewLanguage("Plan9 Shell", []string{"#"}, [][]string{{"", ""}}),
			"Pony":                NewLanguage("Pony", []string{"//"}, [][]string{{"/*", "*/"}}),
			"PowerShell":          NewLanguage("PowerShell", []string{"#"}, [][]string{{"<#", "#>"}}),
			"Polly":               NewLanguage("Polly", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"Protocol Buffers":    NewLanguage("Protocol Buffers", []string{"//"}, [][]string{{"", ""}}),
			"Python":              NewLanguage("Python", []string{"#"}, [][]string{{"\"\"\"", "\"\"\""}}),
			"Q":                   NewLanguage("Q", []string{"/ "}, [][]string{{"\\", "/"}, {"/", "\\"}}),
			"QML":                 NewLanguage("QML", []string{"//"}, [][]string{{"/*", "*/"}}),
			"R":                   NewLanguage("R", []string{"#"}, [][]string{{"", ""}}),
			"Rebol":               NewLanguage("Rebol", []string{";"}, [][]string{{"", ""}}),
			"Red":                 NewLanguage("Red", []string{";"}, [][]string{{"", ""}}),
			"Rego":                NewLanguage("Rego", []string{"#"}, [][]string{{"", ""}}),
			"RMarkdown":           NewLanguage("RMarkdown", []string{}, [][]string{{"", ""}}),
			"RAML":                NewLanguage("RAML", []string{"#"}, [][]string{{"", ""}}),
			"Racket":              NewLanguage("Racket", []string{";"}, [][]string{{"#|", "|#"}}),
			"ReStructuredText":    NewLanguage("ReStructuredText", []string{}, [][]string{{"", ""}}),
			"Ring":                NewLanguage("Ring", []string{"#", "//"}, [][]string{{"/*", "*/"}}),
			"Ruby":                NewLanguage("Ruby", []string{"#"}, [][]string{{":=begin", ":=end"}}),
			"Ruby HTML":           NewLanguage("Ruby HTML", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"Rust":                NewLanguage("Rust", []string{"//", "///", "//!"}, [][]string{{"/*", "*/"}}),
			"Scala":               NewLanguage("Scala", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Sass":                NewLanguage("Sass", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Scheme":              NewLanguage("Scheme", []string{";"}, [][]string{{"#|", "|#"}}),
			"sed":                 NewLanguage("sed", []string{"#"}, [][]string{{"", ""}}),
			"Stan":                NewLanguage("Stan", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Solidity":            NewLanguage("Solidity", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Bourne Shell":        NewLanguage("Bourne Shell", []string{"#"}, [][]string{{"", ""}}),
			"Standard ML":         NewLanguage("Standard ML", []string{}, [][]string{{"(*", "*)"}}),
			"SQL":                 NewLanguage("SQL", []string{"--"}, [][]string{{"/*", "*/"}}),
			"Svelte":              NewLanguage("Svelte", []string{"//"}, [][]string{{"/*", "*/"}, {"<!--", "-->"}}),
			"Swift":               NewLanguage("Swift", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Terra":               NewLanguage("Terra", []string{"--"}, [][]string{{"--[[", "]]"}}),
			"TeX":                 NewLanguage("TeX", []string{"%"}, [][]string{{"", ""}}),
			"Isabelle":            NewLanguage("Isabelle", []string{}, [][]string{{"(*", "*)"}}),
			"TLA":                 NewLanguage("TLA", []string{"\\*"}, [][]string{{"(*", "*)"}}),
			"Tcl/Tk":              NewLanguage("Tcl/Tk", []string{"#"}, [][]string{{"", ""}}),
			"TOML":                NewLanguage("TOML", []string{"#"}, [][]string{{"", ""}}),
			"TypeScript":          NewLanguage("TypeScript", []string{"//"}, [][]string{{"/*", "*/"}}),
			"HCL":                 NewLanguage("HCL", []string{"#", "//"}, [][]string{{"/*", "*/"}}),
			"Umka":                NewLanguage("Umka", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Unity-Prefab":        NewLanguage("Unity-Prefab", []string{}, [][]string{{"", ""}}),
			"MSBuild script":      NewLanguage("MSBuild script", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"Vala":                NewLanguage("Vala", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Verilog":             NewLanguage("Verilog", []string{"//"}, [][]string{{"/*", "*/"}}),
			"VimL":                NewLanguage("VimL", []string{`"`}, [][]string{{"", ""}}),
			"Visual Basic":        NewLanguage("Visual Basic", []string{"'"}, [][]string{{"", ""}}),
			"Vue":                 NewLanguage("Vue", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"Vyper":               NewLanguage("Vyper", []string{"#"}, [][]string{{"\"\"\"", "\"\"\""}}),
			"WiX":                 NewLanguage("WiX", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"XML":                 NewLanguage("XML", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"XML resource":        NewLanguage("XML resource", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"XSLT":                NewLanguage("XSLT", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"XSD":                 NewLanguage("XSD", []string{"<!--"}, [][]string{{"<!--", "-->"}}),
			"YAML":                NewLanguage("YAML", []string{"#"}, [][]string{{"", ""}}),
			"Yacc":                NewLanguage("Yacc", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Yul":                 NewLanguage("Yul", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Zephir":              NewLanguage("Zephir", []string{"//"}, [][]string{{"/*", "*/"}}),
			"Zig":                 NewLanguage("Zig", []string{"//", "///"}, [][]string{{"", ""}}),
			"Zsh":                 NewLanguage("Zsh", []string{"#"}, [][]string{{"", ""}}),
		},
	}
}
