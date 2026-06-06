package parser

import (
	"fmt"
	"strconv"
	"strings"

	"mathlib/server/internal/domain"
)

// BuildImportPreview orchestrates the full structural parser pipeline.
//
// Pipeline:
//  1. Separate answer files from problem files
//  2. For each problem file: decode → scan → parse → classify → pair → assemble
//  3. Return ImportPreviewResponse with drafts, errors, warnings
func BuildImportPreview(files []domain.UploadedFile, defaults map[string]any) domain.ImportPreviewResponse {
	var problemFiles []domain.UploadedFile
	var answerFiles []domain.UploadedFile

	for _, f := range files {
		if isAnswerFile(f.Filename) {
			answerFiles = append(answerFiles, f)
		} else {
			problemFiles = append(problemFiles, f)
		}
	}

	allFilenames := make([]string, 0, len(files))
	for _, f := range files {
		allFilenames = append(allFilenames, f.Filename)
	}

	var allDrafts []domain.ImportPreviewDraft
	var globalErrors []map[string]any
	var globalWarnings []string
	var pairedFiles []string
	var unpairedWarnings []string

	answerMap := buildAnswerMap(answerFiles)

	for _, pf := range problemFiles {
		fileDrafts, fileErrors, fileWarnings, pairedFilename := processProblemFile(pf, answerMap, allFilenames, defaults)

		allDrafts = append(allDrafts, fileDrafts...)
		for _, e := range fileErrors {
			globalErrors = append(globalErrors, map[string]any{
				"file":    pf.Filename,
				"line":    e.Line,
				"message": e.Message,
			})
		}
		globalWarnings = append(globalWarnings, fileWarnings...)
		if pairedFilename != "" {
			pairedFiles = append(pairedFiles, pairedFilename)
		}
	}

	// Handle orphan answer files (paired to nothing)
	pairedSet := make(map[string]bool)
	for _, f := range pairedFiles {
		pairedSet[f] = true
	}
	for _, af := range answerFiles {
		if !pairedSet[af.Filename] {
			unpairedWarnings = append(unpairedWarnings,
				fmt.Sprintf("答案文件 %q 未能匹配任何题目文件", af.Filename))
		}
	}

	// Warn about problem files without answer candidates
	for _, pf := range problemFiles {
		af := findAnswerFile(pf.Filename, allFilenames, answerMap)
		if af == nil {
			globalWarnings = append(globalWarnings,
				fmt.Sprintf("文件 %q 未找到配套解析文件", pf.Filename))
		}
	}

	return domain.ImportPreviewResponse{
		Parsed:            allDrafts,
		Errors:            globalErrors,
		Warnings:          globalWarnings,
		PairedAnswerFiles: pairedFiles,
		UnpairedWarnings:  unpairedWarnings,
	}
}

func findAnswerFile(problemFilename string, allFilenames []string, answerMap map[string]domain.UploadedFile) *domain.UploadedFile {
	paired, found, _ := PairAnswerFile(problemFilename, allFilenames)
	if found {
		if af, ok := answerMap[paired]; ok {
			return &af
		}
	}
	base := strings.TrimSuffix(problemFilename, ".tex")
	for _, candidate := range []string{base + "_answers.tex"} {
		if af, ok := answerMap[candidate]; ok {
			return &af
		}
	}
	return nil
}

func isAnswerFile(filename string) bool {
	return strings.Contains(filename, "配套解析") ||
		strings.HasSuffix(strings.TrimSuffix(filename, ".tex"), "_answers")
}

func buildAnswerMap(answerFiles []domain.UploadedFile) map[string]domain.UploadedFile {
	m := make(map[string]domain.UploadedFile, len(answerFiles))
	for _, f := range answerFiles {
		m[f.Filename] = f
	}
	return m
}

func processProblemFile(
	pf domain.UploadedFile,
	answerMap map[string]domain.UploadedFile,
	allFilenames []string,
	defaults map[string]any,
) ([]domain.ImportPreviewDraft, []ParseError, []string, string) {
	decoded, _, err := DecodeContent(pf.Content)
	if err != nil {
		return nil, []ParseError{{Line: 0, Message: fmt.Sprintf("decode error: %v", err)}}, nil, ""
	}

	// Convert Markdown to LaTeX if needed
	if strings.HasSuffix(strings.ToLower(pf.Filename), ".md") {
		decoded = MdToTex(decoded)
	}

	blocks := ScanBlocks(decoded)
	blocks = skipPreambleBlocks(blocks)
	blocks = trimTrailingEndDocument(blocks)

	enumerateProblems, enumerateErrors := ParseEnumerate(blocks)
	myboxProblems := ParseMyBox(blocks)

	textMarkerProblems := []ProblemBlock{}
	if len(enumerateProblems) == 0 && len(myboxProblems) == 0 {
		textMarkerProblems = ParseTextMarkers(blocks)
	}

	merged := mergeProblems(mergeProblems(enumerateProblems, myboxProblems), textMarkerProblems)
	allProblems := merged
	allErrors := enumerateErrors

	// Fallback to MinerU heuristic parser if no structural matches found
	if len(allProblems) == 0 {
		mineruProblems := ParseMinerU(decoded)
		if len(mineruProblems) > 0 {
			allProblems = mineruProblems
		}
	}

	if len(allProblems) == 0 {
		return nil, allErrors, []string{fmt.Sprintf("文件 %q 未包含可识别的题目环境 (enumerate, mybox, text markers, 或 MinerU 格式)", pf.Filename)}, ""
	}

	sectionTags := ExtractSectionTags(blocks)

	// Answer pairing
	var answerEntries []AnswerEntry

	af := findAnswerFile(pf.Filename, allFilenames, answerMap)
	if af != nil {
		answerBlocks := scanAnswerFile(*af)
		answerEntries, _ = ExtractAnswers(answerBlocks, len(allProblems))
	}

	drafts := make([]domain.ImportPreviewDraft, 0, len(allProblems))
	for i, pb := range allProblems {
		subBlocks := ScanBlocks(pb.Body)
		inferredType, needsReview := InferType(subBlocks)

		draft := domain.ImportPreviewDraft{
			ID:         newImportID(len(drafts)),
			Title:      formatDraftTitle(len(drafts)+1, pb.SectionTags),
			Latex:      filterLatexComments(pb.Body),
			Difficulty: parseDefaultDifficulty(defaults),
			Status:     "success",
			Warnings:   latexWarnings(pb.Body),
			Subject:    mapStringPtr(defaults, "subject"),
			Grade:      mapStringPtr(defaults, "grade"),
			Source:     mapStringPtr(defaults, "source"),
			TagNames:   parseTagNames(defaults["tagNames"]),
			// New structural fields
			InferredType: inferredType,
			NeedsReview:  needsReview,
			SectionTags:  assignSectionTags(pb, sectionTags),
		}

		if i < len(answerEntries) {
			if answerEntries[i].AnswerLatex != "" {
				s := answerEntries[i].AnswerLatex
				draft.AnswerLatex = &s
			}
			if answerEntries[i].SolutionLatex != "" {
				s := answerEntries[i].SolutionLatex
				draft.SolutionLatex = &s
			}
		}
		// Inline answer/solution from MinerU format overrides paired answer file
		if pb.AnswerLatex != "" {
			s := pb.AnswerLatex
			draft.AnswerLatex = &s
		}
		if pb.SolutionLatex != "" {
			s := pb.SolutionLatex
			draft.SolutionLatex = &s
		}

		drafts = append(drafts, draft)
	}

	var pairedFilename string
	if af != nil {
		pairedFilename = af.Filename
	}
	return drafts, allErrors, nil, pairedFilename
}

func mergeProblems(a, b []ProblemBlock) []ProblemBlock {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	// Union: enumerate problems first, then mybox deduplicated
	seen := make(map[int]bool)
	var result []ProblemBlock
	for _, p := range a {
		seen[p.LineStart] = true
		result = append(result, p)
	}
	for _, p := range b {
		if !seen[p.LineStart] {
			result = append(result, p)
		}
	}
	return result
}

func scanAnswerFile(af domain.UploadedFile) []Block {
	decoded, _, err := DecodeContent(af.Content)
	if err != nil {
		return nil
	}
	return ScanBlocks(decoded)
}

func assignSectionTags(pb ProblemBlock, fileSectionTags []string) []string {
	if len(pb.SectionTags) > 0 {
		return pb.SectionTags
	}
	return fileSectionTags
}

// --- Helper functions (re-implemented from service package) ---

func newImportID(index int) string {
	return fmt.Sprintf("draft-%d", index)
}

func parseDefaultDifficulty(defaults map[string]any) domain.Difficulty {
	if raw, ok := defaults["difficulty"]; ok {
		value := fmt.Sprintf("%v", raw)
		switch value {
		case "easy", "medium", "hard", "olympiad":
			return domain.Difficulty(value)
		}
	}
	return domain.DifficultyMedium
}

func mapStringPtr(input map[string]any, key string) *string {
	raw, ok := input[key]
	if !ok {
		return nil
	}
	value := strings.TrimSpace(fmt.Sprintf("%v", raw))
	if value == "" || value == "<nil>" {
		return nil
	}
	return &value
}

func parseTagNames(raw any) []string {
	switch value := raw.(type) {
	case []string:
		return value
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			text := strings.TrimSpace(fmt.Sprintf("%v", item))
			if text != "" {
				out = append(out, text)
			}
		}
		return out
	case string:
		if value == "" {
			return nil
		}
		parts := strings.Split(value, ",")
		out := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
		return out
	default:
		return nil
	}
}

func skipPreambleBlocks(blocks []Block) []Block {
	for i, b := range blocks {
		if b.Type == BlockEnvBegin && b.EnvName == "document" {
			return blocks[i+1:]
		}
	}
	return blocks
}

func trimTrailingEndDocument(blocks []Block) []Block {
	for i, b := range blocks {
		if b.Type == BlockEnvEnd && b.EnvName == "document" {
			return blocks[:i]
		}
	}
	return blocks
}

func filterLatexComments(body string) string {
	lines := strings.Split(body, "\n")
	var filtered []string
	for _, line := range lines {
		if strings.Contains(line, "% [cite:") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}

func formatDraftTitle(index int, sectionTags []string) string {
	if len(sectionTags) > 0 {
		return fmt.Sprintf("%s - 第%d题", sectionTags[0], index)
	}
	return "题目 #" + strconv.Itoa(index)
}

func latexWarnings(input string) []string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}
	var warnings []string
	if strings.Count(trimmed, "{") != strings.Count(trimmed, "}") {
		warnings = append(warnings, "花括号数量不匹配，已保存原始 LaTeX。")
	}
	if strings.Count(trimmed, `\begin{`) != strings.Count(trimmed, `\end{`) {
		warnings = append(warnings, "LaTeX 环境命令数量不匹配，建议检查 begin/end。")
	}
	return warnings
}
