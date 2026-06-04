import type { ReactNode } from 'react';
import { cn } from '@/lib/utils';

export function PageShell({
  children,
  className,
  wide = false,
}: {
  children: ReactNode;
  className?: string;
  wide?: boolean;
}) {
  return (
    <div
      className={cn(
        'mx-auto w-full space-y-5 px-5 py-5 md:space-y-6 md:px-6 md:py-6 xl:px-8',
        wide ? 'max-w-[1440px]' : 'max-w-7xl',
        className
      )}
    >
      {children}
    </div>
  );
}

export function PageHeader({
  title,
  description,
  actions,
  eyebrow,
  badges,
  children,
  className,
  contentClassName,
}: {
  title: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  eyebrow?: ReactNode;
  badges?: ReactNode;
  children?: ReactNode;
  className?: string;
  contentClassName?: string;
}) {
  return (
    <section
      className={cn(
        'rounded-3xl border border-border/70 bg-card/85 px-5 py-5 shadow-sm backdrop-blur md:px-6 md:py-6 xl:px-7',
        className
      )}
    >
      <div className="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
        <div className={cn('min-w-0 space-y-3', contentClassName)}>
          {eyebrow ? <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground/80">{eyebrow}</p> : null}
          <div className="space-y-2">
            <h1 className="text-3xl font-semibold tracking-tight text-foreground">{title}</h1>
            {description ? <div className="max-w-3xl text-sm leading-6 text-muted-foreground">{description}</div> : null}
          </div>
          {badges ? <div className="flex flex-wrap items-center gap-2">{badges}</div> : null}
          {children}
        </div>

        {actions ? <div className="flex flex-wrap items-center gap-2 xl:justify-end">{actions}</div> : null}
      </div>
    </section>
  );
}

export function PageToolbar({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        'rounded-2xl border border-border/70 bg-card/75 p-4 shadow-sm backdrop-blur md:p-5',
        className
      )}
    >
      {children}
    </div>
  );
}

export function PagePanel({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <section
      className={cn(
        'rounded-3xl border border-border/70 bg-card/80 shadow-sm backdrop-blur',
        className
      )}
    >
      {children}
    </section>
  );
}
