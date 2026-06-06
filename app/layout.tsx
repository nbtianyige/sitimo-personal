import type { Metadata, Viewport } from 'next';
import { Analytics } from '@vercel/analytics/next';
import { Toaster } from 'sonner';
import { ThemeProvider } from '@/components/theme-provider';
import { MathProvider } from '@/components/math-provider';
import { AppLayout } from '@/components/app-layout';
import { QueryProvider } from '@/components/query-provider';
import { AppRuntime } from '@/components/app-runtime';
import './globals.css';

export const metadata: Metadata = {
  title: 'Sitimo - 数学题库管理系统',
  description: '面向中文教师的数学题库管理系统，支持 LaTeX 公式编辑、试卷组排与导出',
  generator: 'v0.app',
  icons: {
    icon: [
      {
        url: '/icon-light-32x32.png',
        media: '(prefers-color-scheme: light)',
      },
      {
        url: '/icon-dark-32x32.png',
        media: '(prefers-color-scheme: dark)',
      },
      {
        url: '/icon.svg',
        type: 'image/svg+xml',
      },
    ],
    apple: '/apple-icon.png',
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#0F766E' },
    { media: '(prefers-color-scheme: dark)', color: '#14B8A6' },
  ],
  width: 'device-width',
  initialScale: 1,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="zh-CN"
      suppressHydrationWarning
      className="bg-background"
    >
      <body className="font-sans antialiased">
        <QueryProvider>
          <ThemeProvider
            attribute="class"
            defaultTheme="system"
            enableSystem
            disableTransitionOnChange
          >
            <MathProvider>
              <AppRuntime />
              <AppLayout>{children}</AppLayout>
            </MathProvider>
            <Toaster position="top-right" richColors closeButton />
          </ThemeProvider>
        </QueryProvider>
        {process.env.NODE_ENV === 'production' && <Analytics />}
      </body>
    </html>
  );
}
