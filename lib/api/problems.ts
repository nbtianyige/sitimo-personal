'use client';

import type { Paginated, Problem, ProblemDetail } from '@/lib/types';
import type { ImportStep, ParsedProblemDraft, SearchCondition } from '@/lib/frontend-contracts';
import { apiRequest } from './client';

export type ProblemListQuery = {
  keyword?: string;
  subject?: string;
  grade?: string;
  difficulty?: string;
  type?: string;
  tagIds?: string;
  hasImage?: string;
  scoreMin?: string | number;
  scoreMax?: string | number;
  sortBy?: string;
  sortOrder?: string;
  page?: number;
  pageSize?: number;
  deleted?: boolean;
};

export type ProblemWriteInput = {
  latex: string;
  answerLatex?: string;
  solutionLatex?: string;
  type: Problem['type'];
  difficulty: Problem['difficulty'];
  subjectiveScore?: number;
  subject?: string;
  grade?: string;
  source?: string;
  tagIds: string[];
  imageIds: string[];
  notes?: string;
};

export type ProblemVersion = {
  id: string;
  problemId: string;
  version: number;
  snapshot: Record<string, unknown>;
  createdAt: string;
};

export type ImportPreviewResponse = {
  parsed: ParsedProblemDraft[];
  errors: Array<{ index: number; message: string }>;
  warnings: string[];
  pairedAnswerFiles?: string[];
  unpairedWarnings?: string[];
};

function normalizeProblem<T extends Problem>(problem: T): T {
  return {
    ...problem,
    tagIds: problem.tagIds ?? [],
    imageIds: problem.imageIds ?? [],
    warnings: problem.warnings ?? [],
  };
}

function normalizeProblemDetail(problem: ProblemDetail): ProblemDetail {
  return {
    ...normalizeProblem(problem),
    tags: problem.tags ?? [],
    images: problem.images ?? [],
  };
}

export async function listProblems(query: ProblemListQuery) {
  const data = await apiRequest<Paginated<Problem>>('/problems', { query });
  return {
    ...data,
    items: data.items.map(normalizeProblem),
  };
}

export async function getProblem(id: string) {
  return normalizeProblemDetail(await apiRequest<ProblemDetail>(`/problems/${id}`));
}

export async function createProblem(input: ProblemWriteInput) {
  const data = await apiRequest<{ problem: ProblemDetail; warnings: string[] }>('/problems', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return {
    ...data,
    problem: normalizeProblemDetail(data.problem),
    warnings: data.warnings ?? [],
  };
}

export async function updateProblem(id: string, input: ProblemWriteInput) {
  const data = await apiRequest<{ problem: ProblemDetail; warnings: string[] }>(`/problems/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return {
    ...data,
    problem: normalizeProblemDetail(data.problem),
    warnings: data.warnings ?? [],
  };
}

export async function deleteProblem(id: string) {
  return apiRequest<{ ok: boolean }>(`/problems/${id}`, { method: 'DELETE' });
}

export async function restoreProblem(id: string) {
  return apiRequest<{ ok: boolean }>(`/problems/${id}/restore`, { method: 'POST' });
}

export async function hardDeleteProblem(id: string) {
  return apiRequest<{ ok: boolean }>(`/problems/${id}/hard`, { method: 'DELETE' });
}

export async function listProblemVersions(id: string) {
  return apiRequest<ProblemVersion[]>(`/problems/${id}/versions`);
}

export async function rollbackProblemVersion(id: string, version: number) {
  return normalizeProblemDetail(await apiRequest<ProblemDetail>(`/problems/${id}/versions/${version}`, { method: 'POST' }));
}

export async function previewBatchImport(
  input:
    | {
        latex?: string;
        separatorStart?: string;
        separatorEnd?: string;
        defaults: Record<string, unknown>;
      }
    | FormData,
) {
  const isFormData = typeof FormData !== 'undefined' && input instanceof FormData;
  return apiRequest<ImportPreviewResponse>('/problems/batch-import/preview', {
    method: 'POST',
    body: isFormData ? input : JSON.stringify(input),
  });
}

export async function commitBatchImport(drafts: ParsedProblemDraft[]) {
  const data = await apiRequest<ProblemDetail[]>('/problems/batch-import', {
    method: 'POST',
    body: JSON.stringify({ drafts }),
  });
  return data.map(normalizeProblemDetail);
}

export async function batchTagProblems(problemIds: string[], tagIds: string[], replace: boolean) {
  return apiRequest<{ ok: boolean }>('/problems/batch-tag', {
    method: 'POST',
    body: JSON.stringify({ problemIds, tagIds, replace }),
  });
}

export async function batchDeleteProblems(problemIds: string[]) {
  return apiRequest<{ ok: boolean }>('/problems/batch-delete', {
    method: 'POST',
    body: JSON.stringify({ problemIds }),
  });
}
