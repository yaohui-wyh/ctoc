package main

import (
	"fmt"

	"github.com/yaohui-wyh/ctoc"
)

func main() {
	languages := ctoc.NewDefinedLanguages()
	options := ctoc.NewClocOptions()
	paths := []string{
		".",
	}

	processor := ctoc.NewProcessor(languages, options)
	result, err := processor.Analyze(paths)
	if err != nil {
		fmt.Printf("ctoc fail. error: %v\n", err)
		return
	}

	for _, item := range result.Files {
		fmt.Println(item)
	}
	fmt.Println(result.Total)
	fmt.Printf("%+v", result)
}
