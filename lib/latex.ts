const TEXT_COMMAND_PATTERN = /\\text\{([^{}]+)\}/g;
// Matches \item or \task optionally followed by [label]
const ITEM_PATTERN = /\\(?:item|task)(?:\[[^\]]*\])?/g;

// Detects any math delimiter: \(, \[, $, $$
const HAS_MATH_PATTERN = /\\\(|\\\[|\$\$?/;

export function normalizeLatexForDisplay(value: string): string {
  if (!value) {
    return value;
  }

  // Strip structural/layout LaTeX commands that should not appear in display
  const stripped = stripStructuralCommands(value);

  // Replace \item[label] from description environments with "label: " text
  const withDescLabels = replaceDescriptionItems(stripped);

  // Pre-process \item/\task into A. B. C. D. labels across the whole string
  // (they always appear in text segments, never inside math)
  const preprocessed = replaceItemCommands(withDescLabels);

  // If no math delimiters at all, wrap entire content for MathJax
  if (!HAS_MATH_PATTERN.test(preprocessed)) {
    return `\\(${normalizeMathSegment(preprocessed)}\\)`;
  }

  return rewriteLatexSegments(preprocessed);
}

/**
 * Strip structural/layout LaTeX commands that are not meaningful for display.
 * This runs before math segmentation so it only touches text outside math.
 */
function stripStructuralCommands(value: string): string {
  let s = value;

  // Strip LaTeX comments (% to end of line)
  s = s.replace(/%[^\n]*/g, '');

  // Strip full-document preamble and structural commands
  s = s.replace(/\\documentclass(?:\[[^\]]*\])?\{[^}]*\}\s*/g, '');
  s = s.replace(/\\usepackage(?:\[[^\]]*\])?\{[^}]*\}\s*/g, '');
  s = s.replace(/\\usetikzlibrary\{[^}]*\}\s*/g, '');
  s = s.replace(/\\geometry\{[^}]*\}\s*/g, '');

  // Strip document environment markers, preserving body
  s = s.replace(/\\begin\{document\}\s*/g, '');
  s = s.replace(/\\end\{document\}\s*/g, '');

  // \underline{\hspace{...}} or \underline{\qquad} → ____ (fill-in-blank)
  s = s.replace(/\\underline\{\\(?:hspace\{[^}]*\}|qquad|quad)\}/g, '____');
  // \underline{\hspace{X em}} with space before unit
  s = s.replace(/\\underline\{\\hspace\{[^}]*\}\}/g, '____');
  // bare \underline{...} that contains only whitespace/hspace → ____
  s = s.replace(/\\underline\{\s*\}/g, '____');

  // \noindent must be stripped before replaceNestedBraces removes adjacent commands
  s = s.replace(/\\noindent\b\s*/g, '');

  // \textbf{content} → content
  s = replaceNestedBraces(s, 'textbf');
  // \textit{content} → content
  s = replaceNestedBraces(s, 'textit');
  // \textrm{content} → content
  s = replaceNestedBraces(s, 'textrm');
  // \texttt{content} → content
  s = replaceNestedBraces(s, 'texttt');
  // \emph{content} → content
  s = replaceNestedBraces(s, 'emph');
  // \text{content} is handled later in normalizeTextSegment, skip here

  // \ding{...} → remove
  s = s.replace(/\\ding\{[^}]*\}/g, '');

  // \section{...}, \subsection{...}, \subsubsection{...} → remove
  s = s.replace(/\\(?:sub)*section\*?\{[^}]*\}/g, '');

  // Environment begin/end markers → remove (keep content between them)
  // \begin{tasks}(N) or \begin{tasks} → remove
  s = s.replace(/\\begin\{tasks\}(?:\(\d+\))?\s*/g, '');
  s = s.replace(/\\end\{tasks\}\s*/g, '');
  // \begin{enumerate}[...] or \begin{enumerate} → remove
  s = s.replace(/\\begin\{enumerate\}(?:\[[^\]]*\])?\s*/g, '');
  s = s.replace(/\\end\{enumerate\}\s*/g, '');
  // \begin{itemize}[...] or \begin{itemize} → remove
  s = s.replace(/\\begin\{itemize\}(?:\[[^\]]*\])?\s*/g, '');
  s = s.replace(/\\end\{itemize\}\s*/g, '');
  // Plain \begin{description}...\end{description} (analysis blocks) → strip entirely
  s = s.replace(/\\begin\{description\}(?!\[)[\s\S]*?\\end\{description\}\s*/g, '');
  // \begin{description}[options] (answer choice lists) → strip markers, keep content
  s = s.replace(/\\begin\{description\}\[[^\]]*\]\s*/g, '');
  s = s.replace(/\\end\{description\}\s*/g, '');

  // Layout commands → remove
  s = s.replace(/\\(?:newpage|clearpage|pagebreak)\b\s*/g, '');
  s = s.replace(/\\vspace\*?\{[^}]*\}\s*/g, '');
  s = s.replace(/\\hspace\*?\{[^}]*\}/g, ' ');
  s = s.replace(/\\hfill\b/g, ' ');
  s = s.replace(/\\medskip\b\s*/g, '');
  s = s.replace(/\\bigskip\b\s*/g, '');
  s = s.replace(/\\smallskip\b\s*/g, '');
  s = s.replace(/\\par\b\s*/g, ' ');

  // Collapse multiple spaces/newlines into single space
  s = s.replace(/[ \t]+$/gm, '');
  s = s.replace(/[ \t]+/g, ' ');
  s = s.replace(/\n{3,}/g, '\n\n');
  s = s.trim();

  return s;
}

/**
 * Replace \cmd{content} with content, properly handling nested braces.
 */
function replaceNestedBraces(value: string, cmd: string): string {
  const startMarker = `\\${cmd}{`;
  let result = '';
  let i = 0;

  while (i < value.length) {
    const idx = value.indexOf(startMarker, i);
    if (idx === -1) {
      result += value.slice(i);
      break;
    }

    result += value.slice(i, idx);

    // Find the matching closing brace by counting depth
    let pos = idx + startMarker.length;
    let depth = 1;
    while (pos < value.length && depth > 0) {
      if (value[pos] === '{') depth++;
      else if (value[pos] === '}') depth--;
      if (depth > 0) pos++;
    }

    if (depth === 0) {
      result += value.slice(idx + startMarker.length, pos);
      i = pos + 1;
    } else {
      // Unmatched brace — keep original text as-is
      result += value.slice(idx);
      break;
    }
  }

  return result;
}

/**
 * Replace \item[label] (description env items) with "label: ", preserving label text.
 * Must run before replaceItemCommands so bare \item is still converted to A/B/C/D.
 */
function replaceDescriptionItems(value: string): string {
  return value.replace(/\\item\[([^\]]*)\]/g, '$1: ');
}

/**
 * Replace \item / \task sequences with A. B. C. D. labels.
 * These commands always appear outside math delimiters.
 */
function replaceItemCommands(value: string): string {
  let itemIndex = 0;
  const OPTION_LABELS = ['A', 'B', 'C', 'D', 'E', 'F'];
  return value.replace(ITEM_PATTERN, () => {
    const label = OPTION_LABELS[itemIndex] ?? String(itemIndex + 1);
    itemIndex++;
    return `${label}. `;
  });
}

/**
 * Walk through the string, identify math vs text segments, and apply
 * appropriate transforms. Supports \(...\), \[...\], $...$, $$...$$
 */
function rewriteLatexSegments(value: string): string {
  let cursor = 0;
  let output = '';

  while (cursor < value.length) {
    const next = findNextMathStart(value, cursor);
    if (!next) {
      // Remaining text segment — apply \text{} cleanup
      output += normalizeTextSegment(value.slice(cursor));
      break;
    }

    // Text before math
    output += normalizeTextSegment(value.slice(cursor, next.index));
    output += next.open;

    const end = value.indexOf(next.close, next.index + next.open.length);
    if (end === -1) {
      // Unclosed math — treat rest as math
      output += normalizeMathSegment(value.slice(next.index + next.open.length));
      break;
    }

    output += normalizeMathSegment(value.slice(next.index + next.open.length, end));
    output += next.close;
    cursor = end + next.close.length;
  }

  return output;
}

type MathDelimiter = { open: string; close: string; index: number };

function findNextMathStart(value: string, fromIndex: number): MathDelimiter | null {
  const candidates: MathDelimiter[] = [
    { open: '\\(', close: '\\)', index: value.indexOf('\\(', fromIndex) },
    { open: '\\[', close: '\\]', index: value.indexOf('\\[', fromIndex) },
  ].filter((c) => c.index >= 0);

  // Check for $$ before $ to avoid treating $$ as two separate $
  const dblIdx = value.indexOf('$$', fromIndex);
  if (dblIdx >= 0) {
    candidates.push({ open: '$$', close: '$$', index: dblIdx });
  }

  // Single $ — only if not part of $$
  let singleIdx = value.indexOf('$', fromIndex);
  while (singleIdx >= 0) {
    if (value[singleIdx + 1] !== '$' && (singleIdx === 0 || value[singleIdx - 1] !== '$')) {
      candidates.push({ open: '$', close: '$', index: singleIdx });
      break;
    }
    singleIdx = value.indexOf('$', singleIdx + 1);
  }

  if (candidates.length === 0) return null;
  candidates.sort((a, b) => a.index - b.index);
  return candidates[0];
}

function normalizeTextSegment(segment: string): string {
  let s = segment
    .replace(TEXT_COMMAND_PATTERN, '$1')
    .replace(/\\\\\s*/g, ' ')
    .replace(/\\ /g, ' ')
    .replace(/\\,/g, ' ')
    .replace(/\\;/g, ' ')
    .replace(/\\quad\b/g, '  ')
    .replace(/\\qquad\b/g, '    ');

  const TEXT_SYMBOL_MAP: Record<string, string> = {
    '\\textbullet': '•',
    '\\textendash': '–',
    '\\textemdash': '—',
    '\\textasciitilde': '~',
    '\\textasciicircum': '^',
    '\\textbackslash': '\\',
    '\\textbar': '|',
    '\\textless': '<',
    '\\textgreater': '>',
    '\\textlbrace': '{',
    '\\textrbrace': '}',
    '\\textpercent': '%',
    '\\textdollar': '$',
    '\\textunderscore': '_',
    '\\textampersand': '&',
    '\\textquestiondown': '¿',
    '\\textexclamdown': '¡',
    '\\textsection': '§',
    '\\textparagraph': '¶',
  };

  for (const [cmd, char] of Object.entries(TEXT_SYMBOL_MAP)) {
    const re = new RegExp(cmd.replace(/\\/g, '\\\\') + '\\b', 'g');
    s = s.replace(re, char);
  }

  return s;
}

function normalizeMathSegment(segment: string): string {
  return segment.replace(/°/g, '^\\circ');
}
