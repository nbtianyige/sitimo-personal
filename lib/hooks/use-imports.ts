'use client';

import { useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { createImport, deleteImport, getImportStreamUrl, listImports } from '@/lib/api/imports';
import { useImportStore } from '@/lib/store';
import type { ImportJob } from '@/lib/types';

export function useImports(query: { status?: string; page?: number; pageSize?: number }) {
  return useQuery({
    queryKey: ['imports', query],
    queryFn: () => listImports(query),
  });
}

export function useCreateImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createImport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['imports'] });
      toast.success('导入任务已加入队列');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDeleteImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteImport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['imports'] });
      toast.success('导入任务请求已提交');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useImportStream() {
  const queryClient = useQueryClient();
  const syncFromJobs = useImportStore((state) => state.syncFromJobs);

  useEffect(() => {
    let eventSource: EventSource | null = null;
    let retryCount = 0;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let closed = false;

    const applyJob = (job: ImportJob) => {
      if (!job?.id) return;
      queryClient.setQueriesData<{ items: ImportJob[]; total: number; page: number; pageSize: number } | undefined>(
        { queryKey: ['imports'] },
        (current) => {
          if (!current) return current;
          const existing = current.items.filter((item) => item.id !== job.id);
          const items = [job, ...existing].sort((left, right) => right.createdAt.localeCompare(left.createdAt));
          return { ...current, items };
        }
      );
      const latest = queryClient.getQueriesData<{ items: ImportJob[]; total: number; page: number; pageSize: number }>({ queryKey: ['imports'] });
      const firstPage = latest.find((entry) => Array.isArray(entry[1]?.items))?.[1];
      if (firstPage?.items) {
        syncFromJobs(firstPage.items);
      } else {
        syncFromJobs([job]);
      }
    };

    const connect = () => {
      if (closed) return;
      eventSource = new EventSource(getImportStreamUrl());
      eventSource.onopen = () => {
        retryCount = 0;
      };
      eventSource.onmessage = (event) => {
        try {
          const job = JSON.parse(event.data) as ImportJob;
          applyJob(job);
        } catch {
          // Ignore non-job payloads.
        }
      };
      eventSource.onerror = () => {
        eventSource?.close();
        if (closed) return;
        const delay = Math.min(10000, 500 * 2 ** retryCount);
        retryCount += 1;
        reconnectTimer = setTimeout(connect, delay);
      };
    };

    connect();

    return () => {
      closed = true;
      eventSource?.close();
      if (reconnectTimer) clearTimeout(reconnectTimer);
    };
  }, [queryClient, syncFromJobs]);
}
