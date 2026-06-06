package parser

import (
	"regexp"
	"strings"
)

var (
	// Block math: $$...$$ (multiline or single line)
	mdBlockMathRe = regexp.MustCompile(`\$\$([\s\S]*?)\$\$`)
	// Inline math: $...$ (non-greedy, single-line)
	mdInlineMathRe = regexp.MustCompile(`\$([^$\n]+?)\$`)
	// Bold: **text**
	mdBoldRe = regexp.MustCompile(`\*\*([^*]+?)\*\*`)
	// Italic: *text*
	// Bold has already been replaced, so remaining *...* is safe to treat as italic.
	mdItalicRe = regexp.MustCompile(`\*([^*]+?)\*`)
)

// MdToTex converts simple Markdown to LaTeX.
//
// Supported conversions:
//   - $$...$$  block math → \[...\]
//   - $...$    inline math → \(...\)
//   - **text** bold        → \textbf{text}
//   - *text*  italic       → \textit{text}
//
// Other Markdown features (tables, links, images, code) are left as-is
// so they can be preserved as plain text.
func MdToTex(content string) string {
	// Protect escaped dollars so they are not matched as inline math.
	const escDollar = "\x00ESC_DOLLAR\x00"
	content = strings.ReplaceAll(content, `\$`, escDollar)

	// Convert block math first, so inline regex does not match the same.
	result := mdBlockMathRe.ReplaceAllString(content, `\[$1\]`)

	// Convert inline math after block math.
	result = mdInlineMathRe.ReplaceAllString(result, `\($1\)`)

	// Restore escaped dollars.
	result = strings.ReplaceAll(result, escDollar, `\$`)

	// Convert bold.
	result = mdBoldRe.ReplaceAllString(result, `\textbf{$1}`)

	// Convert italic.
	result = mdItalicRe.ReplaceAllString(result, `\textit{$1}`)

	return result
}
