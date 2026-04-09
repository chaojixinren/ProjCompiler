package prompt

import (
	"strings"

	"projcompiler/internal/spec"
)

func renderPrompt(sections []spec.PromptSection) string {
	var builder strings.Builder
	builder.WriteString("Build this project from scratch. Follow the brief exactly. Keep the implementation lean, but do not omit required behavior.\n\n")

	for idx, section := range sections {
		if idx > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(section.Title)
		builder.WriteString("\n")
		builder.WriteString(section.Content)
	}

	builder.WriteString("\n\nIf something is not specified here, choose the narrowest sensible default instead of inventing extra product scope.")
	return builder.String()
}
