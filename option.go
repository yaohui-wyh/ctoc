package ctoc

import (
	"regexp"

	"github.com/pkoukk/tiktoken-go"
)

// ClocOptions is gocloc processor options.
type ClocOptions struct {
	Debug          bool
	SkipDuplicated bool
	ExcludeExts    map[string]struct{}
	IncludeLangs   map[string]struct{}
	ReNotMatch     *regexp.Regexp
	ReMatch        *regexp.Regexp
	ReNotMatchDir  *regexp.Regexp
	ReMatchDir     *regexp.Regexp
	Tokenizer      *tiktoken.Tiktoken

	// OnCode is triggered for each line of code.
	OnCode func(line string)
	// OnBlack is triggered for each blank line.
	OnBlank func(line string)
	// OnComment is triggered for each line of comments.
	OnComment func(line string)
}

// NewClocOptions create new ClocOptions with default values.
func NewClocOptions() *ClocOptions {
	tke, _ := tiktoken.GetEncoding("cl100k_base")
	return &ClocOptions{
		Debug:          false,
		SkipDuplicated: false,
		ExcludeExts:    make(map[string]struct{}),
		IncludeLangs:   make(map[string]struct{}),
		Tokenizer:      tke,
	}
}
