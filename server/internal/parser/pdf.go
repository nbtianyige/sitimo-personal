package parser

// ConvertPDF 将 PDF 字节数据转换为 .tex 或 .md 文本。
// 它委托给本地工具链（convertWithLocalTools），保持向后兼容。
// 如需使用 Cloud API，请直接使用 NewPDFConverter 和 PDFConverter 接口。
func ConvertPDF(data []byte, originalFilename string) (string, string, error) {
	return convertWithLocalTools(data, originalFilename)
}
