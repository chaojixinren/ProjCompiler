package spec

type PromptBundle struct {
	PromptText        string
	OutputLanguage    string
	Sections          []PromptSection
	BlockingQuestions []OpenQuestion
	Notes             []string
}

type PromptSection struct {
	Title   string
	Content string
}

func (p PromptBundle) Ready() bool {
	return p.PromptText != "" && len(p.BlockingQuestions) == 0
}
