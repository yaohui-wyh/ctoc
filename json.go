package ctoc

// JSONLanguagesResult defines the result of the analysis in JSON format.
type JSONLanguagesResult struct {
	Languages []ClocLanguage `json:"languages"`
	Total     ClocLanguage   `json:"total"`
}

// JSONFilesResult defines the result of the analysis(by files) in JSON format.
type JSONFilesResult struct {
	Files []ClocFile   `json:"files"`
	Total ClocLanguage `json:"total"`
}

// NewJSONLanguagesResultFromCloc returns JSONLanguagesResult with default data set.
func NewJSONLanguagesResultFromCloc(total *Language, sortedLanguages Languages) JSONLanguagesResult {
	var langs []ClocLanguage
	for _, language := range sortedLanguages {
		c := ClocLanguage{
			Name:       language.Name,
			FilesCount: int32(len(language.Files)),
			Code:       language.Code,
			Comments:   language.Comments,
			Blanks:     language.Blanks,
			Tokens:     language.Tokens,
		}
		langs = append(langs, c)
	}
	t := ClocLanguage{
		FilesCount: total.Total,
		Code:       total.Code,
		Comments:   total.Comments,
		Blanks:     total.Blanks,
		Tokens:     total.Tokens,
	}

	return JSONLanguagesResult{
		Languages: langs,
		Total:     t,
	}
}

// NewJSONFilesResultFromCloc returns JSONFilesResult with default data set.
func NewJSONFilesResultFromCloc(total *Language, sortedFiles ClocFiles) JSONFilesResult {
	t := ClocLanguage{
		FilesCount: total.Total,
		Code:       total.Code,
		Comments:   total.Comments,
		Blanks:     total.Blanks,
		Tokens:     total.Tokens,
	}

	return JSONFilesResult{
		Files: sortedFiles,
		Total: t,
	}
}
