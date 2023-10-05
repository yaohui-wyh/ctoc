package gocloc

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

// ClocFile is collecting to line count result.
type ClocFile struct {
	Code     int32  `xml:"code,attr" json:"code"`
	Comments int32  `xml:"comment,attr" json:"comment"`
	Blanks   int32  `xml:"blank,attr" json:"blank"`
	Name     string `xml:"name,attr" json:"name"`
	Lang     string `xml:"language,attr" json:"language"`
}

// ClocFiles is gocloc result set.
type ClocFiles []ClocFile

func (cf ClocFiles) SortByName() {
	sortFunc := func(i, j int) bool {
		return cf[i].Name < cf[j].Name
	}
	sort.Slice(cf, sortFunc)
}

func (cf ClocFiles) SortByComments() {
	sortFunc := func(i, j int) bool {
		if cf[i].Comments == cf[j].Comments {
			return cf[i].Code > cf[j].Code
		}
		return cf[i].Comments > cf[j].Comments
	}
	sort.Slice(cf, sortFunc)
}

func (cf ClocFiles) SortByBlanks() {
	sortFunc := func(i, j int) bool {
		if cf[i].Blanks == cf[j].Blanks {
			return cf[i].Code > cf[j].Code
		}
		return cf[i].Blanks > cf[j].Blanks
	}
	sort.Slice(cf, sortFunc)
}

func (cf ClocFiles) SortByCode() {
	sortFunc := func(i, j int) bool {
		return cf[i].Code > cf[j].Code
	}
	sort.Slice(cf, sortFunc)
}

// AnalyzeFile is analyzing file, this function calls AnalyzeReader() inside.
func AnalyzeFile(filename string, language *Language, opts *ClocOptions) *ClocFile {
	fp, err := os.Open(filename)
	if err != nil {
		// ignore error
		return &ClocFile{Name: filename}
	}
	defer fp.Close()

	return AnalyzeReader(filename, language, fp, opts)
}

// AnalyzeReader is analyzing file for io.Reader.
func AnalyzeReader(filename string, language *Language, file io.Reader, opts *ClocOptions) *ClocFile {
	if opts.Debug {
		fmt.Printf("filename=%v\n", filename)
	}

	clocFile := &ClocFile{
		Name: filename,
		Lang: language.Name,
	}

	isFirstLine := true
	var inComments [][2]string
	buf := getByteSlice()
	defer putByteSlice(buf)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(buf.Bytes(), 1024*1024)

scannerloop:
	for scanner.Scan() {
		lineOrg := scanner.Text()
		line := strings.TrimSpace(lineOrg)

		if len(strings.TrimSpace(line)) == 0 {
			onBlank(clocFile, opts, len(inComments) > 0, line, lineOrg)
			continue
		}

		// shebang line is 'code'
		if isFirstLine && strings.HasPrefix(line, "#!") {
			onCode(clocFile, opts, len(inComments) > 0, line, lineOrg)
			isFirstLine = false
			continue
		}

		if len(inComments) == 0 {
			if isFirstLine {
				line = trimBOM(line)
			}

		singleloop:
			for _, singleComment := range language.lineComments {
				if strings.HasPrefix(line, singleComment) {
					// check if single comment is a prefix of multi comment
					for _, ml := range language.multiLines {
						if ml[0] != "" && strings.HasPrefix(line, ml[0]) {
							break singleloop
						}
					}
					onComment(clocFile, opts, len(inComments) > 0, line, lineOrg)
					continue scannerloop
				}
			}

			if len(language.multiLines) == 0 {
				onCode(clocFile, opts, len(inComments) > 0, line, lineOrg)
				continue scannerloop
			}
		}

		if len(inComments) == 0 && !containsComment(line, language.multiLines) {
			onCode(clocFile, opts, len(inComments) > 0, line, lineOrg)
			continue scannerloop
		}

		lenLine := len(line)
		if len(language.multiLines) == 1 && len(language.multiLines[0]) == 2 && language.multiLines[0][0] == "" {
			onCode(clocFile, opts, len(inComments) > 0, line, lineOrg)
			continue
		}
		codeFlags := make([]bool, len(language.multiLines))
		for pos := 0; pos < lenLine; {
			for idx, ml := range language.multiLines {
				begin, end := ml[0], ml[1]
				lenBegin := len(begin)

				if pos+lenBegin <= lenLine && strings.HasPrefix(line[pos:], begin) && (begin != end || len(inComments) == 0) {
					pos += lenBegin
					inComments = append(inComments, [2]string{begin, end})
					continue
				}

				if n := len(inComments); n > 0 {
					last := inComments[n-1]
					if pos+len(last[1]) <= lenLine && strings.HasPrefix(line[pos:], last[1]) {
						inComments = inComments[:n-1]
						pos += len(last[1])
					}
				} else if pos < lenLine && !unicode.IsSpace(nextRune(line[pos:])) {
					codeFlags[idx] = true
				}
			}
			pos++
		}

		isCode := true
		for _, b := range codeFlags {
			if !b {
				isCode = false
			}
		}

		if isCode {
			onCode(clocFile, opts, len(inComments) > 0, line, lineOrg)
		} else {
			onComment(clocFile, opts, len(inComments) > 0, line, lineOrg)
		}
	}

	return clocFile
}

func onBlank(clocFile *ClocFile, opts *ClocOptions, isInComments bool, line, lineOrg string) {
	clocFile.Blanks++
	if opts.OnBlank != nil {
		opts.OnBlank(line)
	}

	if opts.Debug {
		fmt.Printf("[BLNK, cd:%d, cm:%d, bk:%d, iscm:%v] %s\n",
			clocFile.Code, clocFile.Comments, clocFile.Blanks, isInComments, lineOrg)
	}
}

func onComment(clocFile *ClocFile, opts *ClocOptions, isInComments bool, line, lineOrg string) {
	clocFile.Comments++
	if opts.OnComment != nil {
		opts.OnComment(line)
	}

	if opts.Debug {
		fmt.Printf("[COMM, cd:%d, cm:%d, bk:%d, iscm:%v] %s\n",
			clocFile.Code, clocFile.Comments, clocFile.Blanks, isInComments, lineOrg)
	}
}

func onCode(clocFile *ClocFile, opts *ClocOptions, isInComments bool, line, lineOrg string) {
	clocFile.Code++
	if opts.OnCode != nil {
		opts.OnCode(line)
	}

	if opts.Debug {
		fmt.Printf("[CODE, cd:%d, cm:%d, bk:%d, iscm:%v] %s\n",
			clocFile.Code, clocFile.Comments, clocFile.Blanks, isInComments, lineOrg)
	}
}
