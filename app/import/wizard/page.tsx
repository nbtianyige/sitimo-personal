'use client';

import { useCallback, useMemo, useRef, useState } from 'react';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm, useWatch } from 'react-hook-form';
import { CheckCircle, FileText, BookText, File, FileUp, MoveRight, UploadCloud, AlertTriangle } from 'lucide-react';
import { parseTexFile } from '@/lib/latex-file-parser';
import { LatexCodeEditor } from '@/components/latex-code-editor';
import { MathText } from '@/components/math-text';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Textarea } from '@/components/ui/textarea';
import { STANDARD_GRADES } from '@/lib/constants';
import { useCommitBatchImport, usePreviewBatchImport } from '@/lib/hooks/use-problems';
import { problemImportSchema, type ProblemImportFormValues } from '@/lib/schemas/import';
import { difficultyConfig, type Difficulty } from '@/lib/types';

type FileType = 'tex' | 'md' | 'pdf';

interface FileInfo {
  name: string;
  type: FileType;
  problemCount?: number;
  warnings: string[];
}

const defaultImportValues: ProblemImportFormValues = {
  latexSource: '',
  separatorStart: '\\begin{problem}',
  separatorEnd: '\\end{problem}',
  subject: '数学',
  grade: '高三',
  source: '',
  difficulty: 'medium',
  tagNames: '',
};

function parseTagNames(value: string | undefined) {
  return (value ?? '')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

function simpleMdToTex(content: string): string {
  // Convert $$...$$ block math to \[...\]
  let result = content.replace(/\$\$([\s\S]*?)\$\$/g, '\\[$1\\]');
  // Convert inline $...$ to \(...\)
  // Use a helper to avoid double-converting already converted block math
  result = result.replace(/(?<!\\)\$(.+?)(?<!\\)\$/g, '\\($1\\)');
  return result;
}

export default function ImportWizardPage() {
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [isDragging, setIsDragging] = useState(false);
  const [fileInfo, setFileInfo] = useState<FileInfo | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const previewMutation = usePreviewBatchImport();
  const commitMutation = useCommitBatchImport();
  const form = useForm<ProblemImportFormValues>({
    resolver: zodResolver(problemImportSchema),
    defaultValues: defaultImportValues,
  });

  const latexSource = useWatch({ control: form.control, name: 'latexSource' });
  const subject = useWatch({ control: form.control, name: 'subject' });
  const grade = useWatch({ control: form.control, name: 'grade' });
  const source = useWatch({ control: form.control, name: 'source' });
  const difficulty = useWatch({ control: form.control, name: 'difficulty' });
  const tagNames = useWatch({ control: form.control, name: 'tagNames' });

  const parsedProblems = useMemo(() => previewMutation.data?.parsed ?? [], [previewMutation.data?.parsed]);
  const normalizedProblems = useMemo(
    () =>
      parsedProblems.map((problem) => ({
        ...problem,
        subject: problem.subject ?? subject,
        grade: problem.grade ?? grade,
        source: problem.source ?? source,
        difficulty: problem.difficulty ?? difficulty,
      })),
    [difficulty, grade, parsedProblems, source, subject]
  );

  const successCount = normalizedProblems.filter((problem) => problem.status === 'success').length;
  const failureCount = normalizedProblems.length - successCount;
  const previewReady = normalizedProblems.length > 0;
  const canEnterStep2 = previewMutation.isSuccess;
  const canEnterStep3 = previewMutation.isSuccess;
  const canImportAll = previewReady && failureCount === 0;

  const [rawFile, setRawFile] = useState<File | null>(null);

  const runPreview = form.handleSubmit(async (values) => {
    if (fileInfo?.type === 'pdf' && rawFile) {
      const formData = new FormData();
      formData.append('files', rawFile);
      formData.append(
        'defaults',
        JSON.stringify({
          subject: values.subject,
          grade: values.grade,
          source: values.source,
          difficulty: values.difficulty,
          tagNames: parseTagNames(values.tagNames),
        })
      );
      await previewMutation.mutateAsync(formData);
    } else {
      await previewMutation.mutateAsync({
        latex: values.latexSource,
        separatorStart: values.separatorStart,
        separatorEnd: values.separatorEnd,
        defaults: {
          subject: values.subject,
          grade: values.grade,
          source: values.source,
          difficulty: values.difficulty,
          tagNames: parseTagNames(values.tagNames),
        },
      });
    }
    setStep(2);
  });

  const importAll = async () => {
    if (!canImportAll) return;
    const sharedTags = parseTagNames(tagNames);
    await commitMutation.mutateAsync(
      normalizedProblems.map((problem) => ({
        ...problem,
        tagNames: Array.from(new Set([...(problem.tagNames ?? []), ...sharedTags])),
      }))
    );
  };

  const goToStep = (value: 1 | 2 | 3) => {
    if (value === 1) {
      setStep(1);
      return;
    }
    if (value === 2 && canEnterStep2) {
      setStep(2);
      return;
    }
    if (value === 3 && canEnterStep3) {
      setStep(3);
    }
  };

  const handleFile = useCallback(
    (file: File) => {
      setUploadError(null);
      const name = file.name.toLowerCase();
      let type: FileType | null = null;
      if (name.endsWith('.tex')) type = 'tex';
      else if (name.endsWith('.md')) type = 'md';
      else if (name.endsWith('.pdf')) type = 'pdf';

      if (!type) {
        setUploadError('仅支持 .tex、.md 和 .pdf 文件');
        return;
      }

      if (type === 'pdf') {
        setFileInfo({ name: file.name, type: 'pdf', warnings: [] });
        setRawFile(file);
        form.setValue('latexSource', '');
        return;
      }

      setRawFile(null);
      const reader = new FileReader();
      reader.onload = (e) => {
        const rawContent = e.target?.result as string;
        if (type === 'tex') {
          const result = parseTexFile(rawContent);
          form.setValue('latexSource', result.latex);
          form.setValue('separatorStart', '\\begin{problem}');
          form.setValue('separatorEnd', '\\end{problem}');
          if (result.suggestedSource) form.setValue('source', result.suggestedSource);
          setFileInfo({
            name: file.name,
            type: 'tex',
            problemCount: result.problemCount,
            warnings: result.warnings,
          });
        } else if (type === 'md') {
          const converted = simpleMdToTex(rawContent);
          // Try to infer source from first # heading
          const sourceMatch = rawContent.match(/^#\s+(.+)$/m);
          if (sourceMatch) {
            form.setValue('source', sourceMatch[1].trim());
          }
          form.setValue('latexSource', converted);
          // For markdown, we don't have structured markers by default
          // User can adjust separator as needed
          setFileInfo({
            name: file.name,
            type: 'md',
            warnings: ['Markdown 已转换为 LaTeX 数学公式，但可能需要手动添加分隔符或调整结构。'],
          });
        }
      };
      reader.readAsText(file, 'utf-8');
    },
    [form]
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragging(false);
      const file = e.dataTransfer.files[0];
      if (file) handleFile(file);
    },
    [handleFile]
  );

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  return (
    <Form {...form}>
      <PageShell wide>
        <PageHeader
          eyebrow="批量导入"
          title="导入向导"
          description="上传 LaTeX、Markdown 或 PDF 试卷，按步骤完成解析与入库。PDF 将由后端自动转换为结构化文本。"
          badges={
            <>
              <Badge variant="secondary">成功 {successCount}</Badge>
              <Badge variant={failureCount > 0 ? 'destructive' : 'secondary'}>失败 {failureCount}</Badge>
            </>
          }
        />

        {/* Step navigation */}
        <div className="grid gap-3 md:grid-cols-3">
          {[1, 2, 3].map((value) => {
            const available = value === 1 || (value === 2 ? canEnterStep2 : canEnterStep3);
            const completed = value < step && available;

            return (
              <button
                key={value}
                type="button"
                onClick={() => goToStep(value as 1 | 2 | 3)}
                className={`rounded-2xl border p-4 text-left transition-colors ${
                  step === value
                    ? 'border-primary bg-primary/5'
                    : completed
                      ? 'border-border/70 bg-card'
                      : 'border-dashed border-border/70 bg-muted/20 text-muted-foreground'
                }`}
              >
                <p className="text-xs uppercase tracking-[0.2em]">{completed ? 'Done' : `Step ${value}`}</p>
                <p className="mt-1 font-medium">{value === 1 ? '上传与配置' : value === 2 ? '解析预览' : '确认导入'}</p>
                <p className="mt-2 text-sm text-muted-foreground">
                  {value === 1
                    ? '上传文件并配置解析参数'
                    : value === 2
                      ? '检查每道题的解析结果'
                      : available
                        ? '确认无误后批量入库'
                        : '完成上一步后解锁'}
                </p>
              </button>
            );
          })}
        </div>

        {/* Step 1: Upload */}
        {step === 1 ? (
          <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
            <PagePanel>
              <div className="space-y-4 p-6">
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".tex,.md,.pdf"
                  className="hidden"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) handleFile(file);
                    e.target.value = '';
                  }}
                />
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  onDrop={handleDrop}
                  onDragOver={handleDragOver}
                  onDragLeave={handleDragLeave}
                  className={`flex h-48 w-full flex-col items-center justify-center rounded-2xl border-2 border-dashed text-center transition-colors ${
                    isDragging
                      ? 'border-primary bg-primary/10'
                      : fileInfo
                        ? 'border-green-500/60 bg-green-500/5'
                        : 'border-border hover:border-primary/60 hover:bg-muted/30'
                  }`}
                >
                  {fileInfo ? (
                    <>
                      <CheckCircle className="mb-3 h-10 w-10 text-green-500" />
                      <p className="font-medium text-foreground">{fileInfo.name}</p>
                      {fileInfo.type === 'pdf' ? (
                        <>
                          <File className="mt-2 h-5 w-5 text-muted-foreground" />
                          <p className="mt-1 text-sm text-green-600">PDF 文件已接收，点击解析预览进行转换</p>
                        </>
                      ) : (
                        <p className="mt-1 text-sm text-muted-foreground">
                          {fileInfo.type === 'tex' && fileInfo.problemCount !== undefined
                            ? `检测到 ${fileInfo.problemCount} 道题目`
                            : '文件已加载，右侧可调整解析参数'}
                        </p>
                      )}
                      {fileInfo.warnings.length > 0 && fileInfo.type !== 'pdf' && (
                        <p className="mt-1 text-xs text-amber-600">{fileInfo.warnings.join('；')}</p>
                      )}
                      <p className="mt-2 text-xs text-muted-foreground">点击重新上传</p>
                    </>
                  ) : (
                    <>
                      <UploadCloud className={`mb-3 h-10 w-10 ${isDragging ? 'text-primary' : 'text-muted-foreground'}`} />
                      <p className="font-medium">{isDragging ? '松开以上传' : '拖拽文件到此，或点击选择'}</p>
                      <p className="mt-1 text-xs text-muted-foreground">支持 .tex 源码、.md 文本与 .pdf 试卷</p>
                      <div className="mt-3 flex items-center justify-center gap-4 text-xs text-muted-foreground">
                        <span className="flex items-center gap-1"><FileText className="h-3.5 w-3.5" />.tex</span>
                        <span className="flex items-center gap-1"><BookText className="h-3.5 w-3.5" />.md</span>
                        <span className="flex items-center gap-1"><File className="h-3.5 w-3.5" />.pdf</span>
                      </div>
                    </>
                  )}
                </button>

                {uploadError && (
                  <div className="flex items-center gap-2 rounded-lg bg-amber-50 px-4 py-3 text-sm text-amber-700">
                    <AlertTriangle className="h-4 w-4 shrink-0" />
                    {uploadError}
                  </div>
                )}

                <FormField
                  control={form.control}
                  name="latexSource"
                  render={({ field }) => (
                    <FormItem>
                      <FormControl>
                        <Textarea {...field} rows={14} className="font-mono" placeholder="或将源码直接粘贴到此处..." />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </PagePanel>

            <PagePanel>
              <div className="space-y-4 p-5">
                <h2 className="text-base font-semibold">解析参数</h2>
                <FormField
                  control={form.control}
                  name="separatorStart"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>开始标记</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="separatorEnd"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>结束标记</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="subject"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>学科</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="grade"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>年级</FormLabel>
                      <Select value={field.value} onValueChange={field.onChange}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {STANDARD_GRADES.map((item) => (
                            <SelectItem key={item} value={item}>
                              {item}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="source"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>来源</FormLabel>
                      <FormControl>
                        <Input {...field} value={field.value ?? ''} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="difficulty"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>默认难度</FormLabel>
                      <Select value={field.value} onValueChange={field.onChange}>
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {(Object.keys(difficultyConfig) as Difficulty[]).map((item) => (
                            <SelectItem key={item} value={item}>
                              {difficultyConfig[item].label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="tagNames"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>标签</FormLabel>
                      <FormControl>
                        <Input {...field} value={field.value ?? ''} placeholder="多个标签用逗号分隔" />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button className="w-full" type="button" onClick={runPreview} disabled={previewMutation.isPending}>
                  解析预览
                  <MoveRight className="ml-2 h-4 w-4" />
                </Button>
              </div>
            </PagePanel>
          </div>
        ) : null}

        {/* Step 2: Preview */}
        {step === 2 ? (
          <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
            <PagePanel>
              <div className="border-b border-border/70 px-6 py-4">
                <h2 className="text-base font-semibold">原始源码</h2>
              </div>
              <div className="p-6">
                <LatexCodeEditor value={latexSource} readOnly minHeight={560} />
              </div>
            </PagePanel>

            <PagePanel>
              <div className="flex items-center justify-between border-b border-border/70 px-6 py-4">
                <div>
                  <h2 className="text-base font-semibold">解析结果</h2>
                  <p className="mt-1 text-sm text-muted-foreground">成功 {successCount}，失败 {failureCount}</p>
                </div>
                <Button variant="outline" type="button" onClick={() => setStep(3)} disabled={!previewReady}>
                  下一步
                </Button>
              </div>
              <div className="space-y-4 p-6">
                {normalizedProblems.length === 0 ? (
                  <div className="py-10 text-center text-sm text-muted-foreground">
                    未解析到题目，请返回上一步检查源码与分隔符配置。
                  </div>
                ) : (
                  normalizedProblems.map((problem) => (
                    <div key={problem.id} className="space-y-3 rounded-2xl border border-border/70 p-4">
                      <div className="flex items-center justify-between gap-3">
                        <div>
                          <p className="font-medium">{problem.title}</p>
                          <p className="text-xs text-muted-foreground">{problem.status === 'success' ? '解析成功' : problem.error}</p>
                        </div>
                        <Badge variant={problem.status === 'success' ? 'outline' : 'destructive'}>{problem.status === 'success' ? '成功' : '错误'}</Badge>
                      </div>
                      <div className="rounded-xl bg-muted/40 p-4">
                        <MathText latex={problem.latex} />
                      </div>
                      {problem.warnings?.length ? <div className="text-xs text-amber-600">{problem.warnings.join('；')}</div> : null}
                    </div>
                  ))
                )}
              </div>
            </PagePanel>
          </div>
        ) : null}

        {/* Step 3: Confirm */}
        {step === 3 ? (
          <PagePanel>
            <div className="flex flex-wrap items-center justify-between gap-4 border-b border-border/70 px-6 py-4">
              <div>
                <h2 className="text-base font-semibold">导入确认</h2>
                <p className="mt-1 text-sm text-muted-foreground">
                  {canImportAll ? '所有题目解析正常，可以直接导入。' : '当前仍有解析错误或没有可导入结果，导入按钮已禁用。'}
                </p>
              </div>
              <Button type="button" onClick={importAll} disabled={!canImportAll || commitMutation.isPending}>
                <FileUp className="mr-2 h-4 w-4" />
                导入全部
              </Button>
            </div>
            <div className="overflow-hidden rounded-b-3xl">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>状态</TableHead>
                    <TableHead>预览</TableHead>
                    <TableHead>学科</TableHead>
                    <TableHead>年级</TableHead>
                    <TableHead>来源</TableHead>
                    <TableHead>难度</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {normalizedProblems.map((problem) => (
                    <TableRow key={problem.id}>
                      <TableCell>
                        <Badge variant={problem.status === 'success' ? 'outline' : 'destructive'}>{problem.status === 'success' ? '成功' : '错误'}</Badge>
                      </TableCell>
                      <TableCell className="max-w-[420px]">
                        <MathText latex={problem.latex} className="line-clamp-2" />
                      </TableCell>
                      <TableCell>{problem.subject}</TableCell>
                      <TableCell>{problem.grade}</TableCell>
                      <TableCell>{problem.source}</TableCell>
                      <TableCell>{difficultyConfig[problem.difficulty].label}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
              {normalizedProblems.length === 0 && (
                <div className="px-6 py-8 text-center text-sm text-muted-foreground">没有可导入的题目。</div>
              )}
            </div>
          </PagePanel>
        ) : null}
      </PageShell>
    </Form>
  );
}
