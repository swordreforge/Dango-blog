package validation

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
)

// ValidatePDF 验证PDF文件，防止宏和文件外带
func ValidatePDF(content []byte) error {
	// 检查文件大小限制（100MB）
	if err := CheckFileSize(content, 100*1024*1024); err != nil {
		return err
	}

	// 检查PDF文件头
	if len(content) < 5 || string(content[0:5]) != "%PDF-" {
		return errors.New("无效的PDF文件")
	}

	// 检查危险内容
	contentStr := string(content)
	dangerousPatterns := []string{
		"/JavaScript",
		"/JS",
		"/AA", // Auto Action
		"/OpenAction",
		"/Launch",
		"/SubmitForm",
		"/URI",
		"/GoTo",
		"/GoToR",
		"/EmbeddedFile",
	}

	if err := CheckDangerousPatterns(contentStr, dangerousPatterns); err != nil {
		return err
	}

	return nil
}

// ValidateDOCX 验证DOCX文件
func ValidateDOCX(content []byte) error {
	// 检查文件大小限制（50MB）
	if err := CheckFileSize(content, 50*1024*1024); err != nil {
		return err
	}

	// 检查ZIP文件头（DOCX本质是ZIP文件）
	if len(content) < 4 || content[0] != 0x50 || content[1] != 0x4B ||
		content[2] != 0x03 || content[3] != 0x04 {
		return errors.New("无效的DOCX文件")
	}

	// 解压并检查内容
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return errors.New("解析DOCX文件失败")
	}

	// 检查是否包含宏（.docx不应该包含宏）
	for _, file := range reader.File {
		if strings.Contains(file.Name, "macros") ||
			strings.Contains(file.Name, "vbaProject") ||
			strings.Contains(file.Name, ".bin") {
			return errors.New("DOCX文件包含宏")
		}
	}

	return nil
}