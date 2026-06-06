'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import {
  ArrowRight,
  BookText,
  File,
  FileText,
  LoaderCircle,
  Search,
  Trash2,
  UploadCloud,
  X,
  AlertCircle,
  Download,
  FileCode2,
} from 'lucide-react';
import { PageHeader, PagePanel, PageShell, PageToolbar } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Progress } from '@/components/ui/progress';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ConfirmActionButton } from '@/components/confirm-action-button';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import {
  EXPORT_STATUS_CLASSNAMES,
  EXPORT_STATUS_FILTER_OPTIONS,
  EXPORT_STATUS_LABELS,
  IMPORT_STATUS_CLASSNAMES,
  IMPORT_STATUS_FILTER_OPTIONS,
  IMPORT_STATUS_LABELS,
} from '@/lib/catalogs';
import { getExportDownloadUrl } from '@/lib/api/exports';
import { useDeleteExport, useExports, useExportStream } from '@/lib/hooks/use-exports';
import { useDeleteImport, useImports, useImportStream } from '@/lib/hooks/use-imports';
import { formatAbsoluteDateTime } from '@/lib/format';
import type { ExportJob, ImportJob } from '@/lib/types';
import { cn } from '@/lib/utils';

const formats = [
  {
    id: 'tex',
    label: 'LaTeX 源码',
    extension: '.tex',
    icon: FileText,
    description: '标准 LaTeX 题目源文件，支持 \\begin{problem} 标记拆题或题号风格。',
    features: ['数学公式原生支持', '图片通过相对路径导入', '结构化标记或题号分割'],
    status: 'ready' as const,
  },
  {
    id: 'md',
    label: 'Markdown',
    extension: '.md',
    icon: BookText,
    description: '包含行内 / 行间数学公式的 Markdown 文档。',
    features: ['$...$/$$...$$ 自动转 LaTeX', '保留文本结构与加粗', '可作为轻量式源导入'],
    status: 'ready' as const,
  },
  {
    id: 'pdf',
    label: 'PDF 试卷',
    extension: '.pdf',
    icon: File,
    description: '扫描版或导出版 PDF 试卷，需经外部服务转换为 .tex/markdown 后入库。',
    features: ['自动识别题型与题号', '图片自动提取', '异步转换 + 实时状态推送'],
    status: 'ready' as const,
  },
];

const ONE_DAY_MS = 24 * 60 * 60 * 1000;
const ONE_WEEK_MS = 7 * 24 * 60 * 60 * 1000;
const ONE_MONTH_MS = 30 * 24 * 60 * 60 * 1000;

export default function ImportCenterPage() {
  return (
    <PageShell>
      <PageHeader
        eyebrow="数据导入"
        title="导入中心"
        description="将外部题目批量导入题库。支持多种格式，统一走「上传 → 预览 → 确认入库」流程。"
        actions={
          <Button asChild>
            <Link href="/import/wizard">
              <UploadCloud className="mr-2 h-4 w-4" />
              开始导入
            </Link>
          </Button>
        }
      />

      <div className="grid gap-6 md:grid-cols-3">
        {formats.map((format) => {
          const Icon = format.icon;
          return (
            <PagePanel key={format.id} className="relative flex flex-col">
              <div className="space-y-4 p-5">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                  <Icon className="h-5 w-5 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold text-foreground">
                    {format.label}
                    <span className="ml-2 font-mono text-xs text-muted-foreground">{format.extension}</span>
                  </h3>
                  <p className="mt-1 text-sm leading-6 text-muted-foreground">{format.description}</p>
                </div>
                <ul className="space-y-2">
                  {format.features.map((feature) => (
                    <li key={feature} className="flex items-start gap-2 text-sm text-muted-foreground">
                      <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary/60" />
                      {feature}
                    </li>
                  ))}
                </ul>
              </div>
            </PagePanel>
          );
        })}
      </div>

      <div className="mt-6 flex justify-end">
        <Button asChild size="lg">
          <Link href="/import/wizard">
            进入导入向导
            <ArrowRight className="ml-2 h-4 w-4" />
          </Link>
        </Button>
      </div>

      <div className="mt-10">
        <h2 className="text-lg font-semibold">任务列表</h2>
        <p className="text-sm text-muted-foreground">查看导入与导出的历史任务状态</p>
      </div>

      <TaskTabs />
    </PageShell>
  );
}

function TaskTabs() {
  useImportStream();
  useExportStream();

  return (
    <Tabs defaultValue="import" className="mt-4">
      <TabsList>
        <TabsTrigger value="import">导入任务</TabsTrigger>
        <TabsTrigger value="export">导出任务</TabsTrigger>
      </TabsList>
      <TabsContent value="import">
        <ImportTaskPanel />
      </TabsContent>
      <TabsContent value="export">
        <ExportTaskPanel />
      </TabsContent>
    </Tabs>
  );
}

function ImportTaskPanel() {
  const [query, setQuery] = useState('');
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([]);
  const [dateFilter, setDateFilter] = useState('all');
  const now = useCurrentTimestamp();

  const importsQuery = useImports({ page: 1, pageSize: 100 });
  const deleteImportMutation = useDeleteImport();

  const rows = useMemo(() => importsQuery.data?.items ?? [], [importsQuery.data?.items]);

  const filteredRows = useMemo(
    () =>
      rows.filter((row) => {
        const matchesQuery = !query.trim() || row.filename.toLowerCase().includes(query.trim().toLowerCase());
        const matchesStatus = selectedStatuses.length === 0 || selectedStatuses.includes(row.status);
        const age = now - new Date(row.createdAt).getTime();
        const matchesDate =
          dateFilter === 'all' ||
          (dateFilter === 'today' && age <= ONE_DAY_MS) ||
          (dateFilter === 'week' && age <= ONE_WEEK_MS) ||
          (dateFilter === 'month' && age <= ONE_MONTH_MS);
        return matchesQuery && matchesStatus && matchesDate;
      }),
    [dateFilter, now, query, rows, selectedStatuses]
  );

  const activeRow = filteredRows.find((row) => row.status === 'processing') ?? filteredRows.find((row) => row.status === 'pending');

  const toggleStatus = (status: string) => {
    setSelectedStatuses((current) => (current.includes(status) ? current.filter((item) => item !== status) : [...current, status]));
  };

  return (
    <div className="space-y-4">
      {activeRow ? (
        <PagePanel className="border-primary/20 bg-primary/5">
          <div className="flex items-center justify-between gap-4 p-5">
            <div className="flex-1">
              <p className="font-medium">
                {activeRow.status === 'processing' ? '正在处理' : '正在排队'}《{activeRow.filename}》
              </p>
              <div className="mt-3 flex items-center gap-3">
                <Progress value={activeRow.progress ?? 0} className="flex-1" />
                <span className="text-sm text-muted-foreground">{activeRow.progress ?? 0}%</span>
              </div>
              <p className="mt-2 text-xs text-muted-foreground">已用时 {formatElapsed(activeRow, now)}</p>
            </div>
            <ConfirmActionButton
              variant="ghost"
              size="icon"
              onConfirm={() => deleteImportMutation.mutateAsync(activeRow.id)}
              pending={deleteImportMutation.isPending}
              title={activeRow.status === 'processing' ? '确认取消导入任务' : '确认移除排队任务'}
              description={`确定要终止《${activeRow.filename}》的当前导入任务吗？`}
              confirmLabel="确认终止"
            >
              <X className="h-4 w-4" />
            </ConfirmActionButton>
          </div>
        </PagePanel>
      ) : null}

      <PageToolbar className="flex flex-col gap-3">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center">
          <div className="relative min-w-[260px] flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索文件名..." aria-label="搜索文件名" className="pl-9" />
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <Select value={dateFilter} onValueChange={setDateFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="日期" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部日期</SelectItem>
                <SelectItem value="today">今天</SelectItem>
                <SelectItem value="week">近 7 天</SelectItem>
                <SelectItem value="month">近 30 天</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {IMPORT_STATUS_FILTER_OPTIONS.map((status) => (
            <button
              key={status.value}
              type="button"
              onClick={() => toggleStatus(status.value)}
              aria-label={`筛选: ${status.label}`}
              className={`rounded-full border px-3 py-1.5 text-sm transition-colors ${
                selectedStatuses.includes(status.value)
                  ? 'border-primary bg-primary/8 text-primary'
                  : 'border-border text-muted-foreground'
              }`}
            >
              {status.label}
            </button>
          ))}
        </div>
      </PageToolbar>

      <PagePanel className="overflow-hidden">
        {filteredRows.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="py-3 px-4">文件名</TableHead>
                <TableHead className="py-3 px-4">状态</TableHead>
                <TableHead className="py-3 px-4">进度</TableHead>
                <TableHead className="py-3 px-4">创建时间</TableHead>
                <TableHead className="py-3 px-4">耗时</TableHead>
                <TableHead className="text-right py-3 px-4">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredRows.map((row) => (
                <TableRow key={row.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <FileText className="h-4 w-4 text-primary" />
                      <span>{row.filename}</span>
                    </div>
                  </TableCell>
                  <TableCell>{renderImportStatus(row)}</TableCell>
                  <TableCell>
                    {row.status === 'processing' ? (
                      <div className="flex items-center gap-2">
                        <Progress value={row.progress ?? 0} className="w-24" />
                        <span className="text-sm text-muted-foreground">{row.progress ?? 0}%</span>
                      </div>
                    ) : (
                      <span className="text-sm text-muted-foreground">-</span>
                    )}
                  </TableCell>
                  <TableCell>{formatAbsoluteDateTime(row.createdAt)}</TableCell>
                  <TableCell>{formatElapsed(row, now)}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      {row.status === 'failed' ? (
                        <Dialog>
                          <DialogTrigger asChild>
                            <Button variant="ghost" size="icon" aria-label="查看错误">
                              <AlertCircle className="h-4 w-4 text-destructive" />
                            </Button>
                          </DialogTrigger>
                          <DialogContent>
                            <DialogHeader>
                              <DialogTitle>错误详情</DialogTitle>
                            </DialogHeader>
                            <pre className="whitespace-pre-wrap rounded-xl bg-muted p-4 text-sm">{row.errorMessage ?? '未知错误'}</pre>
                          </DialogContent>
                        </Dialog>
                      ) : null}
                      <ConfirmActionButton
                        variant="ghost"
                        size="icon"
                        onConfirm={() => deleteImportMutation.mutateAsync(row.id)}
                        pending={deleteImportMutation.isPending}
                        title={row.status === 'pending' || row.status === 'processing' ? '确认终止导入任务' : '确认删除导入记录'}
                        description={
                          row.status === 'pending' || row.status === 'processing'
                            ? `确定要终止《${row.filename}》的导入任务吗？`
                            : `确定要删除《${row.filename}》的导入记录吗？`
                        }
                        confirmLabel={row.status === 'pending' || row.status === 'processing' ? '确认终止' : '确认删除'}
                      >
                        {row.status === 'pending' || row.status === 'processing' ? <X className="h-4 w-4" /> : <Trash2 className="h-4 w-4" />}
                      </ConfirmActionButton>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <Empty className="min-h-[320px] border-none bg-transparent">
            <EmptyMedia variant="icon">
              <FileText className="h-6 w-6" />
            </EmptyMedia>
            <EmptyHeader>
              <EmptyTitle>没有匹配的导入记录</EmptyTitle>
              <EmptyDescription>试试调整状态或日期筛选条件。</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </PagePanel>
    </div>
  );
}

function ExportTaskPanel() {
  const [query, setQuery] = useState('');
  const [selectedStatuses, setSelectedStatuses] = useState<string[]>([]);
  const [formatFilter, setFormatFilter] = useState('all');
  const [dateFilter, setDateFilter] = useState('all');
  const now = useCurrentTimestamp();

  const exportsQuery = useExports({ page: 1, pageSize: 100 });
  const deleteExportMutation = useDeleteExport();

  const rows = useMemo(() => exportsQuery.data?.items ?? [], [exportsQuery.data?.items]);

  const filteredRows = useMemo(
    () =>
      rows.filter((row) => {
        const matchesQuery = !query.trim() || row.paperTitle.toLowerCase().includes(query.trim().toLowerCase());
        const matchesStatus = selectedStatuses.length === 0 || selectedStatuses.includes(row.status);
        const matchesFormat = formatFilter === 'all' || row.format === formatFilter;
        const age = now - new Date(row.createdAt).getTime();
        const matchesDate =
          dateFilter === 'all' ||
          (dateFilter === 'today' && age <= ONE_DAY_MS) ||
          (dateFilter === 'week' && age <= ONE_WEEK_MS) ||
          (dateFilter === 'month' && age <= ONE_MONTH_MS);
        return matchesQuery && matchesStatus && matchesFormat && matchesDate;
      }),
    [dateFilter, formatFilter, now, query, rows, selectedStatuses]
  );

  const activeRow = filteredRows.find((row) => row.status === 'processing') ?? filteredRows.find((row) => row.status === 'pending');

  const toggleStatus = (status: string) => {
    setSelectedStatuses((current) => (current.includes(status) ? current.filter((item) => item !== status) : [...current, status]));
  };

  return (
    <div className="space-y-4">
      {activeRow ? (
        <PagePanel className="border-primary/20 bg-primary/5">
          <div className="flex items-center justify-between gap-4 p-5">
            <div className="flex-1">
              <p className="font-medium">
                {activeRow.status === 'processing' ? '正在生成' : '正在排队'}《{activeRow.paperTitle}》
              </p>
              <div className="mt-3 flex items-center gap-3">
                <Progress value={activeRow.progress ?? 0} className="flex-1" />
                <span className="text-sm text-muted-foreground">{activeRow.progress ?? 0}%</span>
              </div>
              <p className="mt-2 text-xs text-muted-foreground">已用时 {formatElapsed(activeRow, now)}</p>
            </div>
            <ConfirmActionButton
              variant="ghost"
              size="icon"
              onConfirm={() => deleteExportMutation.mutateAsync(activeRow.id)}
              pending={deleteExportMutation.isPending}
              title={activeRow.status === 'processing' ? '确认取消导出任务' : '确认移除排队任务'}
              description={`确定要终止《${activeRow.paperTitle}》的当前导出任务吗？`}
              confirmLabel="确认终止"
            >
              <X className="h-4 w-4" />
            </ConfirmActionButton>
          </div>
        </PagePanel>
      ) : null}

      <PageToolbar className="flex flex-col gap-3">
        <div className="flex flex-col gap-3 xl:flex-row xl:items-center">
          <div className="relative min-w-[260px] flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索任务标题..." aria-label="搜索任务标题" className="pl-9" />
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <Select value={formatFilter} onValueChange={setFormatFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="格式" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部格式</SelectItem>
                <SelectItem value="pdf">PDF</SelectItem>
                <SelectItem value="latex">LaTeX</SelectItem>
              </SelectContent>
            </Select>

            <Select value={dateFilter} onValueChange={setDateFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="日期" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部日期</SelectItem>
                <SelectItem value="today">今天</SelectItem>
                <SelectItem value="week">近 7 天</SelectItem>
                <SelectItem value="month">近 30 天</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {EXPORT_STATUS_FILTER_OPTIONS.map((status) => (
            <button
              key={status.value}
              type="button"
              onClick={() => toggleStatus(status.value)}
              aria-label={`筛选: ${status.label}`}
              className={`rounded-full border px-3 py-1.5 text-sm transition-colors ${
                selectedStatuses.includes(status.value)
                  ? 'border-primary bg-primary/8 text-primary'
                  : 'border-border text-muted-foreground'
              }`}
            >
              {status.label}
            </button>
          ))}
        </div>
      </PageToolbar>

      <PagePanel className="overflow-hidden">
        {filteredRows.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="py-3 px-4">任务标题</TableHead>
                <TableHead className="py-3 px-4">格式</TableHead>
                <TableHead className="py-3 px-4">版本</TableHead>
                <TableHead className="py-3 px-4">状态</TableHead>
                <TableHead className="py-3 px-4">创建时间</TableHead>
                <TableHead className="py-3 px-4">耗时</TableHead>
                <TableHead className="text-right py-3 px-4">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredRows.map((row) => (
                <TableRow key={row.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      {row.format === 'pdf' ? <FileText className="h-4 w-4 text-primary" /> : <FileCode2 className="h-4 w-4 text-primary" />}
                      <span>{row.paperTitle}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{row.format.toUpperCase()}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{row.variant === 'student' ? '学生' : row.variant === 'answer' ? '答案' : '双版本'}</Badge>
                  </TableCell>
                  <TableCell>{renderExportStatus(row)}</TableCell>
                  <TableCell>{formatAbsoluteDateTime(row.createdAt)}</TableCell>
                  <TableCell>{formatElapsed(row, now)}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      {row.status === 'done' ? (
                        <Button variant="ghost" size="icon" aria-label="下载导出文件" asChild>
                          <a href={getExportDownloadUrl(row.id)} download>
                            <Download className="h-4 w-4" />
                          </a>
                        </Button>
                      ) : null}

                      {row.status === 'failed' ? (
                        <Dialog>
                          <DialogTrigger asChild>
                            <Button variant="ghost" size="icon" aria-label="查看错误">
                              <AlertCircle className="h-4 w-4 text-destructive" />
                            </Button>
                          </DialogTrigger>
                          <DialogContent>
                            <DialogHeader>
                              <DialogTitle>错误详情</DialogTitle>
                            </DialogHeader>
                            <pre className="whitespace-pre-wrap rounded-xl bg-muted p-4 text-sm">{row.errorMessage ?? '未知错误'}</pre>
                          </DialogContent>
                        </Dialog>
                      ) : null}

                      <ConfirmActionButton
                        variant="ghost"
                        size="icon"
                        onConfirm={() => deleteExportMutation.mutateAsync(row.id)}
                        pending={deleteExportMutation.isPending}
                        title={row.status === 'pending' || row.status === 'processing' ? '确认终止导出任务' : '确认删除导出记录'}
                        description={
                          row.status === 'pending' || row.status === 'processing'
                            ? `确定要终止《${row.paperTitle}》的导出任务吗？`
                            : `确定要删除《${row.paperTitle}》的导出记录吗？`
                        }
                        confirmLabel={row.status === 'pending' || row.status === 'processing' ? '确认终止' : '确认删除'}
                      >
                        {row.status === 'pending' || row.status === 'processing' ? <X className="h-4 w-4" /> : <Trash2 className="h-4 w-4" />}
                      </ConfirmActionButton>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <Empty className="min-h-[320px] border-none bg-transparent">
            <EmptyMedia variant="icon">
              <FileText className="h-6 w-6" />
            </EmptyMedia>
            <EmptyHeader>
              <EmptyTitle>没有匹配的导出记录</EmptyTitle>
              <EmptyDescription>试试调整状态、日期或格式筛选条件。</EmptyDescription>
            </EmptyHeader>
          </Empty>
        )}
      </PagePanel>
    </div>
  );
}

function renderImportStatus(row: ImportJob) {
  return (
    <Badge variant="outline" className={cn('gap-2', IMPORT_STATUS_CLASSNAMES[row.status])}>
      {row.status === 'processing' && <LoaderCircle className="h-4 w-4 animate-spin" />}
      {IMPORT_STATUS_LABELS[row.status]}
    </Badge>
  );
}

function renderExportStatus(row: ExportJob) {
  return (
    <Badge variant="outline" className={cn('gap-2', EXPORT_STATUS_CLASSNAMES[row.status])}>
    {row.status === 'processing' && <LoaderCircle className="h-4 w-4 animate-spin" />}
    {EXPORT_STATUS_LABELS[row.status]}
  </Badge>
  );
}

function formatElapsed(row: { createdAt: string; startedAt?: string; completedAt?: string }, now: number) {
  const start = row.startedAt ? new Date(row.startedAt).getTime() : new Date(row.createdAt).getTime();
  const end = row.completedAt ? new Date(row.completedAt).getTime() : now;
  const minutes = Math.max(0, Math.round((end - start) / 60000));
  if (minutes < 1) return '少于 1 分钟';
  if (minutes < 60) return `${minutes} 分钟`;
  const hours = Math.floor(minutes / 60);
  const remain = minutes % 60;
  return remain > 0 ? `${hours} 小时 ${remain} 分钟` : `${hours} 小时`;
}

function useCurrentTimestamp(intervalMs = 60_000) {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), intervalMs);
    return () => window.clearInterval(timer);
  }, [intervalMs]);

  return now;
}
