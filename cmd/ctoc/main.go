package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/yaohui-wyh/ctoc"
)

// Version is version string for gocloc command
var Version string

// GitCommit is git commit hash string for gocloc command
var GitCommit string

// OutputTypeDefault is cloc's text output format for --output-type option
const OutputTypeDefault string = "default"

// OutputTypeClocXML is Cloc's XML output format for --output-type option
const OutputTypeClocXML string = "cloc-xml"

// OutputTypeSloccount is Sloccount output format for --output-type option
const OutputTypeSloccount string = "sloccount"

// OutputTypeJSON is JSON output format for --output-type option
const OutputTypeJSON string = "json"

const fileHeader string = "File"
const languageHeader string = "Language"
const commonHeader string = "files          blank        comment           code           tokens"
const defaultOutputSeparator string = "-------------------------------------------------------------------------" +
	"-------------------------------------------------------------------------" +
	"-------------------------------------------------------------------------"

var rowLen = 96

// CmdOptions is gocloc command options.
// It is necessary to use notation that follows go-flags.
type CmdOptions struct {
	ByFile         bool   `long:"by-file" description:"report results for every encountered source file"`
	SortTag        string `long:"sort" default:"code" description:"sort based on a certain column" choice:"name" choice:"files" choice:"blank" choice:"comment" choice:"code" choice:"tokens"`
	OutputType     string `long:"output-type" default:"default" description:"output type [values: default,cloc-xml,sloccount,json]"`
	ExcludeExt     string `long:"exclude-ext" description:"exclude file name extensions (separated commas)"`
	IncludeLang    string `long:"include-lang" description:"include language name (separated commas)"`
	Match          string `long:"match" description:"include file name (regex)"`
	NotMatch       string `long:"not-match" description:"exclude file name (regex)"`
	MatchDir       string `long:"match-d" description:"include dir name (regex)"`
	NotMatchDir    string `long:"not-match-d" description:"exclude dir name (regex)"`
	Debug          bool   `long:"debug" description:"dump debug log for developer"`
	SkipDuplicated bool   `long:"skip-duplicated" description:"skip duplicated files"`
	ShowLang       bool   `long:"show-lang" description:"print about all languages and extensions"`
	ShowVersion    bool   `long:"version" description:"print version info"`
}

type outputBuilder struct {
	opts   *CmdOptions
	result *ctoc.Result
}

func newOutputBuilder(result *ctoc.Result, opts *CmdOptions) *outputBuilder {
	return &outputBuilder{
		opts,
		result,
	}
}

func (o *outputBuilder) WriteHeader() {
	maxPathLen := o.result.MaxPathLength
	headerLen := 28
	header := languageHeader

	if o.opts.ByFile {
		headerLen = maxPathLen + 1
		rowLen = maxPathLen + len(commonHeader) + 2
		header = fileHeader
	}
	if o.opts.OutputType == OutputTypeDefault {
		fmt.Printf("%.[2]*[1]s\n", defaultOutputSeparator, rowLen)
		fmt.Printf("%-[2]*[1]s %[3]s\n", header, headerLen, commonHeader)
		fmt.Printf("%.[2]*[1]s\n", defaultOutputSeparator, rowLen)
	}
}

func (o *outputBuilder) WriteFooter() {
	total := o.result.Total
	maxPathLen := o.result.MaxPathLength

	if o.opts.OutputType == OutputTypeDefault {
		fmt.Printf("%.[2]*[1]s\n", defaultOutputSeparator, rowLen)
		if o.opts.ByFile {
			fmt.Printf("%-[1]*[2]v %6[3]v %14[4]v %14[5]v %14[6]v %14[7]v\n",
				maxPathLen, "TOTAL", total.Total, total.Blanks, total.Comments, total.Code, total.Tokens)
		} else {
			fmt.Printf("%-27v %6v %14v %14v %14v %14v\n",
				"TOTAL", total.Total, total.Blanks, total.Comments, total.Code, total.Tokens)
		}
		fmt.Printf("%.[2]*[1]s\n", defaultOutputSeparator, rowLen)
	}
}

func writeResultWithByFile(opts *CmdOptions, result *ctoc.Result) {
	clocFiles := result.Files
	total := result.Total
	maxPathLen := result.MaxPathLength

	var sortedFiles ctoc.ClocFiles
	for _, file := range clocFiles {
		sortedFiles = append(sortedFiles, *file)
	}
	switch opts.SortTag {
	case "name":
		sortedFiles.SortByName()
	case "comment":
		sortedFiles.SortByComments()
	case "blank":
		sortedFiles.SortByBlanks()
	case "tokens":
		sortedFiles.SortByTokens()
	default:
		sortedFiles.SortByCode()
	}

	switch opts.OutputType {
	case OutputTypeClocXML:
		t := ctoc.XMLTotalFiles{
			Code:    total.Code,
			Comment: total.Comments,
			Blank:   total.Blanks,
		}
		f := &ctoc.XMLResultFiles{
			Files: sortedFiles,
			Total: t,
		}
		xmlResult := ctoc.XMLResult{
			XMLFiles: f,
		}
		xmlResult.Encode()
	case OutputTypeSloccount:
		for _, file := range sortedFiles {
			p := ""
			if strings.HasPrefix(file.Name, "./") || string(file.Name[0]) == "/" {
				splitPaths := strings.Split(file.Name, string(os.PathSeparator))
				if len(splitPaths) >= 3 {
					p = splitPaths[1]
				}
			}
			fmt.Printf("%v\t%v\t%v\t%v\n",
				file.Code, file.Lang, p, file.Name)
		}
	case OutputTypeJSON:
		jsonResult := ctoc.NewJSONFilesResultFromCloc(total, sortedFiles)
		buf, err := json.Marshal(jsonResult)
		if err != nil {
			fmt.Println(err)
			panic("json marshal error")
		}
		os.Stdout.Write(buf)
	default:
		for _, file := range sortedFiles {
			clocFile := file
			fmt.Printf("%-[1]*[2]s %21[3]v %14[4]v %14[5]v %14[6]v\n",
				maxPathLen, file.Name, clocFile.Blanks, clocFile.Comments, clocFile.Code, clocFile.Tokens)
		}
	}
}

func (o *outputBuilder) WriteResult() {
	// write header
	o.WriteHeader()

	clocLangs := o.result.Languages
	total := o.result.Total

	if o.opts.ByFile {
		writeResultWithByFile(o.opts, o.result)
	} else {
		var sortedLanguages ctoc.Languages
		for _, language := range clocLangs {
			if len(language.Files) != 0 {
				sortedLanguages = append(sortedLanguages, *language)
			}
		}
		switch o.opts.SortTag {
		case "name":
			sortedLanguages.SortByName()
		case "files":
			sortedLanguages.SortByFiles()
		case "comment":
			sortedLanguages.SortByComments()
		case "blank":
			sortedLanguages.SortByBlanks()
		case "tokens":
			sortedLanguages.SortByTokens()
		default:
			sortedLanguages.SortByCode()
		}

		switch o.opts.OutputType {
		case OutputTypeClocXML:
			xmlResult := ctoc.NewXMLResultFromCloc(total, sortedLanguages, ctoc.XMLResultWithLangs)
			xmlResult.Encode()
		case OutputTypeJSON:
			jsonResult := ctoc.NewJSONLanguagesResultFromCloc(total, sortedLanguages)
			buf, err := json.Marshal(jsonResult)
			if err != nil {
				fmt.Println(err)
				panic("json marshal error")
			}
			os.Stdout.Write(buf)
		default:
			for _, language := range sortedLanguages {
				fmt.Printf("%-27v %6v %14v %14v %14v %14v\n",
					language.Name, len(language.Files), language.Blanks, language.Comments, language.Code, language.Tokens)
			}
		}
	}

	// write footer
	o.WriteFooter()
}

func main() {
	var opts CmdOptions
	clocOpts := ctoc.NewClocOptions()
	// parse command line options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "ctoc"
	parser.Usage = "[OPTIONS] PATH[...]"

	paths, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	// value for language result
	languages := ctoc.NewDefinedLanguages()

	if opts.ShowVersion {
		fmt.Printf("%s (%s)\n", Version, GitCommit)
		return
	}

	if opts.ShowLang {
		fmt.Println(languages.GetFormattedString())
		return
	}

	if len(paths) <= 0 {
		parser.WriteHelp(os.Stdout)
		return
	}

	// check sort tag option with other options
	if opts.ByFile && opts.SortTag == "files" {
		fmt.Println("`--sort files` option cannot be used in conjunction with the `--by-file` option")
		os.Exit(1)
	}

	// setup option for exclude extensions
	for _, ext := range strings.Split(opts.ExcludeExt, ",") {
		e, ok := ctoc.Exts[ext]
		if ok {
			clocOpts.ExcludeExts[e] = struct{}{}
		} else {
			clocOpts.ExcludeExts[ext] = struct{}{}
		}
	}

	// directory and file matching options
	if opts.Match != "" {
		clocOpts.ReMatch = regexp.MustCompile(opts.Match)
	}
	if opts.NotMatch != "" {
		clocOpts.ReNotMatch = regexp.MustCompile(opts.NotMatch)
	}
	if opts.MatchDir != "" {
		clocOpts.ReMatchDir = regexp.MustCompile(opts.MatchDir)
	}
	if opts.NotMatchDir != "" {
		clocOpts.ReNotMatchDir = regexp.MustCompile(opts.NotMatchDir)
	}

	// setup option for include languages
	for _, lang := range strings.Split(opts.IncludeLang, ",") {
		if _, ok := languages.Langs[lang]; ok {
			clocOpts.IncludeLangs[lang] = struct{}{}
		}
	}

	clocOpts.Debug = opts.Debug
	clocOpts.SkipDuplicated = opts.SkipDuplicated

	processor := ctoc.NewProcessor(languages, clocOpts)
	result, err := processor.Analyze(paths)
	if err != nil {
		fmt.Printf("fail ctoc analyze. error: %v\n", err)
		return
	}

	builder := newOutputBuilder(result, &opts)
	builder.WriteResult()
}
