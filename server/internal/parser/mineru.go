package parser

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MinerUCloudConverter 调用 MinerU Cloud API (https://mineru.net) 把扫描版 PDF
// 转成结构化的 Markdown / LaTeX 文本。它适合处理没有文本层的图片/扫描版试卷。
type MinerUCloudConverter struct {
	apiKey  string
	apiBase string
	client  *http.Client
}

func NewMinerUCloudConverter(apiKey, apiBase string) *MinerUCloudConverter {
	base := strings.TrimRight(apiBase, "/")
	if base == "" {
		base = "https://mineru.net/api/v4"
	}
	return &MinerUCloudConverter{
		apiKey:  apiKey,
		apiBase: base,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (m *MinerUCloudConverter) Name() string {
	return "mineru-cloud"
}

// Convert 把 PDF 二进制数据提交到 MinerU Cloud，轮询完成后返回转换后的文本。
// 返回 (content, extension, error)，扩展名通常为 ".md"。
func (m *MinerUCloudConverter) Convert(pdfData []byte, filename string) (string, string, error) {
	// 步骤 1：获取预签名上传 URL
	batchID, uploadURL, err := m.requestUploadURL(filename)
	if err != nil {
		return "", "", fmt.Errorf("mineru upload-url: %w", err)
	}

	// 步骤 2：上传 PDF 到预签名 URL
	if err := m.uploadFile(pdfData, uploadURL); err != nil {
		return "", "", fmt.Errorf("mineru upload: %w", err)
	}

	// 步骤 3：轮询转换结果
	resultURL, err := m.pollResult(batchID, 600)
	if err != nil {
		return "", "", fmt.Errorf("mineru poll: %w", err)
	}

	// 步骤 4：下载结果 ZIP，解压并读取 .md / .tex 内容
	content, ext, err := m.downloadAndExtract(resultURL, filename)
	if err != nil {
		return "", "", fmt.Errorf("mineru download: %w", err)
	}

	return content, ext, nil
}

// 探测结果：返回 (batch_id, 上传URL列表)。
// 上传 URL 为预签名 URL（通常为 S3/OSS），PUT 时不要添加任何额外 Header。
func (m *MinerUCloudConverter) requestUploadURL(filename string) (string, string, error) {
	payload := map[string]any{
		"files":           []map[string]string{{"name": filename}},
		"model_version":   "vlm",
		"language":        "ch",
		"enable_formula":  true,
		"enable_table":    true,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, m.apiBase+"/file-urls/batch", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("%d %s", resp.StatusCode, string(b))
	}

	var envelope struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			BatchID  string   `json:"batch_id"`
			FileURLs []string `json:"file_urls"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b, &envelope); err != nil {
		return "", "", err
	}
	if envelope.Code != 0 {
		return "", "", fmt.Errorf("code=%d: %s", envelope.Code, envelope.Msg)
	}
	if len(envelope.Data.FileURLs) == 0 {
		return "", "", fmt.Errorf("no file_urls returned")
	}
	return envelope.Data.BatchID, envelope.Data.FileURLs[0], nil
}

// uploadFile 将 PDF 数据 PUT 到预签名 URL。
// 注意：预签名 URL 通常来自 S3/OSS，不能加额外 Header，否则会破坏签名。
func (m *MinerUCloudConverter) uploadFile(data []byte, uploadURL string) error {
	req, err := http.NewRequest(http.MethodPut, uploadURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
func (m *MinerUCloudConverter) pollResult(batchID string, maxWaitSec int) (string, error) {
	start := time.Now()
	interval := 2 * time.Second
	maxInterval := 30 * time.Second
	url := m.apiBase + "/extract-results/batch/" + batchID

	for time.Since(start) < time.Duration(maxWaitSec)*time.Second {
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Bearer "+m.apiKey)

		resp, err := m.client.Do(req)
		if err != nil {
			time.Sleep(interval)
			interval = min(interval*2, maxInterval)
			continue
		}

		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var envelope struct {
			Code int `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				ExtractResult []struct {
					State      string `json:"state"`
					FullZipURL string `json:"full_zip_url"`
					ErrMsg     string `json:"err_msg"`
				} `json:"extract_result"`
			} `json:"data"`
		}
		if err := json.Unmarshal(b, &envelope); err != nil {
			time.Sleep(interval)
			interval = min(interval*2, maxInterval)
			continue
		}
		if envelope.Code != 0 {
			return "", fmt.Errorf("code=%d: %s", envelope.Code, envelope.Msg)
		}

		if len(envelope.Data.ExtractResult) > 0 {
			res := envelope.Data.ExtractResult[0]
			switch res.State {
			case "done":
				return res.FullZipURL, nil
			case "failed":
				return "", fmt.Errorf("task failed: %s", res.ErrMsg)
			}
		}

		time.Sleep(interval)
		interval = min(interval*2, maxInterval)
	}
	return "", fmt.Errorf("polling timed out after %ds", maxWaitSec)
}

// downloadAndExtract 下载 ZIP 到临时目录，解压后找到 .md 或 .tex 文件并读取内容。
func (m *MinerUCloudConverter) downloadAndExtract(zipURL, originalFilename string) (string, string, error) {
	// 下载 ZIP
	resp, err := m.client.Get(zipURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("download %d", resp.StatusCode)
	}

	// 先读到内存（典型结果 ZIP 通常 < 10MB）
	zipData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", "", err
	}

	// 优先找主输出文件：full.md（默认主输出） > 任何 .md > 任何 .tex
	var mdFile, texFile *zip.File
	for _, f := range zr.File {
		name := strings.ToLower(f.Name)
		switch {
		case name == "full.md":
			mdFile = f
		case mdFile == nil && strings.HasSuffix(name, ".md"):
			mdFile = f
		case strings.HasSuffix(name, ".tex"):
			if texFile == nil { texFile = f }
		}
	}

	target := mdFile
	ext := ".md"
	if target == nil {
		target = texFile
		ext = ".tex"
	}
	if target == nil {
		return "", "", fmt.Errorf("ZIP 中未找到 .md 或 .tex 文件")
	}

	rc, err := target.Open()
	if err != nil {
		return "", "", err
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return "", "", err
	}
	return string(content), ext, nil
}
