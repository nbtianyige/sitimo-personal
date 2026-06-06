package parser

// BlockType represents the type of a parsed LaTeX block.
type BlockType int

const (
	BlockEnvBegin  BlockType = iota // \begin{env}
	BlockEnvEnd                      // \end{env}
	BlockItem                        // \item
	BlockCommand                     // \section*, \subsection*, etc.
	BlockText                        // plain text
	BlockComment                     // % comment
)

// Block represents a single parsed LaTeX block element.
type Block struct {
	Type      BlockType
	Content   string // raw text content
	LineStart int    // line number in source
	Label     string // for \item[optional] or \begin{env}[args]
	EnvName   string // environment name for EnvBegin/EnvEnd
	EnvArgs   string // env options like [label=\textbf{题\arabic*}]
}

// ProblemBlock represents an extracted problem from the parsed result.
type ProblemBlock struct {
	Body          string   // problem body (without enumerate wrapping)
	LineStart     int      // start line
	Label         string   // item label text
	Pattern       string   // "A", "B", "C", "D" or "mineru"
	SectionTags   []string
	AnswerLatex   string   // extracted inline answer (MinerU style)
	SolutionLatex string   // extracted inline solution (MinerU style)
}

// ParseResult holds the overall parsing result.
type ParseResult struct {
	Problems []ProblemBlock
	Errors   []ParseError
}

// ParseError records a parsing error with its source location.
type ParseError struct {
	Line    int
	Message string
}
