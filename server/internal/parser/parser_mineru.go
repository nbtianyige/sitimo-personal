package parser

import (
	"regexp"
	"sort"
	"strings"
)

var (
	mineruAnswerEnvRe   = regexp.MustCompile(`(?s)\\begin\s*\{\s*answer\s*\}(.*?)\\end\s*\{\s*answer\s*\}`)
	mineruSolutionEnvRe = regexp.MustCompile(`(?s)\\begin\s*\{\s*solution\s*\}(.*?)\\end\s*\{\s*solution\s*\}`)
)

// problem-start patterns (line-anchored, matching at the start of a line
// after any horizontal whitespace)
var mineruStartRes = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^[^\S\n]*\d+[\.．。、]`),
	regexp.MustCompile(`(?m)^[^\S\n]*[（(]\d+[)）][\.．。、]?`),
	regexp.MustCompile(`(?m)^[^\S\n]*[一二三四五六七八九十]+[、。．\.]`),
	regexp.MustCompile(`(?m)^[^\S\n]*\\textbf\s*\{\s*(?:\d+|[一二三四五六七八九十]+)[\.．。、]?\s*\}`),
}

// IsMinerUFormat returns true if the raw text contains MinerU-specific markers.
func IsMinerUFormat(raw string) bool {
	return strings.Contains(raw, `\begin{answer}`) ||
		strings.Contains(raw, `\begin{solution}`) ||
		strings.Contains(raw, `\begin{CJK}`)
}

type matchPos struct {
	start int
	end   int
	text  string
}

// findProblemStarts scans the raw text for lines that begin with a problem number.
func findProblemStarts(raw string) []matchPos {
	var rawMatches []matchPos
	for _, re := range mineruStartRes {
		for _, m := range re.FindAllStringIndex(raw, -1) {
			rawMatches = append(rawMatches, matchPos{
				start: m[0],
				end:   m[1],
				text:  raw[m[0]:m[1]],
			})
		}
	}

	sort.Slice(rawMatches, func(i, j int) bool {
		if rawMatches[i].start == rawMatches[j].start {
			return rawMatches[i].end < rawMatches[j].end
		}
		return rawMatches[i].start < rawMatches[j].start
	})

	// Merge overlaps: same start → keep longest; overlapping different starts → keep first.
	var merged []matchPos
	for _, m := range rawMatches {
		if len(merged) > 0 && m.start < merged[len(merged)-1].end {
			last := &merged[len(merged)-1]
			if m.start == last.start && m.end > last.end {
				*last = m
			}
			continue
		}
		merged = append(merged, m)
	}

	return merged
}

// ParseMinerU splits a raw .tex string (typically from MinerU PDF export) into
// individual ProblemBlock entries using heuristic line-anchored problem-number
// detection.
//
// It extracts \begin{answer} ... \end{answer} and \begin{solution} ... \end{solution}
// environments from each problem body, saving them in AnswerLatex / SolutionLatex.
func ParseMinerU(raw string) []ProblemBlock {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	starts := findProblemStarts(raw)
	if len(starts) == 0 {
		// No numbering found – treat the entire text as a single problem
		answer, solution, body := extractInlineAnswerSolution(raw)
		return []ProblemBlock{{
			Body:          body,
			LineStart:     1,
			Label:         "",
			Pattern:       "mineru",
			AnswerLatex:   answer,
			SolutionLatex: solution,
		}}
	}

	var problems []ProblemBlock
	for i, pos := range starts {
		segStart := pos.start
		segEnd := len(raw)
		if i+1 < len(starts) {
			segEnd = starts[i+1].start
		}

		segment := strings.TrimSpace(raw[segStart:segEnd])
		if segment == "" {
			continue
		}

		answer, solution, body := extractInlineAnswerSolution(segment)
		lineNum := strings.Count(raw[:segStart], "\n") + 1

		problems = append(problems, ProblemBlock{
			Body:          body,
			LineStart:     lineNum,
			Label:         cleanLabel(pos.text),
			Pattern:       "mineru",
			AnswerLatex:   answer,
			SolutionLatex: solution,
		})
	}

	return problems
}

// extractInlineAnswerSolution finds \begin{answer} / \begin{solution} environments
// in a problem body and returns (answer, solution, bodyWithoutEnvs).
func extractInlineAnswerSolution(body string) (string, string, string) {
	var ans, sol string
	if m := mineruAnswerEnvRe.FindStringSubmatch(body); m != nil {
		ans = strings.TrimSpace(m[1])
		body = mineruAnswerEnvRe.ReplaceAllString(body, "")
	}
	if m := mineruSolutionEnvRe.FindStringSubmatch(body); m != nil {
		sol = strings.TrimSpace(m[1])
		body = mineruSolutionEnvRe.ReplaceAllString(body, "")
	}
	return ans, sol, strings.TrimSpace(body)
}

// cleanLabel strips simple LaTeX wrappers from a detected problem label.
func cleanLabel(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, `\textbf{`) && strings.HasSuffix(s, `}`) {
		inner := strings.TrimSpace(s[8 : len(s)-1])
		s = strings.TrimRight(inner, " .")
	}
	s = strings.Trim(s, "{}")
	return s
}
