package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
)

// errorMessages maps error codes to user-friendly messages.
var errorMessages = map[string]string{
	// Database/repository errors
	"db_unavailable":              "数据库连接失败，请稍后重试",
	"list_problems_failed":       "获取题目列表失败，请稍后重试",
	"get_problem_failed":         "获取题目详情失败，请稍后重试",
	"create_problem_failed":      "创建题目失败，请稍后重试",
	"update_problem_failed":      "更新题目失败，请稍后重试",
	"delete_problem_failed":      "删除题目失败，请稍后重试",
	"restore_problem_failed":     "恢复题目失败，请稍后重试",
	"hard_delete_problem_failed": "永久删除题目失败，请稍后重试",
	"list_versions_failed":       "获取版本历史失败，请稍后重试",
	"rollback_failed":             "回滚版本失败，请稍后重试",
	"import_failed":               "导入题目失败，请稍后重试",
	"parse_multipart_failed":     "解析上传文件失败，请重试",
	"invalid_defaults_json":      "默认值参数格式错误",
	"no_files":                   "请至少上传一个 .tex 文件",
	"invalid_file_type":          "仅支持 .tex 文件",
	"file_read_failed":           "读取文件失败，请重试",
	"batch_tag_failed":           "批量标记题目失败，请稍后重试",
	"batch_delete_failed":        "批量删除题目失败，请稍后重试",

	// Image errors
	"list_images_failed":              "获取图片列表失败，请稍后重试",
	"get_image_failed":                "获取图片详情失败，请稍后重试",
	"invalid_multipart":                "无效的上传请求，请重试",
	"file_required":                    "请选择要上传的文件",
	"read_file_failed":                 "读取文件失败，请重试",
	"upload_image_failed":              "上传图片失败，请稍后重试",
	"update_image_failed":              "更新图片失败，请稍后重试",
	"delete_image_failed":              "删除图片失败，请稍后重试",
	"restore_image_failed":             "恢复图片失败，请稍后重试",
	"hard_delete_image_failed":         "永久删除图片失败，请稍后重试",
	"edit_image_failed":                "编辑图片失败，请稍后重试",
	"load_image_file_failed":           "加载图片文件失败，请稍后重试",
	"load_image_thumbnail_failed":      "加载图片缩略图失败，请稍后重试",
	"batch_delete_images_failed":       "批量删除图片失败，请稍后重试",

	// Tag errors
	"list_tags_failed":      "获取标签列表失败，请稍后重试",
	"create_tag_failed":     "创建标签失败，请稍后重试",
	"update_tag_failed":     "更新标签失败，请稍后重试",
	"delete_tag_failed":     "删除标签失败，请稍后重试",
	"merge_tag_failed":      "合并标签失败，请稍后重试",

	// Paper errors
	"list_papers_failed":           "获取试卷列表失败，请稍后重试",
	"get_paper_failed":             "获取试卷详情失败，请稍后重试",
	"create_paper_failed":          "创建试卷失败，请稍后重试",
	"update_paper_failed":          "更新试卷失败，请稍后重试",
	"load_paper_failed":            "加载试卷失败，请稍后重试",
	"update_paper_items_failed":    "更新试卷题目失败，请稍后重试",
	"duplicate_paper_failed":      "复制试卷失败，请稍后重试",

	// Search errors
	"invalid_conditions":            "搜索条件无效，请检查后重试",
	"search_failed":                  "搜索失败，请稍后重试",
	"list_search_history_failed":    "获取搜索历史失败，请稍后重试",
	"delete_search_history_failed": "删除搜索历史失败，请稍后重试",
	"list_saved_searches_failed":    "获取收藏搜索失败，请稍后重试",
	"create_saved_search_failed":   "创建收藏搜索失败，请稍后重试",
	"delete_saved_search_failed":   "删除收藏搜索失败，请稍后重试",

	// Export errors
	"create_export_failed":       "创建导出任务失败，请稍后重试",
	"list_exports_failed":        "获取导出列表失败，请稍后重试",
	"get_export_failed":          "获取导出详情失败，请稍后重试",
	"load_export_failed":         "加载导出任务失败，请稍后重试",
	"cancel_export_failed":       "取消导出任务失败，请稍后重试",
	"delete_export_failed":       "删除导出任务失败，请稍后重试",
	"download_export_failed":     "下载导出文件失败，请稍后重试",

	// Meta errors
	"meta_grades_failed":       "获取年级信息失败，请稍后重试",
	"meta_stats_failed":        "获取统计数据失败，请稍后重试",
	"recent_problems_failed":   "获取最近题目失败，请稍后重试",
	"recent_exports_failed":    "获取最近导出失败，请稍后重试",

	// Settings errors
	"get_settings_failed":         "获取设置失败，请稍后重试",
	"update_settings_failed":       "更新设置失败，请稍后重试",
	"reset_demo_data_failed":      "重置演示数据失败，请稍后重试",
	"export_all_failed":           "导出全部数据失败，请稍后重试",
	"import_all_failed":           "导入全部数据失败，请稍后重试",
	"sweep_orphans_failed":        "清理孤立文件失败，请稍后重试",
	"load_demo_data_failed":       "加载演示数据失败，请稍后重试",
	"clear_demo_data_failed":      "清除演示数据失败，请稍后重试",
	"get_demo_data_status_failed": "获取演示数据状态失败，请稍后重试",

	// Validation errors
	"validation_failed": "数据验证失败，请检查输入",

	// Generic errors
	"internal_error": "服务器内部错误，请稍后重试",
	"not_found":      "请求的资源不存在",
}

// defaultErrorMessage is returned when an error code is not in the map.
const defaultErrorMessage = "操作失败，请稍后重试"

var logger zerolog.Logger

type envelope struct {
	Data  any       `json:"data"`
	Error *apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Data: payload, Error: nil})
}

func respondError(w http.ResponseWriter, status int, code string, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if logger.GetLevel() != zerolog.Disabled {
		logger.Error().Err(err).Str("code", code).Msg("request error")
	}

	userMessage := defaultErrorMessage
	if msg, ok := errorMessages[code]; ok {
		userMessage = msg
	}
	if err != nil {
		userMessage = userMessage + " [" + err.Error() + "]"
	}

	_ = json.NewEncoder(w).Encode(envelope{
		Data: nil,
		Error: &apiError{
			Code:    code,
			Message: userMessage,
		},
	})
}

func SetLogger(l zerolog.Logger) {
	logger = l
}
