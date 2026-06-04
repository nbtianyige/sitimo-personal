'use client';

import { use, useEffect, useMemo } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { ArrowLeft, Save } from 'lucide-react';
import { toast } from 'sonner';
import { LatexCodeEditor } from '@/components/latex-code-editor';
import { MathText } from '@/components/math-text';
import { PageHeader, PagePanel, PageShell } from '@/components/page-shell';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Slider } from '@/components/ui/slider';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { buildGradeOptions, STANDARD_GRADES } from '@/lib/constants';
import { AppError } from '@/lib/api/client';
import { applyValidationErrors } from '@/lib/forms';
import { useImages } from '@/lib/hooks/use-images';
import { useCreateProblem, useProblem, useUpdateProblem } from '@/lib/hooks/use-problems';
import { useTags } from '@/lib/hooks/use-tags';
import { problemSchema, type ProblemFormValues } from '@/lib/schemas/problem';
import { difficultyConfig, problemTypeConfig, type Difficulty } from '@/lib/types';

const UNSPECIFIED_GRADE = '__unspecified__';

export default function ProblemEditPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const isNew = id === 'new';
  const router = useRouter();
  const problemQuery = useProblem(isNew ? undefined : id);
  const tagsQuery = useTags();
  const imagesQuery = useImages({ page: 1, pageSize: 50 });
  const createMutation = useCreateProblem();
  const updateMutation = useUpdateProblem(id);

  const form = useForm<ProblemFormValues>({
    resolver: zodResolver(problemSchema),
    defaultValues: {
      latex: '',
      answerLatex: '',
      solutionLatex: '',
      type: 'solve',
      difficulty: 'medium',
      subjectiveScore: 5,
      grade: '',
      source: '',
      notes: '',
      tagIds: [],
      imageIds: [],
    },
  });

  useEffect(() => {
    if (!problemQuery.data) return;
    const problem = problemQuery.data;
    form.reset({
      latex: problem.latex,
      answerLatex: problem.answerLatex ?? '',
      solutionLatex: problem.solutionLatex ?? '',
      type: problem.type,
      difficulty: problem.difficulty,
      subjectiveScore: problem.subjectiveScore ?? 5,
      grade: problem.grade ?? '',
      source: problem.source ?? '',
      notes: problem.notes ?? '',
      tagIds: problem.tagIds,
      imageIds: problem.imageIds,
    });
  }, [problemQuery.data, form]);

  const onSubmit = async (data: ProblemFormValues) => {
    const payload = {
      latex: data.latex,
      answerLatex: data.answerLatex || undefined,
      solutionLatex: data.solutionLatex || undefined,
      type: data.type,
      difficulty: data.difficulty,
      subjectiveScore: data.subjectiveScore,
      grade: data.grade || undefined,
      source: data.source || undefined,
      notes: data.notes || undefined,
      tagIds: data.tagIds,
      imageIds: data.imageIds,
    };

    try {
      const result = isNew
        ? await createMutation.mutateAsync(payload)
        : await updateMutation.mutateAsync(payload);
      if (result.warnings.length > 0) {
        toast.warning(result.warnings.join('；'));
      }
      router.push(`/problems/${result.problem.id}`);
    } catch (error) {
      applyValidationErrors(form, error);
      if (error instanceof AppError && error.code !== 'validation_failed') {
        toast.error(error.message);
      }
    }
  };

  const latex = form.watch('latex');
  const answerLatex = form.watch('answerLatex');
  const solutionLatex = form.watch('solutionLatex');
  const tagIds = form.watch('tagIds');
  const imageIds = form.watch('imageIds');
  const subjectiveScore = form.watch('subjectiveScore') ?? 5;
  const difficulty = form.watch('difficulty');
  const grade = form.watch('grade');
  const gradeOptions = useMemo(
    () => buildGradeOptions(problemQuery.data?.grade ? [problemQuery.data.grade] : STANDARD_GRADES),
    [problemQuery.data?.grade]
  );

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <PageShell wide>
          <PageHeader
            eyebrow="题目编辑器"
            title={isNew ? '新建题目' : '编辑题目'}
            description="统一在一个页面完成题干、答案、解析、标签和关联图像的录入与预览。"
            badges={
              <>
                <Badge variant="secondary">{problemTypeConfig[form.watch('type')]}</Badge>
                <Badge variant="outline" style={{ borderColor: difficultyConfig[difficulty].color, color: difficultyConfig[difficulty].color }}>
                  {difficultyConfig[difficulty].label}
                </Badge>
                <Badge variant="secondary">{grade || '未指定年级'}</Badge>
              </>
            }
            actions={
              <>
                <Button variant="ghost" asChild>
                  <Link href={isNew ? '/problems' : `/problems/${id}`}>
                    <ArrowLeft className="mr-2 h-4 w-4" />
                    返回
                  </Link>
                </Button>
                <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
                  <Save className="mr-2 h-4 w-4" />
                  保存题目
                </Button>
              </>
            }
          />

          <div className="grid gap-6 xl:grid-cols-[minmax(0,1.25fr)_380px]">
            <PagePanel>
              <div className="p-6">
                <Tabs defaultValue="question">
                  <TabsList className="mb-4">
                    <TabsTrigger value="question">题干</TabsTrigger>
                    <TabsTrigger value="answer">答案</TabsTrigger>
                    <TabsTrigger value="solution">解析</TabsTrigger>
                  </TabsList>

                  <TabsContent value="question" className="space-y-4">
                    <FormField
                      control={form.control}
                      name="latex"
                      render={({ field }) => (
                        <FormItem>
                          <FormControl>
                            <LatexCodeEditor
                              value={field.value}
                              onChange={field.onChange}
                              minHeight={300}
                              placeholder="输入题干 LaTeX..."
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <div className="rounded-2xl border border-border/70 bg-muted/30 p-6">
                      <p className="mb-3 text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground/80">实时预览</p>
                      <MathText latex={latex} className="leading-7" />
                    </div>
                  </TabsContent>

                  <TabsContent value="answer" className="space-y-4">
                    <FormField
                      control={form.control}
                      name="answerLatex"
                      render={({ field }) => (
                        <FormItem>
                          <FormControl>
                            <LatexCodeEditor
                              value={field.value ?? ''}
                              onChange={field.onChange}
                              minHeight={240}
                              placeholder="输入答案 LaTeX..."
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <div className="rounded-2xl border border-border/70 bg-muted/30 p-6">
                      {answerLatex ? (
                        <MathText latex={answerLatex} />
                      ) : (
                        <span className="text-sm text-muted-foreground">暂无答案预览</span>
                      )}
                    </div>
                  </TabsContent>

                  <TabsContent value="solution" className="space-y-4">
                    <FormField
                      control={form.control}
                      name="solutionLatex"
                      render={({ field }) => (
                        <FormItem>
                          <FormControl>
                            <LatexCodeEditor
                              value={field.value ?? ''}
                              onChange={field.onChange}
                              minHeight={260}
                              placeholder="输入解析 LaTeX..."
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                    <div className="rounded-2xl border border-border/70 bg-muted/30 p-6">
                      {solutionLatex ? (
                        <MathText latex={solutionLatex} />
                      ) : (
                        <span className="text-sm text-muted-foreground">暂无解析预览</span>
                      )}
                    </div>
                  </TabsContent>
                </Tabs>
              </div>
            </PagePanel>

            <div className="space-y-4">
              <PagePanel>
                <div className="space-y-4 p-5">
                  <h2 className="text-base font-semibold">基础信息</h2>
                  <div className="grid gap-4 sm:grid-cols-2">
                    <FormField
                      control={form.control}
                      name="type"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>题型</FormLabel>
                          <Select value={field.value} onValueChange={field.onChange}>
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              {Object.entries(problemTypeConfig).map(([value, label]) => (
                                <SelectItem key={value} value={value}>
                                  {label}
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
                      name="difficulty"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>难度</FormLabel>
                          <Select value={field.value} onValueChange={field.onChange}>
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              {(Object.keys(difficultyConfig) as Difficulty[]).map((value) => (
                                <SelectItem key={value} value={value}>
                                  {difficultyConfig[value].label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>

                  <FormField
                    control={form.control}
                    name="subjectiveScore"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>主观难度 {subjectiveScore}</FormLabel>
                        <FormControl>
                          <Slider value={[field.value ?? 5]} onValueChange={([val]) => field.onChange(val)} min={1} max={10} step={1} />
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
                        <Select value={field.value || UNSPECIFIED_GRADE} onValueChange={(value) => field.onChange(value === UNSPECIFIED_GRADE ? '' : value)}>
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value={UNSPECIFIED_GRADE}>未指定</SelectItem>
                            {gradeOptions.map((item) => (
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
                          <Input value={field.value ?? ''} onChange={field.onChange} placeholder="如：2023 北京高考" />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="notes"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>备注</FormLabel>
                        <FormControl>
                          <Textarea value={field.value ?? ''} onChange={field.onChange} rows={4} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
              </PagePanel>

              <PagePanel>
                <div className="space-y-4 p-5">
                  <h2 className="text-base font-semibold">标签</h2>
                  <div className="flex flex-wrap gap-2">
                    {(tagsQuery.data ?? []).map((tag) => {
                      const active = tagIds.includes(tag.id);
                      return (
                        <button
                          key={tag.id}
                          type="button"
                          onClick={() => {
                            const next = active ? tagIds.filter((item) => item !== tag.id) : [...tagIds, tag.id];
                            form.setValue('tagIds', next);
                          }}
                        >
                          <Badge
                            variant={active ? 'default' : 'outline'}
                            style={!active ? { borderColor: tag.color, color: tag.color } : undefined}
                          >
                            {tag.name}
                          </Badge>
                        </button>
                      );
                    })}
                  </div>
                </div>
              </PagePanel>

              <PagePanel>
                <div className="space-y-4 p-5">
                  <div>
                    <h2 className="text-base font-semibold">关联图像</h2>
                    <p className="mt-1 text-sm text-muted-foreground">勾选后会把图像与当前题目绑定。</p>
                  </div>
                  <div className="grid gap-3">
                    {(imagesQuery.data?.items ?? []).map((image) => (
                      <label key={image.id} className="flex items-center gap-3 rounded-xl border border-border/70 p-3">
                        <Checkbox
                          checked={imageIds.includes(image.id)}
                          onCheckedChange={() => {
                            const next = imageIds.includes(image.id)
                              ? imageIds.filter((item) => item !== image.id)
                              : [...imageIds, image.id];
                            form.setValue('imageIds', next);
                          }}
                        />
                        <img src={image.thumbnailUrl} alt={image.filename} className="h-12 w-16 rounded-lg object-cover" />
                        <div className="min-w-0">
                          <p className="truncate text-sm font-medium">{image.filename}</p>
                          <p className="text-xs text-muted-foreground">{image.description ?? '未填写描述'}</p>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>
              </PagePanel>
            </div>
          </div>
        </PageShell>
      </form>
    </Form>
  );
}
