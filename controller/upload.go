package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"myblog-gogogo/service"
)

// UploadHandler 上传处理器
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求方法不允许",
			"error":   "仅支持 POST 请求",
			"code":    "METHOD_NOT_ALLOWED",
		})
		return
	}

	// 解析表单数据，限制最大文件大小为 100MB
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "解析表单失败",
			"error":   err.Error(),
			"code":    "PARSE_FORM_FAILED",
		})
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "获取文件失败，请确保选择了文件",
			"error":   err.Error(),
			"code":    "NO_FILE_PROVIDED",
		})
		return
	}
	defer file.Close()

	// 检查文件大小（前端已限制10MB，后端再次验证）
	if header.Size > 10*1024*1024 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "文件过大",
			"error":   fmt.Sprintf("文件大小 %.2fMB 超过 10MB 限制", float64(header.Size)/(1024*1024)),
			"code":    "FILE_TOO_LARGE",
		})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(header.Filename[strings.LastIndex(header.Filename, "."):])
	supportedTypes := map[string]bool{
		".md": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".webp": true, ".bmp": true, ".svg": true,
	}

	if !supportedTypes[ext] {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "不支持的文件类型",
			"error":   fmt.Sprintf("仅支持 .md, .jpg, .jpeg, .png, .gif, .webp, .bmp, .svg 文件，当前文件类型: %s", ext),
			"code":    "UNSUPPORTED_FILE_TYPE",
		})
		return
	}

	// 获取日期参数（可选）
	year := r.FormValue("year")
	month := r.FormValue("month")
	day := r.FormValue("day")

	// 创建上传服务
	uploadService := service.NewUploadService()

	// 处理上传
	result, err := uploadService.HandleUpload(file, header, year, month, day)
	if err != nil {
		// 根据错误信息判断具体错误类型
		errMsg := err.Error()
		if strings.Contains(errMsg, "不支持的文件类型") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "不支持的文件类型",
				"error":   errMsg,
				"code":    "UNSUPPORTED_FILE_TYPE",
			})
		} else if strings.Contains(errMsg, "创建目录失败") || strings.Contains(errMsg, "写入文件失败") {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "文件保存失败",
				"error":   errMsg,
				"code":    "FILE_SAVE_FAILED",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "上传失败",
				"error":   errMsg,
				"code":    "UPLOAD_FAILED",
			})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "上传成功",
		"data":    result,
		"code":    "UPLOAD_SUCCESS",
	})
}
