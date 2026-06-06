'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Home,
  FileText,
  Image,
  Tag,
  Search,
  FileStack,
  Trash2,
  Settings,
  ChevronLeft,
  ChevronRight,
  UploadCloud,
} from 'lucide-react';
import { BrandMark } from '@/components/brand-mark';
import { cn } from '@/lib/utils';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { useSidebarStore } from '@/lib/store';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

const navItems = [
  { label: '首页', icon: Home, href: '/' },
  { label: '题库', icon: FileText, href: '/problems' },
  { label: '图库', icon: Image, href: '/images' },
  { label: '标签', icon: Tag, href: '/tags' },
  { label: '搜索', icon: Search, href: '/search' },
  { label: '试卷', icon: FileStack, href: '/papers' },
  { label: '导入工具', icon: UploadCloud, href: '/import' },
  { label: '回收站', icon: Trash2, href: '/trash' },
  { label: '设置', icon: Settings, href: '/settings' },
];

export function AppSidebar() {
  const pathname = usePathname();
  const { isCollapsed, toggle } = useSidebarStore();

  return (
    <TooltipProvider delayDuration={0}>
      <aside
        className={cn(
          'fixed left-0 top-0 z-40 flex h-screen flex-col border-r border-sidebar-border bg-sidebar transition-all duration-300',
          isCollapsed ? 'w-16' : 'w-60'
        )}
      >
        {/* Logo */}
        <div
          className={cn(
            'box-border flex h-14 shrink-0 items-center border-b border-sidebar-border px-4',
            isCollapsed && 'justify-center px-2'
          )}
        >
          <Link href="/" className="flex items-center gap-2">
            <BrandMark />
            {!isCollapsed && (
              <span className="text-lg font-semibold text-sidebar-foreground">
                Sitimo
              </span>
            )}
          </Link>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto py-4">
          <ul className="space-y-1 px-2">
            {navItems.map((item) => {
              const isActive =
                pathname === item.href ||
                (item.href !== '/' && pathname.startsWith(item.href));
              const Icon = item.icon;

              const linkContent = (
                <Link
                  href={item.href}
                  className={cn(
                    'group relative flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-sidebar-accent text-sidebar-primary'
                      : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
                    isCollapsed && 'justify-center px-2'
                  )}
                >
                  {/* Active indicator bar */}
                  {isActive && (
                    <div className="absolute left-0 top-1/2 h-6 w-[3px] -translate-y-1/2 rounded-r-full bg-primary" />
                  )}
                  <Icon className="h-5 w-5 shrink-0" />
                  {!isCollapsed && <span>{item.label}</span>}
                </Link>
              );

              return (
                <li key={item.href}>
                  {isCollapsed ? (
                    <Tooltip>
                      <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
                      <TooltipContent side="right" sideOffset={10}>
                        {item.label}
                      </TooltipContent>
                    </Tooltip>
                  ) : (
                    linkContent
                  )}
                </li>
              );
            })}
          </ul>
        </nav>

        {/* Bottom section */}
        <div className="border-t border-sidebar-border p-2">
          {/* User */}
          <div
            className={cn(
              'mb-2 flex items-center gap-3 rounded-md px-3 py-2',
              isCollapsed && 'justify-center px-2'
            )}
          >
            <Avatar className="h-8 w-8">
              <AvatarFallback className="bg-primary text-primary-foreground text-sm">
                教
              </AvatarFallback>
            </Avatar>
            {!isCollapsed && (
              <span className="text-sm font-medium text-sidebar-foreground">
                教师
              </span>
            )}
          </div>

          {/* Collapse button */}
          <Button
            variant="ghost"
            size="sm"
            onClick={toggle}
            className={cn(
              'w-full justify-center text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
              !isCollapsed && 'justify-start'
            )}
          >
            {isCollapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <>
                <ChevronLeft className="mr-2 h-4 w-4" />
                <span>收起侧边栏</span>
              </>
            )}
          </Button>
        </div>
      </aside>
    </TooltipProvider>
  );
}
