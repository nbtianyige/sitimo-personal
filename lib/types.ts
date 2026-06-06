export type Difficulty = 'easy' | 'medium' | 'hard' | 'olympiad';
export type ProblemType = 'multiple_choice' | 'fill_blank' | 'solve' | 'proof' | 'other';

export type Problem = {
  id: string;
  code: string; // 如 "P-2024-0001"
  latex: string; // 题干 LaTeX
  answerLatex?: string; // 答案（录入时一起传）
  solutionLatex?: string; // 解析
  type: ProblemType;
  difficulty: Difficulty;
  subjectiveScore?: number; // 主观难度 1-10（录入时编辑）
  subject?: string;
  grade?: string; // 独立字段,不用标签
  source?: string; // "2023 北京高考"等
  tagIds: string[]; // 知识点标签（扁平）
  imageIds: string[]; // 关联图像
  notes?: string;
  createdAt: string;
  updatedAt: string;
  version: number;
  isDeleted: boolean;
  warnings?: string[];
};

export type ProblemDetail = Problem & {
  tags: Tag[];
  images: ImageAsset[];
};

export type ImageAsset = {
  id: string;
  filename: string;
  mime: string;
  size: number;
  width: number;
  height: number;
  url: string;
  thumbnailUrl: string;
  tagIds: string[];
  linkedProblemIds: string[];
  description?: string;
  createdAt: string;
  updatedAt?: string;
  isDeleted: boolean;
};

export type Tag = {
  id: string;
  name: string;
  category: 'topic' | 'source' | 'custom';
  color: string; // hex
  description?: string;
  problemCount: number;
};

export type Paper = {
  id: string;
  title: string;
  subtitle?: string;
  schoolName?: string;
  examName?: string;
  subject?: string;
  duration?: number;
  totalScore?: number;
  items: PaperItem[];
  layout: {
    columns: 1 | 2;
    fontSize: number;
    lineHeight: number;
    paperSize: 'A4' | 'B5' | 'Letter';
    showAnswerVersion: boolean;
  };
  createdAt: string;
  updatedAt: string;
};

export type PaperStatus = 'draft' | 'completed' | 'review';

export type PaperItem = {
  id: string;
  problemId: string;
  score: number;
  orderIndex: number;
  imagePosition?: 'inline' | 'below' | 'right'; // 图像位置
  blankLines?: number;
};

export type PaperItemDetail = PaperItem & {
  problem?: ProblemDetail;
};

export type PaperDetail = Paper & {
  description?: string;
  status: PaperStatus;
  instructions?: string;
  footerText?: string;
  header: Record<string, unknown>;
  itemDetails: PaperItemDetail[];
};

export type ExportJob = {
  id: string;
  paperId: string;
  paperTitle: string;
  format: 'latex' | 'pdf';
  variant: 'student' | 'answer' | 'both';
  status: 'pending' | 'processing' | 'done' | 'failed';
  progress?: number;
  downloadUrl?: string;
  errorMessage?: string;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
};

export type ImportJob = {
  id: string;
  filename: string;
  inputType: string;
  status: 'pending' | 'processing' | 'done' | 'failed';
  progress: number;
  result?: unknown;
  errorMessage?: string;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
};

export type SearchResult = ProblemDetail & {
  snippet: string;
};

export type SearchHistoryEntry = {
  id: string;
  query: string;
  filters: Record<string, unknown>;
  resultCount: number;
  createdAt: string;
};

export type SavedSearchEntry = {
  id: string;
  name: string;
  query: string;
  filters: Record<string, unknown>;
  createdAt: string;
};

export type Paginated<T> = {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
};

export type MetaStats = {
  problemCount: number;
  imageCount: number;
  tagCount: number;
  exportCount: number;
  recentProblemGain: number;
};

// 难度配置
export const difficultyConfig: Record<Difficulty, { label: string; color: string }> = {
  easy: { label: '容易', color: 'var(--difficulty-easy)' },
  medium: { label: '中等', color: 'var(--difficulty-medium)' },
  hard: { label: '困难', color: 'var(--difficulty-hard)' },
  olympiad: { label: '竞赛', color: 'var(--difficulty-olympiad)' },
};

// 题型配置
export const problemTypeConfig: Record<ProblemType, string> = {
  multiple_choice: '选择题',
  fill_blank: '填空题',
  solve: '解答题',
  proof: '证明题',
  other: '其他',
};
