package validation

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
)

// ValidateSVG 验证SVG文件，防止XXE攻击
func ValidateSVG(content []byte) error {
	// 检查文件大小限制（10MB）
	if err := CheckFileSize(content, 10*1024*1024); err != nil {
		return err
	}

	// 检查是否包含外部实体引用
	contentStr := string(content)
	dangerousPatterns := []string{
		"<!ENTITY",
		"<!DOCTYPE",
		"SYSTEM",
		"xlink:href",
		"<script",
		"javascript:",
		"data:",
		"vbscript:",
	}

	if err := CheckDangerousPatterns(contentStr, dangerousPatterns); err != nil {
		return err
	}

	// 尝试解析XML，但不处理外部实体
	decoder := xml.NewDecoder(bytes.NewReader(content))
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity

	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New("SVG文件格式错误")
		}
	}

	return nil
}
