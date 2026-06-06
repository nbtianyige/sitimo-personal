'use client';

import type { ImportJob, Paginated } from '@/lib/types';
import { apiRequest, API_BASE } from './client';

export async function createImport(formData: FormData) {
  return apiRequest<ImportJob>('/imports', {
    method: 'POST',
    body: formData,
  });
}

export async function listImports(query: { status?: string; page?: number; pageSize?: number }) {
  return apiRequest<Paginated<ImportJob>>('/imports', { query });
}

export async function getImport(id: string) {
  return apiRequest<ImportJob>(`/imports/${id}`);
}

export async function deleteImport(id: string) {
  return apiRequest<{ ok: boolean }>(`/imports/${id}`, { method: 'DELETE' });
}

export function getImportStreamUrl() {
  return `${API_BASE}/imports/stream`;
}
