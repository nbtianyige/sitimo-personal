'use client';

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { ExportJob, ImportJob, Problem } from './types';

interface BasketItem {
  id: string;
  code: string;
  latex: string;
  difficulty: Problem['difficulty'];
}

interface BasketStore {
  items: BasketItem[];
  addItem: (item: BasketItem) => void;
  removeItem: (id: string) => void;
  clearBasket: () => void;
  hasItem: (id: string) => boolean;
}

export const useBasketStore = create<BasketStore>()(
  persist(
    (set, get) => ({
      items: [],
      addItem: (item) =>
        set((state) => {
          if (state.items.some((i) => i.id === item.id)) {
            return state;
          }
          return { items: [...state.items, item] };
        }),
      removeItem: (id) =>
        set((state) => ({
          items: state.items.filter((item) => item.id !== id),
        })),
      clearBasket: () => set({ items: [] }),
      hasItem: (id) => get().items.some((item) => item.id === id),
    }),
    {
      name: 'mathlib-basket',
    }
  )
);

interface SidebarStore {
  isCollapsed: boolean;
  toggle: () => void;
  setCollapsed: (collapsed: boolean) => void;
}

export const useSidebarStore = create<SidebarStore>()(
  persist(
    (set) => ({
      isCollapsed: false,
      toggle: () => set((state) => ({ isCollapsed: !state.isCollapsed })),
      setCollapsed: (collapsed) => set({ isCollapsed: collapsed }),
    }),
    {
      name: 'mathlib-sidebar',
    }
  )
);

type ActiveExportTask = Pick<ExportJob, 'paperId' | 'paperTitle' | 'format' | 'variant' | 'status'> & {
  id: string;
  progress: number;
  startedAt: string;
};

interface ExportStore {
  activeTask: ActiveExportTask | null;
  syncFromJobs: (jobs: ExportJob[]) => void;
  cancelActiveExport: () => void;
}

export const useExportStore = create<ExportStore>()((set, get) => ({
  activeTask: null,
  syncFromJobs: (jobs) => {
    const activeJob = [...jobs].find((job) => job.status === 'processing') ?? jobs.find((job) => job.status === 'pending');

    if (!activeJob) {
      if (get().activeTask) {
        set({ activeTask: null });
      }
      return;
    }

    set({
      activeTask: {
        id: activeJob.id,
        paperId: activeJob.paperId,
        paperTitle: activeJob.paperTitle,
        format: activeJob.format,
        variant: activeJob.variant,
        status: activeJob.status,
        progress: activeJob.progress ?? 0,
        startedAt: activeJob.createdAt,
      },
    });
  },
  cancelActiveExport: () => {
    set({ activeTask: null });
  },
}));

type ActiveImportTask = Pick<ImportJob, 'filename' | 'status'> & {
  id: string;
  progress: number;
  startedAt: string;
};

interface ImportStore {
  activeTask: ActiveImportTask | null;
  syncFromJobs: (jobs: ImportJob[]) => void;
  cancelActiveImport: () => void;
}

export const useImportStore = create<ImportStore>()((set, get) => ({
  activeTask: null,
  syncFromJobs: (jobs) => {
    const activeJob = [...jobs].find((job) => job.status === 'processing') ?? jobs.find((job) => job.status === 'pending');

    if (!activeJob) {
      if (get().activeTask) {
        set({ activeTask: null });
      }
      return;
    }

    set({
      activeTask: {
        id: activeJob.id,
        filename: activeJob.filename,
        status: activeJob.status,
        progress: activeJob.progress ?? 0,
        startedAt: activeJob.createdAt,
      },
    });
  },
  cancelActiveImport: () => {
    set({ activeTask: null });
  },
}));

