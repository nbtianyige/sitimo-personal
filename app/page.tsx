'use client';

import Link from 'next/link';
import { Download, FileStack, FileText, Image as ImageIcon, ImagePlus, Plus, Search, Tag } from 'lucide-react';
import { MathText } from '@/components/math-text';
import { OnboardingTour } from '@/components/onboarding-tour';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty';
import { Skeleton } from '@/components/ui/skeleton';
import { formatRelativeTime } from '@/lib/format';
import { cn } from '@/lib/utils';
import { useMetaStats, useRecentExports, useRecentProblems } from '@/lib/hooks/use-meta';
import { difficultyConfig, type ExportJob, type Problem } from '@/lib/types';

const quickActions = [
  { href: '/problems/import', label: '上传题目', description: '批量解析 LaTeX 并入库', icon: FileText },
  { href: '/images', label: '上传图像', description: '管理几何图、函数图和素材', icon: ImagePlus },
  { href: '/papers/new', label: '新建试卷', description: '从题目篮子快速组卷', icon: FileStack },
  { href: '/search', label: '高级搜索', description: '组合条件定位题目', icon: Search },
];

const statusConfig: Record<ExportJob['status'], { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  pending: { label: '排队中', variant: 'secondary' },
  processing: { label: '生成中', variant: 'outline' },
  done: { label: '完成', variant: 'default' },
  failed: { label: '失败', variant: 'destructive' },
};

export default function DashboardPage() {
  const statsQuery = useMetaStats();
  const recentProblemsQuery = useRecentProblems(5);
  const recentExportsQuery = useRecentExports(5);

  const stats = statsQuery.data;
  const recentProblems = recentProblemsQuery.data ?? [];
  const recentExports = recentExportsQuery.data ?? [];

  return (
    <PageShell>
      <PageHeader
        eyebrow="总览"
        title="你好，管理员"
        description={
          stats ? (
            <>
              当前题库共有 <strong className="text-foreground">{stats.problemCount}</strong> 道题、
              <strong className="text-foreground"> {stats.imageCount}</strong> 张图像，最近 7 天新增
              <strong className="text-foreground"> {stats.recentProblemGain}</strong> 道题。
            </>
          ) : (
            <Skeleton className="h-5 w-[420px]" />
          )
        }
        badges={
          stats ? (
            <>
              <Badge variant="secondary">标签 {stats.tagCount}</Badge>
              <Badge variant="secondary">本月导出 {stats.exportCount}</Badge>
            </>
          ) : null
        }
        actions={
          <>
            <Button asChild>
              <Link href="/problems/new">
                <Plus className="mr-2 h-4 w-4" />
                新建题目
              </Link>
            </Button>
            <Button variant="outline" asChild>
              <Link href="/images">
                <ImagePlus className="mr-2 h-4 w-4" />
                上传图像
              </Link>
            </Button>
          </>
        }
        className="border-primary/20 bg-[linear-gradient(135deg,hsl(var(--primary)/0.12),transparent_60%),linear-gradient(180deg,hsl(var(--card)),hsl(var(--card)))]"
      />

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard icon={FileText} label="题目总数" value={String(stats?.problemCount ?? '—')} iconClassName="bg-primary/12 text-primary" />
        <StatCard icon={ImageIcon} label="图像总数" value={String(stats?.imageCount ?? '—')} iconClassName="bg-sky-500/12 text-sky-500" />
        <StatCard icon={Tag} label="标签总数" value={String(stats?.tagCount ?? '—')} iconClassName="bg-accent/14 text-accent" />
        <StatCard icon={Download} label="本月导出" value={String(stats?.exportCount ?? '—')} iconClassName="bg-violet-500/12 text-violet-500" />
      </div>

      <div className="grid gap-4 xl:grid-cols-2">
        <PagePanel className="overflow-hidden">
          <SectionHeader title="最近编辑的题目" href="/problems" />
          <div className="divide-y">
            {recentProblems.length > 0 ? recentProblems.map((problem) => <RecentProblemRow key={problem.id} problem={problem} />) : <EmptyList label="还没有最近编辑记录" description="开始录入题目后，这里会展示最近修改过的内容。" />}
          </div>
        </PagePanel>

        <PagePanel className="overflow-hidden">
          <SectionHeader title="最近导出" href="/import" />
          <div className="divide-y">
            {recentExports.length > 0 ? (
              recentExports.map((job) => {
                const status = statusConfig[job.status];
                return (
                  <div key={job.id} className="flex items-center gap-4 px-5 py-4">
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">{job.paperTitle}</p>
                      <p className="mt-1 text-xs text-muted-foreground">{formatRelativeTime(job.createdAt)}</p>
                    </div>
                    <Badge variant="outline" className="shrink-0">
                      {job.format.toUpperCase()}
                    </Badge>
                    <Badge variant={status.variant} className="shrink-0">
                      {status.label}
                    </Badge>
                  </div>
                );
              })
            ) : (
              <EmptyList label="还没有导出记录" description="完成一次试卷导出后，这里会展示最近的任务状态。" />
            )}
          </div>
        </PagePanel>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {quickActions.map((action) => {
          const Icon = action.icon;
          return (
            <Link key={action.href} href={action.href} className="group">
              <PagePanel className="h-full transition-all hover:-translate-y-0.5 hover:border-primary/50">
                <div className="flex h-full flex-col justify-between gap-4 p-5">
                  <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground transition-colors group-hover:bg-primary/10 group-hover:text-primary">
                    <Icon className="h-6 w-6" />
                  </div>
                  <div className="space-y-1">
                    <p className="font-medium">{action.label}</p>
                    <p className="text-sm leading-6 text-muted-foreground">{action.description}</p>
                  </div>
                </div>
              </PagePanel>
            </Link>
          );
        })}
      </div>

      <OnboardingTour open={false} onOpenChange={() => {}} onFinish={() => {}} />
    </PageShell>
  );
}

function SectionHeader({ title, href }: { title: string; href: string }) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-border/70 px-5 py-4">
      <div>
        <h2 className="text-base font-semibold">{title}</h2>
      </div>
      <Link href={href} className="text-sm text-muted-foreground transition-colors hover:text-primary">
        查看全部
      </Link>
    </div>
  );
}

function StatCard({
  icon: Icon,
  label,
  value,
  iconClassName,
}: {
  icon: typeof FileText;
  label: string;
  value: string;
  iconClassName: string;
}) {
  return (
    <PagePanel>
      <div className="flex items-center gap-4 p-5">
        <div className={cn("flex h-12 w-12 items-center justify-center rounded-2xl", iconClassName)}>
          <Icon className="h-6 w-6" />
        </div>
        <div>
          <p className="text-3xl font-bold">{value}</p>
          <p className="text-sm text-muted-foreground">{label}</p>
        </div>
      </div>
    </PagePanel>
  );
}

function RecentProblemRow({ problem }: { problem: Problem }) {
  const config = difficultyConfig[problem.difficulty];

  return (
    <Link href={`/problems/${problem.id}`} className="flex items-center gap-4 px-5 py-4 transition-transform hover:-translate-x-0.5">
      <div className="min-w-0 flex-1">
        <p className="font-mono text-xs text-muted-foreground">{problem.code}</p>
        <div className="mt-1 line-clamp-2 text-sm leading-7">
          <MathText latex={problem.latex} />
        </div>
      </div>
      <Badge variant="outline" style={{ borderColor: config.color, color: config.color }}>
        {config.label}
      </Badge>
      <span className="text-xs text-muted-foreground">{formatRelativeTime(problem.updatedAt)}</span>
    </Link>
  );
}

function EmptyList({ label, description }: { label: string; description: string }) {
  return (
    <div className="px-5 py-8">
      <Empty className="border-none bg-transparent">
        <EmptyHeader className="items-start text-left">
          <EmptyTitle className="text-base">{label}</EmptyTitle>
          <EmptyDescription className="text-left">{description}</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </div>
  );
}
