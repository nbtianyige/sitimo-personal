package parser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// convertWithLocalTools 尝试用本地命令行工具把 PDF 文件转换为 LaTeX/Markdown 文本。
// 按下列优先级选择可用的工具：
//   1. magic-pdf (本地 MinerU CLI) — 对中文试卷效果最好
//   2. pandoc — 通用转换，保持格式
//   3. pdftotext — 纯文本提取，仅适用于有文本层的 PDF
func convertWithLocalTools(data []byte, originalFilename string) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", "sitimo-pdf-*")
	if err != nil {
		return "", "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpPDF := filepath.Join(tmpDir, "input.pdf")
	if err := os.WriteFile(tmpPDF, data, 0o644); err != nil {
		return "", "", fmt.Errorf("write temp pdf: %w", err)
	}

	// Strategy 1: 本地 MinerU (magic-pdf) — 对扫描/图版 PDF 效果最好
	if hasCommand("magic-pdf") {
		return convertWithLocalMinerU(tmpPDF, tmpDir)
	}

	// Strategy 2: pandoc — 通用 PDF 转 LaTeX
	if hasCommand("pandoc") {
		return convertWithPandoc(tmpPDF)
	}

	// Strategy 3: pdftotext — 纯文本提取（仅对文字层 PDF 有效）
	if hasCommand("pdftotext") {
		return convertWithPdftotext(tmpPDF)
	}

	return "", "", fmt.Errorf("本机未安装 PDF 转换工具（magic-pdf / pandoc / pdftotext）。当前环境为 %s，请检查配置：%s",
		runtime.GOOS, getToolInstallHelp())
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func convertWithLocalMinerU(pdfPath, outDir string) (string, string, error) {
	cmd := exec.Command("magic-pdf", "-p", pdfPath, "-o", outDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("magic-pdf 转换失败: %w\noutput: %s", err, string(out))
	}

	var candidates []string
	_ = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".tex" || ext == ".md" {
			candidates = append(candidates, path)
		}
		return nil
	})

	if len(candidates) == 0 {
		return "", "", fmt.Errorf("magic-pdf 完成但输出目录中未找到 .tex/.md 文件: %s", outDir)
	}

	content, err := os.ReadFile(candidates[0])
	if err != nil {
		return "", "", fmt.Errorf("读取转换后文件: %w", err)
	}

	ext := filepath.Ext(candidates[0])
	return string(content), ext, nil
}

func convertWithPandoc(pdfPath string) (string, string, error) {
	cmd := exec.Command("pandoc", "-f", "pdf", "-t", "latex", pdfPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("pandoc 转换失败: %w\noutput: %s", err, string(out))
	}
	return string(out), ".tex", nil
}

func convertWithPdftotext(pdfPath string) (string, string, error) {
	cmd := exec.Command("pdftotext", "-layout", pdfPath, "-")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("pdftotext 转换失败: %w\noutput: %s", err, string(out))
	}
	return string(out), ".md", nil
}

func getToolInstallHelp() string {
	switch {
	case hasCommand("magic-pdf"):
		return "magic-pdf 可用但当前 PDF 可能是扫描版，逐行提取无文字层内容极其有限"
	default:
		return "推荐安装 MinerU CLI (magic-pdf) 以支持扫描版 PDF，或在 .env 中配置 MINERU_API_KEY 使用 Cloud 版"
	}
}
