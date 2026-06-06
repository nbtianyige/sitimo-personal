'use client';

import Link from 'next/link';
import { FileText, BookText, File, ArrowRight, UploadCloud } from 'lucide-react';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Button } from '@/components/ui/button';

const formats = [
  {
    id: 'tex',
    label: 'LaTeX 源码',
    extension: '.tex',
    icon: FileText,
    description: '标准 LaTeX 题目源文件，支持 \begin{problem} 标记拆题或题号风格。',
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
                    <span className="ml-2 font-mono text-xs text-muted-foreground">
                      {format.extension}
                    </span>
                  </h3>
                  <p className="mt-1 text-sm leading-6 text-muted-foreground">
                    {format.description}
                  </p>
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
    </PageShell>
  );
}
