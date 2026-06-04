export interface ParseTexFileResult {
  latex: string;
  problemCount: number;
  suggestedSource: string;
  warnings: string[];
}

const SKIP_SECTION_KEYWORDS = ['答案', '解析', '解答', '参考', '简析'];
const PROBLEM_MIN_CHARS = 60;

export function parseTexFile(content: string): ParseTexFileResult {
  const warnings: string[] = [];

  const suggestedSource = extractSuggestedSource(content);

  const docStart = content.indexOf('\\begin{document}');
  let body = docStart >= 0 ? content.slice(docStart + '\\begin{document}'.length) : content;
  const endDocIdx = body.lastIndexOf('\\end{document}');
  if (endDocIdx >= 0) {
    body = body.slice(0, endDocIdx);
  }

  const problems = extractProblems(body, warnings);

  if (problems.length === 0) {
    warnings.push('未检测到题目，请确认文件格式或手动粘贴内容。');
  }

  const latex = problems.map((p) => `\\begin{problem}\n${p}\n\\end{problem}`).join('\n\n');

  return { latex, problemCount: problems.length, suggestedSource, warnings };
}

function extractSuggestedSource(content: string): string {
  const m = content.match(/\\fancyhead\[R\]\{([^}]+)\}/);
  if (!m) return '';
  return m[1].replace(/\\,\s*/g, ' ').replace(/[\x00-\x1f\x7f]/g, '').trim().slice(0, 200);
}

interface Section {
  title: string;
  content: string;
}

// extractBracedGroup extracts {content} with brace-matching at position pos.
// Returns the braced content and the position after the closing }.
function extractBracedGroup(s: string, pos: number): { content: string; endPos: number } | null {
  if (pos >= s.length || s[pos] !== '{') return null;
  let depth = 1;
  let i = pos + 1;
  while (i < s.length && depth > 0) {
    if (s[i] === '{') depth++;
    else if (s[i] === '}') depth--;
    i++;
  }
  if (depth !== 0) return null;
  return { content: s.slice(pos + 1, i - 1), endPos: i };
}

function splitBySections(body: string): Section[] {
  const sections: Section[] = [];
  const re = /\\(?:sub)?section\*?\{/g;
  let lastTitle = '';
  let lastEnd = 0;
  let isFirst = true;
  let m: RegExpExecArray | null;

  while ((m = re.exec(body)) !== null) {
    const cmdEnd = m.index + m[0].length - 1; // position of the opening {
    const braced = extractBracedGroup(body, cmdEnd);
    if (!braced) continue;
    const title = braced.content;
    const fullEnd = braced.endPos;

    if (isFirst) {
      if (m.index > 0) sections.push({ title: '', content: body.slice(0, m.index) });
      isFirst = false;
    } else {
      sections.push({ title: lastTitle, content: body.slice(lastEnd, m.index) });
    }
    lastTitle = title;
    lastEnd = fullEnd;
    re.lastIndex = fullEnd;
  }

  sections.push({ title: lastTitle, content: body.slice(lastEnd) });
  return sections;
}

function isSkipSection(title: string): boolean {
  return SKIP_SECTION_KEYWORDS.some((kw) => title.includes(kw));
}

// --- Enumerate block extraction ---

function findEnumerateBlocks(content: string): string[] {
  const blocks: string[] = [];
  let pos = 0;

  while (pos < content.length) {
    const startIdx = content.indexOf('\\begin{enumerate}', pos);
    if (startIdx < 0) break;

    let searchPos = startIdx + '\\begin{enumerate}'.length;
    if (searchPos < content.length && content[searchPos] === '[') {
      const close = content.indexOf(']', searchPos);
      if (close >= 0) searchPos = close + 1;
    }
    const innerStart = searchPos;

    let depth = 1;
    while (searchPos < content.length && depth > 0) {
      if (content.startsWith('\\begin{enumerate}', searchPos)) {
        depth++;
        searchPos += '\\begin{enumerate}'.length;
      } else if (content.startsWith('\\end{enumerate}', searchPos)) {
        depth--;
        if (depth === 0) break;
        searchPos += '\\end{enumerate}'.length;
      } else {
        searchPos++;
      }
    }

    if (depth === 0) {
      blocks.push(content.slice(innerStart, searchPos));
      pos = searchPos + '\\end{enumerate}'.length;
    } else {
      break;
    }
  }

  return blocks;
}

function extractTopLevelItems(enumContent: string): string[] {
  const items: string[] = [];
  let depth = 0;
  let current = '';
  let pos = 0;

  while (pos < enumContent.length) {
    if (enumContent.startsWith('\\begin{', pos)) {
      depth++;
      const close = enumContent.indexOf('}', pos + 7);
      if (close < 0) {
        current += enumContent[pos++];
        continue;
      }
      current += enumContent.slice(pos, close + 1);
      pos = close + 1;
    } else if (enumContent.startsWith('\\end{', pos)) {
      depth--;
      const close = enumContent.indexOf('}', pos + 5);
      if (close < 0) {
        current += enumContent[pos++];
        continue;
      }
      current += enumContent.slice(pos, close + 1);
      pos = close + 1;
    } else if (depth === 0 && enumContent.startsWith('\\item', pos)) {
      const nextCh = pos + 5 < enumContent.length ? enumContent[pos + 5] : undefined;
      const isItem = nextCh === undefined || nextCh === ' ' || nextCh === '\n' || nextCh === '\t' || nextCh === '[';
      if (isItem) {
        if (current.trim()) items.push(current.trim());
        current = '';
        pos += 5;
        if (pos < enumContent.length && enumContent[pos] === '[') {
          const close = enumContent.indexOf(']', pos);
          if (close >= 0) pos = close + 1;
        }
        continue;
      }
      current += enumContent[pos++];
    } else {
      current += enumContent[pos++];
    }
  }

  if (current.trim()) items.push(current.trim());
  return items;
}

// --- Text-marker pattern extraction ---

function findTextMarkerProblems(content: string): string[] {
  const problems: string[] = [];
  // Find all text-marker positions
  const markers: { index: number }[] = [];

  // Try multiple regex patterns for text markers
  const patterns = [
    // \noindent \textbf{例1.} or \noindent \textbf{例 1.}
    /(?:^|\n)\s*\\noindent\s+\\textbf\{例\s*\d+[.、]?\s*\}/g,
    // \textbf{例1.} style
    /(?:^|\n)\\textbf\{例\s*\d+[.、]?\s*\}/g,
    // Bare-number markers: \textbf{1.} or \textbf{1、}
    /(?:^|\n)\s*\\noindent\s+\\textbf\{\s*\d+[.、]\s*\}/g,
    /(?:^|\n)\\textbf\{\s*\d+[.、]\s*\}/g,
  ];

  for (const re of patterns) {
    re.lastIndex = 0;
    let m: RegExpExecArray | null;
    while ((m = re.exec(content)) !== null) {
      // Start of problem is right after the marker
      const markerEnd = m.index + m[0].length;
      // Skip leading whitespace/newlines
      let bodyStart = markerEnd;
      while (bodyStart < content.length && (content[bodyStart] === ' ' || content[bodyStart] === '\t')) {
        bodyStart++;
      }
      if (content[bodyStart] === '\n') bodyStart++;
      markers.push({ index: bodyStart });
    }
  }

  if (markers.length === 0) return problems;

  // Sort markers by position
  markers.sort((a, b) => a.index - b.index);

  // Deduplicate by index (different regexes may match same position)
  const deduped: { index: number }[] = [];
  for (const m of markers) {
    if (deduped.length === 0 || m.index - deduped[deduped.length - 1].index > 5) {
      deduped.push(m);
    }
  }

  if (deduped.length === 0) return problems;

  // Extract problem content (from marker to next marker or end of content)
  for (let i = 0; i < deduped.length; i++) {
    const start = deduped[i].index;
    const end = i + 1 < deduped.length ? findMarkerStart(content, deduped[i + 1]) : content.length;
    const body = content.slice(start, end).trim();
    if (body.length >= PROBLEM_MIN_CHARS) {
      problems.push(body);
    }
  }

  return problems;
}

function findMarkerStart(content: string, marker: { index: number }): number {
  // Walk back from marker.index to find the beginning of the \textbf{ or \noindent line
  let pos = marker.index;
  while (pos > 0 && content[pos - 1] !== '\n') {
    pos--;
  }
  return pos;
}

// --- Problem detection ---

function isProblemItem(item: string): boolean {
  // Skip items starting with bold knowledge labels like \textbf{定义：}
  if (/^\s*\\textbf\{[^}]*[：:][^}]*\}/.test(item)) return false;
  if (/^\s*\\textbf\{[^}]+\}[：:]/.test(item)) return false;

  // Multiple choice with tasks environment
  if (item.includes('\\begin{tasks}')) return true;
  // Inline A. B. C. D. options (inside itemize or enumerate)
  if (/\\item\s*\[A[\.\s]/.test(item)) return true;
  if (/A\.\s*(\\quad|\s)+B\./.test(item)) return true;
  // Minipage-based ABCD options
  if (/\\begin\{minipage\}/.test(item) && /[AB]\.\s/.test(item)) return true;
  // Fill-in-blank markers
  if (item.includes('\\underline')) return true;
  // Chinese exam blank placeholders
  if (/（\\quad[）\s]|\(\\quad[）\s)]/.test(item)) return true;

  // Problem with clear question verbs and sufficient length
  const stripped = item.replace(/%[^\n]*/g, '').replace(/\s+/g, ' ').trim();
  const startsWithBold = item.trimStart().startsWith('\\textbf');
  if (!startsWithBold && stripped.length >= PROBLEM_MIN_CHARS && /求|证明|解方程|计算|化简|则|等于|判断|若.*则/.test(stripped)) return true;

  // Chinese exam bracket （  ）or ( ) patterns
  if (/（\s*\\quad\s*）|\(\s*\\quad\s*\)/.test(item)) return true;

  return false;
}

// --- Strip solution/answer content from problem body ---

function stripSolutionContent(item: string): string {
  // Remove everything starting from \textbf{答案...}, \textbf{【解析】}, \textbf{【解】}, etc.
  const patterns = [
    /\\textbf\{\s*【解析】\s*\}[\s\S]*$/,
    /\\textbf\{\s*【解】\s*\}[\s\S]*$/,
    /\\textbf\{\s*【详解】\s*\}[\s\S]*$/,
    /\\textbf\{\s*【答案】\s*\}[\s\S]*$/,
    /\\textbf\{\s*解[：:][^}]*\}[\s\S]*$/,
    /\\textbf\{\s*(?:参考)?答案[：:][^}]*\}[\s\S]*$/,
    /\\textbf\{\s*解析[：:][^}]*\}[\s\S]*$/,
  ];

  for (const re of patterns) {
    const idx = item.search(re);
    if (idx >= 0) {
      return item.slice(0, idx).trim();
    }
  }
  return item;
}

function isDuplicateByPrefix(candidate: string, existing: string[], limit: number): boolean {
  if (candidate.length < 40) return false;
  const prefix = candidate.slice(0, 40);
  for (let i = 0; i < limit; i++) {
    if (existing[i].startsWith(prefix)) return true;
  }
  return false;
}

// --- Main extraction ---

function extractProblems(body: string, warnings: string[]): string[] {
  const problems: string[] = [];
  const sections = splitBySections(body);

  for (const section of sections) {
    if (isSkipSection(section.title)) continue;

    // First try enumerate-based extraction
    const enumBlocks = findEnumerateBlocks(section.content);
    for (const block of enumBlocks) {
      for (const item of extractTopLevelItems(block)) {
        if (isProblemItem(item)) {
          problems.push(stripSolutionContent(item));
        }
      }
    }

    // Also try text-marker extraction for sections with non-enumerate problem patterns.
    // A section may contain both enumerate blocks (e.g. ABCD option lists) and
    // text-marker problems (e.g. \textbf{例1.}), so we always run both passes.
    const enumCount = problems.length;
    for (const problem of findTextMarkerProblems(section.content)) {
      const stripped = stripSolutionContent(problem);
      // Avoid double-counting if the same content was already extracted via enumerate
      if (!isDuplicateByPrefix(stripped, problems, enumCount)) {
        problems.push(stripped);
      }
    }
  }

  if (problems.length === 0 && warnings.length === 0) {
    warnings.push('未从 enumerate 环境或文本标记中检测到题目。');
  }

  return problems;
}
