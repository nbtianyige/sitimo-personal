package parser

import (
	"fmt"
	"os/exec"
)

// PDFConverter 是一个通用 PDF → 文本/LaTeX 转换器接口。
// 它可以由本地命令行工具、Cloud API（如 MinerU）或 LLM Vision 等实现。
type PDFConverter interface {
	// Convert 接收 PDF 二进制数据和原始文件名，返回 (转换后文本, 目标扩展名, 错误)。
	// 扩展名通常为 ".md" 或 ".tex"，供下游 parser 根据扩展名决定后续处理。
	Convert(pdfData []byte, filename string) (string, string, error)
	// Name 返回该转换器的可识别名称，用于日志和调试。
	Name() string
}

// NewPDFConverter 根据配置创建一个合适的 PDFConverter 实例。
// 支持的 converterType:
//   - "mineru-cloud": 使用 MinerU Cloud API（需要 apiKey）。
//   - "local": 使用本地命令行工具链回退（magic-pdf → pandoc → pdftotext）。
//   - "disabled": 禁用 PDF 转换，直接返回不支持的错误。
func NewPDFConverter(converterType, apiKey, apiBase string) PDFConverter {
	switch converterType {
	case "mineru-cloud":
		if apiKey == "" {
			// 缺少 API Key 时优雅降级为 local 工具
			return NewLocalToolsConverter()
		}
		return NewMinerUCloudConverter(apiKey, apiBase)
	case "disabled":
		return &disabledConverter{}
	default:
		// 默认 "local"
		return NewLocalToolsConverter()
	}
}

// LocalToolsConverter 使用本地命令行工具把 PDF 转成文本/LaTeX。
// 按优先级尝试 magic-pdf → pandoc → pdftotext。
type LocalToolsConverter struct{}

func NewLocalToolsConverter() *LocalToolsConverter {
	return &LocalToolsConverter{}
}

func (*LocalToolsConverter) Name() string           { return "local-tools" }
func (*LocalToolsConverter) Convert(data []byte, filename string) (string, string, error) {
	// 委托给现有包函数，保持行为不变
	return convertWithLocalTools(data, filename)
}

// HasLocalPDFTool 检查本地是否安装了至少一个可用的 PDF 转换工具。
func HasLocalPDFTool() bool {
	for _, cmd := range []string{"magic-pdf", "pandoc", "pdftotext"} {
		if _, err := exec.LookPath(cmd); err == nil {
			return true
		}
	}
	return false
}

// disabledConverter 用于显式禁用 PDF 转换的场景。
type disabledConverter struct{}

func (d *disabledConverter) Name() string { return "disabled" }
func (d *disabledConverter) Convert(_ []byte, _ string) (string, string, error) {
	return "", "", fmt.Errorf("PDF 转换已禁用。请先将 PDF 转为 .tex 或 .md 后重试")
}
