'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Check, Eye, Filter, Image as ImageIcon, LayoutGrid, List, Pencil, Plus, Search, ShoppingBasket, Trash2, X,
  FileUp,
} from 'lucide-react';
import { MathText } from '@/components/math-text';
import { ConfirmActionButton } from '@/components/confirm-action-button';
import { PageHeader, PagePanel, PageShell, PageToolbar } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Input } from '@/components/ui/input';
import { Pagination, PaginationContent, PaginationItem, PaginationNext, PaginationPrevious } from '@/components/ui/pagination';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { buildGradeOptions, STANDARD_GRADES } from '@/lib/constants';
import { formatRelativeTime } from '@/lib/format';
import { useGrades } from '@/lib/hooks/use-meta';
import { useBatchDeleteProblems, useProblems } from '@/lib/hooks/use-problems';
import { useTags } from '@/lib/hooks/use-tags';
import { useBasketStore } from '@/lib/store';
import { difficultyConfig, type Difficulty, type Problem, type Tag } from '@/lib/types';
import { cn } from '@/lib/utils';

const difficultyOptions: Difficulty[] = ['easy', 'medium', 'hard', 'olympiad'];

export default function ProblemsPage() {
  const [search, setSearch] = useState('');
  const [grade, setGrade] = useState('all');
  const [selectedDifficulties, setSelectedDifficulties] = useState<Difficulty[]>([]);
  const [tagId, setTagId] = useState('all');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [page, setPage] = useState(1);
  const [selectedProblems, setSelectedProblems] = useState<string[]>([]);

  const gradesQuery = useGrades();
  const tagsQuery = useTags();
  const problemsQuery = useProblems({
    keyword: search,
    grade: grade === 'all' ? undefined : grade,
    difficulty: selectedDifficulties.length > 0 ? selectedDifficulties.join(',') : undefined,
    tagIds: tagId === 'all' ? undefined : tagId,
    sortBy: 'updated_at',
    sortOrder: 'desc',
    page,
    pageSize: 24,
  });
  const batchDeleteMutation = useBatchDeleteProblems();

  const addItem = useBasketStore((state) => state.addItem);
  const removeItem = useBasketStore((state) => state.removeItem);
  const hasItem = useBasketStore((state) => state.hasItem);

  const gradeOptions = useMemo(
    () => buildGradeOptions(gradesQuery.data ?? STANDARD_GRADES),
    [gradesQuery.data]
  );
  const problems = problemsQuery.data?.items ?? [];
  const total = problemsQuery.data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / 24));
  const tagMap = useMemo(() => new Map((tagsQuery.data ?? []).map((tag) => [tag.id, tag])), [tagsQuery.data]);
  const activeFilterCount = [grade !== 'all', tagId !== 'all', selectedDifficulties.length > 0, Boolean(search.trim())].filter(Boolean).length;
  const activeFilters = useMemo(() => {
    const filters: string[] = [];

    if (search.trim()) {
      filters.push(`关键词：${search.trim()}`);
    }
    if (grade !== 'all') {
      filters.push(`年级：${grade}`);
    }
    if (tagId !== 'all') {
      const tagName = tagMap.get(tagId)?.name ?? '已选标签';
      filters.push(`标签：${tagName}`);
    }
    if (selectedDifficulties.length > 0) {
      selectedDifficulties.forEach((difficulty) => {
        filters.push(`难度：${difficultyConfig[difficulty].label}`);
      });
    }

    return filters;
  }, [grade, search, selectedDifficulties, tagId, tagMap]);

  const resetFilters = () => {
    setSearch('');
    setGrade('all');
    setTagId('all');
    setSelectedDifficulties([]);
    setPage(1);
  };

  const toggleDifficulty = (difficulty: Difficulty) => {
    setSelectedDifficulties((current) =>
      current.includes(difficulty) ? current.filter((item) => item !== difficulty) : [...current, difficulty]
    );
    setPage(1);
  };

  const handleBatchDelete = async () => {
    await batchDeleteMutation.mutateAsync(selectedProblems);
    setSelectedProblems([]);
  };

  return (
    <PageShell wide>
      <PageHeader
        eyebrow="题库"
        title="题目管理"
        description="按年级、标签和难度实时筛选题目；批量选择后可统一清理或加入题目篮子。"
        badges={
          <>
            <Badge variant="secondary">共 {total} 道题</Badge>
            <Badge variant="secondary">启用筛选 {activeFilterCount}</Badge>
          </>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button variant="outline" asChild>
              <Link href="/import/wizard">
                <FileUp className="mr-2 h-4 w-4" />
                导入
              </Link>
            </Button>
            <Button asChild>
              <Link href="/problems/new">
                <Plus className="mr-2 h-4 w-4" />
                新建题目
              </Link>
            </Button>
          </div>
        }
      />

      <div className="grid items-start gap-6 xl:grid-cols-[320px_minmax(0,1fr)]">
        <PagePanel className="self-start xl:sticky xl:top-20">
          <div className="space-y-5 p-5">
            <div className="flex items-start justify-between gap-3">
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-sm font-medium">
                  <Filter className="h-4 w-4" />
                  筛选条件
                </div>
                <p className="text-sm leading-6 text-muted-foreground">筛选项会直接请求真实数据，调整后立即刷新右侧结果。</p>
              </div>
              <Button type="button" variant="ghost" size="sm" onClick={resetFilters} disabled={activeFilterCount === 0}>
                清空
              </Button>
            </div>

            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(event) => {
                  setSearch(event.target.value);
                  setPage(1);
                }}
                placeholder="搜题干、编号..."
                aria-label="搜索题目或编号"
                className="pl-9"
              />
            </div>

            <div className="grid gap-3">
              <Select
                value={grade}
                onValueChange={(value) => {
                  setGrade(value);
                  setPage(1);
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="年级" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部年级</SelectItem>
                  {gradeOptions.map((item) => (
                    <SelectItem key={item} value={item}>
                      {item}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select
                value={tagId}
                onValueChange={(value) => {
                  setTagId(value);
                  setPage(1);
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="标签" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">全部标签</SelectItem>
                  {(tagsQuery.data ?? []).map((tag) => (
                    <SelectItem key={tag.id} value={tag.id}>
                      {tag.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-3 rounded-2xl border border-border/70 bg-background/30 p-4">
              <div className="flex items-center justify-between gap-3">
                <p className="text-sm font-medium">当前筛选</p>
                <span className="text-xs text-muted-foreground">{activeFilterCount === 0 ? '未启用' : `${activeFilterCount} 项`}</span>
              </div>
              {activeFilters.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {activeFilters.map((filter) => (
                    <Badge key={filter} variant="secondary" className="max-w-full truncate">
                      {filter}
                    </Badge>
                  ))}
                </div>
              ) : (
                <p className="text-sm leading-6 text-muted-foreground">当前展示全部题目，输入关键词或选择条件后会在这里汇总。</p>
              )}
            </div>

            <Separator />

            <div className="space-y-3">
              <p className="text-sm font-medium">难度</p>
              <div className="grid gap-2">
                {difficultyOptions.map((difficulty) => (
                  <label key={difficulty} className="flex items-center gap-2 text-sm">
                    <Checkbox checked={selectedDifficulties.includes(difficulty)} onCheckedChange={() => toggleDifficulty(difficulty)} />
                    <span>{difficultyConfig[difficulty].label}</span>
                  </label>
                ))}
              </div>
            </div>
          </div>
        </PagePanel>

        <div className="space-y-4">
          <PageToolbar className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div className="space-y-1">
              <p className="text-sm font-medium text-foreground">结果与批量操作</p>
              <p className="text-sm text-muted-foreground">
                {selectedProblems.length > 0 ? `已选中 ${selectedProblems.length} 道题，可直接批量处理。` : `当前共有 ${total} 道题，支持网格与列表两种浏览方式。`}
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              {selectedProblems.length > 0 ? (
                <>
                  <Badge variant="secondary">已选 {selectedProblems.length}</Badge>
                  <Button variant="ghost" size="sm" onClick={() => setSelectedProblems([])}>
                    <X className="mr-2 h-4 w-4" />
                    清空选择
                  </Button>
                  <ConfirmActionButton
                    variant="destructive"
                    size="sm"
                    onConfirm={handleBatchDelete}
                    pending={batchDeleteMutation.isPending}
                    title="确认批量删除"
                    description={`确定要删除选中的 ${selectedProblems.length} 道题目吗？这些题目会移入回收站。`}
                    confirmLabel="确认删除"
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    删除所选
                  </ConfirmActionButton>
                </>
              ) : null}
              <ToggleGroup type="single" value={viewMode} onValueChange={(value) => value && setViewMode(value as typeof viewMode)}>
                <ToggleGroupItem value="grid" size="sm" aria-label="网格视图">
                  <LayoutGrid className="h-4 w-4" />
                </ToggleGroupItem>
                <ToggleGroupItem value="list" size="sm" aria-label="列表视图">
                  <List className="h-4 w-4" />
                </ToggleGroupItem>
              </ToggleGroup>
            </div>
          </PageToolbar>

          {problems.length === 0 && !problemsQuery.isLoading ? (
            <PagePanel>
              <Empty className="min-h-[320px] border-none bg-transparent">
                <EmptyMedia variant="icon">
                  <Filter className="h-6 w-6" />
                </EmptyMedia>
                <EmptyHeader>
                  <EmptyTitle>当前没有匹配题目</EmptyTitle>
                  <EmptyDescription>试试清空筛选、缩短关键词，或者直接新建一道题目。</EmptyDescription>
                </EmptyHeader>
                <Button asChild>
                  <Link href="/problems/new">新建题目</Link>
                </Button>
              </Empty>
            </PagePanel>
          ) : (
            <div className={cn(viewMode === 'grid' ? 'grid items-stretch gap-4 md:grid-cols-2 2xl:grid-cols-3' : 'space-y-3')}>
              {problems.map((problem) => (
                <ProblemCard
                  key={problem.id}
                  problem={problem}
                  compact={viewMode === 'list'}
                  tags={(problem.tagIds ?? []).map((id) => tagMap.get(id)).filter((tag): tag is Tag => Boolean(tag))}
                  isSelected={selectedProblems.includes(problem.id)}
                  inBasket={hasItem(problem.id)}
                  onToggleSelected={() =>
                    setSelectedProblems((current) =>
                      current.includes(problem.id) ? current.filter((item) => item !== problem.id) : [...current, problem.id]
                    )
                  }
                  onToggleBasket={() => {
                    if (hasItem(problem.id)) {
                      removeItem(problem.id);
                      return;
                    }
                    addItem({
                      id: problem.id,
                      code: problem.code,
                      latex: problem.latex,
                      difficulty: problem.difficulty,
                    });
                  }}
                />
              ))}
            </div>
          )}

          {totalPages > 1 ? (
            <PageToolbar className="flex justify-center">
              <Pagination>
                <PaginationContent>
                  <PaginationItem>
                    <PaginationPrevious href="#" onClick={(event) => { event.preventDefault(); setPage((value) => Math.max(1, value - 1)); }} />
                  </PaginationItem>
                  <PaginationItem>
                    <span className="px-3 text-sm text-muted-foreground">
                      第 {page} / {totalPages} 页
                    </span>
                  </PaginationItem>
                  <PaginationItem>
                    <PaginationNext href="#" onClick={(event) => { event.preventDefault(); setPage((value) => Math.min(totalPages, value + 1)); }} />
                  </PaginationItem>
                </PaginationContent>
              </Pagination>
            </PageToolbar>
          ) : null}
        </div>
      </div>
    </PageShell>
  );
}

function ProblemCard({
  problem,
  tags,
  compact,
  inBasket,
  isSelected,
  onToggleSelected,
  onToggleBasket,
}: {
  problem: Problem;
  tags: Tag[];
  compact: boolean;
  inBasket: boolean;
  isSelected: boolean;
  onToggleSelected: () => void;
  onToggleBasket: () => void;
}) {
  const difficulty = difficultyConfig[problem.difficulty];

  return (
    <PagePanel className="group flex h-full flex-col overflow-hidden transition-colors hover:border-primary/40">
      <div className="flex flex-1 flex-col gap-4 p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 items-start gap-3">
            <Checkbox checked={isSelected} onCheckedChange={onToggleSelected} className="mt-1" />
            <div className="min-w-0 space-y-3">
              <p className="font-mono text-xs tracking-wide text-muted-foreground">{problem.code}</p>
              <div className={cn('relative overflow-hidden', compact ? 'min-h-[112px] max-h-[156px]' : 'min-h-[132px] max-h-[186px]')}>
                <div className="pr-1 text-sm leading-7 text-foreground">
                  <MathText latex={problem.latex} />
                </div>
                <div className="pointer-events-none absolute inset-x-0 bottom-0 h-14 bg-gradient-to-t from-card via-card/95 to-transparent" />
              </div>
            </div>
          </div>
          <Badge variant="outline" className="shrink-0" style={{ borderColor: difficulty.color, color: difficulty.color }}>
            {difficulty.label}
          </Badge>
        </div>

        <div className="flex min-h-[3.25rem] flex-wrap content-start gap-2 overflow-hidden">
          {problem.subject ? <Badge variant="secondary">{problem.subject}</Badge> : null}
          {problem.grade ? <Badge variant="secondary">{problem.grade}</Badge> : null}
          {tags.slice(0, 3).map((tag) => (
            <Badge key={tag.id} variant="outline" style={{ borderColor: tag.color, color: tag.color }}>
              {tag.name}
            </Badge>
          ))}
        </div>
      </div>

      <div className="border-t border-border/70 bg-background/20 px-5 py-4">
        <div className="mb-3 flex items-center justify-between text-xs text-muted-foreground">
          <span>{formatRelativeTime(problem.updatedAt)}</span>
          <span className="flex items-center gap-1">
            <ImageIcon className="h-3.5 w-3.5" />
            {problem.imageIds?.length ?? 0}
          </span>
        </div>

        <div className="grid gap-2 sm:grid-cols-3">
          <Button variant="outline" size="sm" asChild className="w-full">
            <Link href={`/problems/${problem.id}`}>
              <Eye className="mr-2 h-4 w-4" />
              查看
            </Link>
          </Button>
          <Button variant="outline" size="sm" asChild className="w-full">
            <Link href={`/problems/${problem.id}/edit`}>
              <Pencil className="mr-2 h-4 w-4" />
              编辑
            </Link>
          </Button>
          <Button size="sm" className="w-full" variant={inBasket ? 'outline' : 'default'} onClick={onToggleBasket}>
            {inBasket ? <Check className="mr-2 h-4 w-4" /> : <ShoppingBasket className="mr-2 h-4 w-4" />}
            {inBasket ? '已选' : '入篮'}
          </Button>
        </div>
      </div>
    </PagePanel>
  );
}
