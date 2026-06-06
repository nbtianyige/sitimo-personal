'use client';

import { Fragment } from 'react';
import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { HelpCircle, LoaderCircle, Menu, Moon, ShoppingBasket, Sun, X } from 'lucide-react';
import { useTheme } from 'next-themes';
import { Button } from '@/components/ui/button';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import { Progress } from '@/components/ui/progress';
import { useDeleteExport } from '@/lib/hooks/use-exports';
import { useBasketStore, useExportStore, useSidebarStore } from '@/lib/store';

const pathLabels: Record<string, string> = {
  '/': '首页',
  '/problems': '题库',
  '/problems/new': '新建题目',
  '/images': '图库',
  '/tags': '标签管理',
  '/search': '高级搜索',
  '/papers': '试卷列表',
  '/import': '导入中心',
  '/trash': '回收站',
  '/settings': '设置',
};

function getBreadcrumbs(pathname: string) {
  const segments = pathname.split('/').filter(Boolean);
  const breadcrumbs: { label: string; href: string; isLast: boolean }[] = [];

  // 首页
  breadcrumbs.push({ label: '首页', href: '/', isLast: segments.length === 0 });

  let currentPath = '';
  segments.forEach((segment, index) => {
    currentPath += `/${segment}`;
    const isLast = index === segments.length - 1;

    // 处理动态路由
    if (segment.match(/^[a-zA-Z0-9-]+$/) && !pathLabels[currentPath]) {
      // 可能是ID，跳过或显示代码
      if (segments[index - 1] === 'problems') {
        breadcrumbs.push({
          label: segment.toUpperCase(),
          href: currentPath,
          isLast,
        });
      } else if (segments[index - 1] === 'papers') {
        breadcrumbs.push({
          label: '试卷详情',
          href: currentPath,
          isLast,
        });
      }
    } else {
      breadcrumbs.push({
        label: pathLabels[currentPath] || segment,
        href: currentPath,
        isLast,
      });
    }
  });

  return breadcrumbs;
}

interface AppHeaderProps {
  onOpenBasket: () => void;
}

export function AppHeader({ onOpenBasket }: AppHeaderProps) {
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();
  const { toggle } = useSidebarStore();
  const basketItems = useBasketStore((state) => state.items);
  const activeExportTask = useExportStore((state) => state.activeTask);
  const deleteExportMutation = useDeleteExport();

  const breadcrumbs = getBreadcrumbs(pathname);

  return (
    <header className="sticky top-0 z-30 bg-background/95 backdrop-blur">
      <div className="relative box-border flex h-14 items-center justify-between border-b border-border px-4">
        <div className="flex items-center gap-3">
          <Button
            variant="ghost"
            size="icon"
            onClick={toggle}
            className="h-8 w-8"
          >
            <Menu className="h-4 w-4" />
            <span className="sr-only">切换侧边栏</span>
          </Button>

          <Breadcrumb>
            <BreadcrumbList>
              {breadcrumbs.map((item, index) => (
                <Fragment key={item.href}>
                  <BreadcrumbItem>
                    {item.isLast ? (
                      <BreadcrumbPage>{item.label}</BreadcrumbPage>
                    ) : (
                      <BreadcrumbLink asChild>
                        <Link href={item.href}>{item.label}</Link>
                      </BreadcrumbLink>
                    )}
                  </BreadcrumbItem>
                  {index < breadcrumbs.length - 1 && <BreadcrumbSeparator />}
                </Fragment>
              ))}
            </BreadcrumbList>
          </Breadcrumb>
        </div>

        <div className="flex items-center gap-2">
          {activeExportTask ? (
            <div className="hidden min-w-[280px] items-center gap-3 rounded-full border bg-muted/60 px-3 py-2 xl:flex">
              <LoaderCircle className="h-4 w-4 animate-spin text-primary" />
              <div className="min-w-0 flex-1">
                <div className="flex items-center justify-between gap-3 text-xs">
                  <p className="truncate font-medium">
                    正在生成《{activeExportTask.paperTitle}》
                  </p>
                  <span className="shrink-0 text-muted-foreground">
                    {activeExportTask.progress}%
                  </span>
                </div>
                <Progress value={activeExportTask.progress} className="mt-1 h-1.5" />
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 shrink-0"
                onClick={() => deleteExportMutation.mutate(activeExportTask.id)}
                disabled={activeExportTask.status !== 'pending' || deleteExportMutation.isPending}
              >
                <X className="h-4 w-4" />
                <span className="sr-only">取消导出</span>
              </Button>
            </div>
          ) : null}

          <Button
            variant="ghost"
            size="icon"
            onClick={onOpenBasket}
            className="relative h-8 w-8"
          >
            <ShoppingBasket className="h-4 w-4" />
            {basketItems.length > 0 && (
              <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-accent text-[10px] font-semibold text-accent-foreground">
                {basketItems.length}
              </span>
            )}
            <span className="sr-only">题目篮子</span>
          </Button>

          <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
            className="h-8 w-8"
          >
            <Sun className="h-4 w-4 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
            <Moon className="absolute h-4 w-4 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
            <span className="sr-only">切换主题</span>
          </Button>

          <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
            <Link href="/help">
              <HelpCircle className="h-4 w-4" />
              <span className="sr-only">帮助</span>
            </Link>
          </Button>
        </div>

        {activeExportTask ? (
          <div className="absolute inset-x-0 bottom-0">
            <Progress value={activeExportTask.progress} className="h-0.5 rounded-none" />
          </div>
        ) : null}
      </div>
    </header>
  );
}
